package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"event-service/internal/core/domain"
	service "event-service/internal/layers/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	pgxpool "github.com/jackc/pgx/v5/pgxpool"
)

// maxTxAttempts — сколько раз WithTx пробует выполнить транзакцию, если
// Postgres откатил её из-за конфликта сериализации или дедлока.
const maxTxAttempts = 5

type postgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) service.EventRepository {
	return &postgresRepository{
		pool: pool,
	}
}

// querier — общее у pgxpool.Pool и pgx.Tx, чтобы одни и те же методы
// репозитория работали и сами по себе, и внутри WithTx.
type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type txKey struct{}

// db отдаёт транзакцию, если метод вызван изнутри WithTx, иначе — пул. Всё
// проходит через classifyingQuerier, поэтому недоступность базы размечается
// в одном месте, а не на каждом возврате ошибки.
func (r *postgresRepository) db(ctx context.Context) querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return classifyingQuerier{tx}
	}
	return classifyingQuerier{r.pool}
}

// classifyingQuerier отделяет "база недоступна" от остальных ошибок: по
// первому вызывающему есть что решать (повторить запрос), второе — наш
// внутренний сбой.
type classifyingQuerier struct {
	q querier
}

func (c classifyingQuerier) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	tag, err := c.q.Exec(ctx, sql, args...)
	return tag, classifyDBError(err)
}

func (c classifyingQuerier) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	rows, err := c.q.Query(ctx, sql, args...)
	return rows, classifyDBError(err)
}

func (c classifyingQuerier) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return classifyingRow{c.q.QueryRow(ctx, sql, args...)}
}

type classifyingRow struct {
	row pgx.Row
}

func (r classifyingRow) Scan(dest ...any) error {
	return classifyDBError(r.row.Scan(dest...))
}

// classifyDBError помечает недоступность базы. Всё остальное возвращается
// как есть — в том числе pgx.ErrNoRows, который вызывающие сравнивают
// напрямую, и ошибки сериализации, по которым WithTx решает, повторять ли
// транзакцию.
func classifyDBError(err error) error {
	if err == nil {
		return nil
	}

	var connErr *pgconn.ConnectError
	if errors.As(err, &connErr) {
		return domain.Wrap(domain.KindUnavailable, err, "database is unavailable")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// класс 08 — ошибки соединения, 57P01/57P03 — сервер выключается или
		// ещё не готов принимать запросы
		if strings.HasPrefix(pgErr.Code, "08") || pgErr.Code == "57P01" || pgErr.Code == "57P03" {
			return domain.Wrap(domain.KindUnavailable, err, "database is unavailable")
		}
	}

	return err
}

// WithTx выполняет fn в одной транзакции с уровнем изоляции SERIALIZABLE:
// все методы репозитория, вызванные с переданным в fn ctx, идут через неё.
// Если Postgres откатывает транзакцию из-за конфликта сериализации (40001)
// или дедлока (40P01), fn выполняется заново — поэтому fn должна читать всё
// нужное внутри себя и не иметь побочных эффектов вне БД (Kafka — только
// после WithTx). Вложенный вызов просто продолжает внешнюю транзакцию.
func (r *postgresRepository) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return fn(ctx)
	}

	var err error
	for attempt := 1; attempt <= maxTxAttempts; attempt++ {
		err = pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
			return fn(context.WithValue(ctx, txKey{}, tx))
		})
		if !isRetryableTxError(err) {
			return err
		}

		select {
		case <-time.After(time.Duration(attempt) * 20 * time.Millisecond):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Повторы кончились, а конфликт остался: вызывающий может повторить
	// запрос целиком — это ровно то, что означает gRPC-код Aborted.
	return domain.Wrap(domain.KindAborted, err, "transaction conflict after %d attempts, please retry", maxTxAttempts)
}

// isUniqueViolation — гонка на вставке, которую сервис уже проверял чтением
// (игрок уже в ивенте/команде): отдаём её как понятный конфликт, а не как
// внутреннюю ошибку.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isRetryableTxError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "40001" || pgErr.Code == "40P01"
}

// eventColumns/scanEvent — общий разбор строки events. time_finish в БД NULL,
// пока ивент не завершён, поэтому читается через указатель: раньше туда
// писался нулевой time.Time (0001-01-01), и "не завершён" выглядело как
// настоящая дата.
func scanEvent(row pgx.Row) (domain.Event, error) {
	var event domain.Event
	var timeFinish *time.Time

	err := row.Scan(
		&event.EventID,
		&event.Name,
		&event.UserCreateID,
		&event.EnemySideLeader,
		&event.UserCount,
		&event.TimeStart,
		&timeFinish,
		&event.CreateTime,
		&event.WinnerSide,
		&event.TargetGameCount,
		&event.GameCount,
		&event.Status,
		&event.RentServerSent,
	)
	if err != nil {
		return domain.Event{}, err
	}

	if timeFinish != nil {
		event.TimeFinish = *timeFinish
	}

	return event, nil
}

func (r *postgresRepository) CreateEvent(ctx context.Context, event domain.Event) error {
	query := `
	INSERT INTO events (event_id, name, user_create_id, enemy_side_leader_id, time_start, time_finish, create_time, user_count, target_game_count, game_count, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	// time_finish у нового ивента — настоящий NULL, а не нулевой time.Time.
	var timeFinish *time.Time
	if !event.TimeFinish.IsZero() {
		timeFinish = &event.TimeFinish
	}

	_, err := r.db(ctx).Exec(ctx, query, event.EventID, event.Name, event.UserCreateID, event.EnemySideLeader, event.TimeStart, timeFinish, event.CreateTime, event.UserCount, event.TargetGameCount, event.GameCount, event.Status)
	if err != nil {
		return err
	}

	return nil
}

func (r *postgresRepository) GetEventsByCreatorId(ctx context.Context, userCreateID string) ([]*domain.Event, error) {
	query := `
		SELECT event_id, name, user_create_id, enemy_side_leader_id, user_count, time_start, time_finish, create_time, winner_side, target_game_count, game_count, status, rent_server_sent
		FROM events
		WHERE user_create_id = $1
	`

	rows, err := r.db(ctx).Query(ctx, query, userCreateID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]*domain.Event, 0)
	for rows.Next() {
		evn, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, &evn)
	}

	return events, rows.Err()
}

func (r *postgresRepository) GetLastEventByCreatorId(ctx context.Context, userCreateID string) (*domain.Event, error) {
	query := `
		SELECT event_id, name, user_create_id, enemy_side_leader_id, user_count, time_start, time_finish, create_time, winner_side, target_game_count, game_count, status, rent_server_sent
		FROM events
		WHERE user_create_id = $1
		ORDER BY create_time DESC
		LIMIT 1
	`

	evn, err := scanEvent(r.db(ctx).QueryRow(ctx, query, userCreateID))
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &evn, nil
}

func (r *postgresRepository) GetEventByID(ctx context.Context, eventID string) (domain.Event, error) {
	return r.getEventByID(ctx, eventID, "")
}

// GetEventByIDForUpdate — то же, что GetEventByID, но с SELECT ... FOR UPDATE:
// строка ивента остаётся заблокированной до конца транзакции WithTx.
// Все пишущие транзакции по ивенту начинают с неё, поэтому между собой они
// выполняются строго по очереди (в том числе из разных инстансов сервиса).
func (r *postgresRepository) GetEventByIDForUpdate(ctx context.Context, eventID string) (domain.Event, error) {
	return r.getEventByID(ctx, eventID, "FOR UPDATE")
}

func (r *postgresRepository) getEventByID(ctx context.Context, eventID, lockClause string) (domain.Event, error) {
	query := `
		SELECT event_id, name, user_create_id, enemy_side_leader_id, user_count, time_start, time_finish, create_time, winner_side, target_game_count, game_count, status, rent_server_sent
		FROM events
		WHERE event_id = $1
	` + lockClause
	event, err := scanEvent(r.db(ctx).QueryRow(ctx, query, eventID))
	if err == pgx.ErrNoRows {
		return domain.Event{}, nil
	}

	if err != nil {
		return domain.Event{}, err
	}

	return event, nil
}

func (r *postgresRepository) GetEventMembersList(ctx context.Context, eventID string) ([]domain.User, error) {
	query := `
		SELECT user_event_id, user_id, event_id, clan_id, enemy, join_time
		FROM users
		WHERE event_id = $1
	`

	rows, err := r.db(ctx).Query(ctx, query, eventID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]domain.User, 0)

	for rows.Next() {
		usr, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, usr)
	}

	return users, rows.Err()
}

func (r *postgresRepository) GetEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error) {
	query := `
		SELECT event_id, name, user_create_id, enemy_side_leader_id, user_count, time_start, time_finish, create_time, winner_side, target_game_count, game_count, status, rent_server_sent
		FROM events
		WHERE name = $1
	`
	rows, err := r.db(ctx).Query(ctx, query, eventName)
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]domain.Event, 0)
	for rows.Next() {
		evn, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, evn)
	}

	return events, rows.Err()
}

// UpdateTimeEvent переносит старт — только пока ивент ещё pending (после
// контрольной проверки время уже зафиксировано таймерами аренды/старта).
func (r *postgresRepository) UpdateTimeEvent(ctx context.Context, eventID string, newTimeStart time.Time) error {
	query := `
		UPDATE events
		SET time_start = $1
		WHERE event_id = $2 AND status = 'pending'
	`

	tag, err := r.db(ctx).Exec(ctx, query, newTimeStart, eventID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return domain.FailedPrecondition("cannot change time of event %q: it is not pending", eventID)
	}
	return nil
}

func (r *postgresRepository) UpdateTimeFinishEvent(ctx context.Context, eventID string, newTimeFinish time.Time) error {
	query := `
		UPDATE events
		SET time_finish = $1
		WHERE event_id = $2
	`

	_, err := r.db(ctx).Exec(ctx, query, newTimeFinish, eventID)
	return err
}

// UpdateEventStatus переводит ивент в новый статус — но только из ожидаемого
// предыдущего (pending->confirmed или confirmed->in_progress), чтобы
// повторный или опоздавший вызов (например, после рестарта сервиса) не мог
// откатить ивент назад в более раннюю фазу жизненного цикла. Если переход
// не состоялся (ивент уже в другом статусе) — это ошибка, а не тихий no-op:
// иначе вызывающий продолжил бы работу (публикация в Kafka, старт игр) так,
// будто статус сменился.
func (r *postgresRepository) UpdateEventStatus(ctx context.Context, eventID string, status domain.EventStatus) error {
	var fromStatuses []string
	switch status {
	case domain.EventStatusConfirmed:
		fromStatuses = []string{string(domain.EventStatusPending)}
	case domain.EventStatusInProgress:
		fromStatuses = []string{string(domain.EventStatusConfirmed)}
	case domain.EventStatusDeclined:
		// декленд бывает только для ивента, который так и остался pending —
		// не набрал минимум игроков к контрольной точке.
		fromStatuses = []string{string(domain.EventStatusPending)}
	case domain.EventStatusCanceled:
		// отменить вручную (CancelEvent) можно и pending, и уже confirmed —
		// но не то, что уже in_progress/finished/declined.
		fromStatuses = []string{string(domain.EventStatusPending), string(domain.EventStatusConfirmed)}
	default:
		return fmt.Errorf("UpdateEventStatus: unsupported target status %q", status)
	}

	query := `
		UPDATE events
		SET status = $1
		WHERE event_id = $2 AND status = ANY($3)
	`

	tag, err := r.db(ctx).Exec(ctx, query, status, eventID, fromStatuses)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return domain.FailedPrecondition("cannot move event %q to status %q: current status must be one of %v", eventID, status, fromStatuses)
	}
	return nil
}

// UpdateEventWinner/UpdateEventLoser удалены: писали в event_team_winner/event_team_loser,
// которых больше нет в схеме (заменены на events.winner_side). Не были нигде вызваны
// из сервисного слоя — мёртвый код. Вернём в виде UpdateEventWinnerSide, когда будем
// проектировать games/teams-флоу.

func (r *postgresRepository) RenameEvent(ctx context.Context, eventID, oldName, newName string) error {
	query := `
		UPDATE events
		SET name = $1
		WHERE event_id = $2 AND name = $3
	`

	_, err := r.db(ctx).Exec(ctx, query, newName, eventID, oldName)
	return err
}

// DeleteEvent убран: раньше он реально удалял ивент со всеми зависимыми
// строками (team_members/team/users) — теперь и ручная отмена (CancelEvent,
// статус "canceled"), и автоматическая (не набрали минимум игроков к
// контрольной точке, статус "declined") просто переводят ивент в статус
// через UpdateEventStatus, а сама строка и вся история участников/команд
// остаются в базе.

func (r *postgresRepository) IsUserInEvent(ctx context.Context, eventID, userID string) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM users WHERE event_id = $1 AND user_id = $2)
	`

	var exists bool
	if err := r.db(ctx).QueryRow(ctx, query, eventID, userID).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *postgresRepository) JoinToEvent(ctx context.Context, userEventID, eventID, userID, clanID string, enemy bool, joinTime time.Time) error {
	query := `
		INSERT INTO users (user_event_id, user_id, event_id, clan_id, enemy, join_time)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db(ctx).Exec(ctx, query, userEventID, userID, eventID, clanID, enemy, joinTime)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.AlreadyExists("user %q has already joined event %q", userID, eventID)
		}
		return err
	}

	updateQuery := `
		UPDATE events
		SET user_count = user_count + 1
		WHERE event_id = $1
	`

	_, err = r.db(ctx).Exec(ctx, updateQuery, eventID)
	return err
}

func (r *postgresRepository) LeaveEvent(ctx context.Context, userID, eventID string) error {
	query := `
		DELETE FROM users
		WHERE user_id = $1 AND event_id = $2
	`

	tag, err := r.db(ctx).Exec(ctx, query, userID, eventID)
	if err != nil {
		return err
	}
	// user_count уменьшаем только если строка реально удалилась — иначе
	// повторный LeaveEvent того же игрока уводил бы счётчик в минус.
	if tag.RowsAffected() == 0 {
		return domain.NotFound("user %q not found in event %q", userID, eventID)
	}

	updateQuery := `
		UPDATE events
		SET user_count = user_count - 1
		WHERE event_id = $1
	`

	_, err = r.db(ctx).Exec(ctx, updateQuery, eventID)
	return err
}

// UpdateTeamMemberRole меняет роль игрока внутри одной команды. Роли уровня
// всего ивента нет (в users колонки role нет намеренно): роль
// имеет смысл только в рамках конкретной игры.
func (r *postgresRepository) UpdateTeamMemberRole(ctx context.Context, teamID, userEventID string, role domain.Role) error {
	query := `
		UPDATE team_members
		SET role = $1
		WHERE team_id = $2 AND user_event_id = $3
	`

	tag, err := r.db(ctx).Exec(ctx, query, string(role), teamID, userEventID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("user is not a member of team %q", teamID)
	}
	return nil
}

func (r *postgresRepository) GetUserByID(ctx context.Context, eventID, userID string) (domain.User, error) {
	query := `
		SELECT user_event_id, user_id, event_id, clan_id, enemy, join_time
		FROM users
		WHERE user_id = $1 AND event_id = $2
	`

	user, err := scanUser(r.db(ctx).QueryRow(ctx, query, userID, eventID))
	if err == pgx.ErrNoRows {
		return domain.User{}, domain.NotFound("user %q not found in event %q", userID, eventID)
	}

	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

// GetUserByUserEventID — тот же участник, но по внутреннему user_event_id.
// Нужен, чтобы по team.side_leader_id узнать сторону команды (users.enemy):
// войти в команду противоположной стороны нельзя.
func (r *postgresRepository) GetUserByUserEventID(ctx context.Context, userEventID string) (domain.User, error) {
	query := `
		SELECT user_event_id, user_id, event_id, clan_id, enemy, join_time
		FROM users
		WHERE user_event_id = $1
	`

	user, err := scanUser(r.db(ctx).QueryRow(ctx, query, userEventID))
	if err == pgx.ErrNoRows {
		return domain.User{}, domain.NotFound("user_event_id %q not found", userEventID)
	}

	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func scanUser(row pgx.Row) (domain.User, error) {
	var user domain.User
	err := row.Scan(
		&user.UserEventID,
		&user.UserID,
		&user.EventID,
		&user.ClanID,
		&user.Enemy,
		&user.JoinTime,
	)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

// CheckSixClanMembers считает ДРУГИХ участников этой КОМАНДЫ (team_members,
// исключая userEventID) с тем же кланом — clan_id хранится в users (по
// ивенту, не по команде), поэтому JOIN. Считается на уровне команды, а не
// всего ивента: другой микросервис перед стартом конкретной игры смотрит на
// team_members.six_clan_members у каждого участника и решает, приглашать его
// или нет. Порог: не менее 5 ДРУГИХ однокланников в команде (итого 6 вместе
// с самим игроком).
func (r *postgresRepository) CheckSixClanMembers(ctx context.Context, teamID, userEventID, clanID string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM team_members tm
		JOIN users u ON u.user_event_id = tm.user_event_id
		WHERE tm.team_id = $1
		AND u.clan_id = $2
		AND tm.user_event_id != $3
	`

	var count int
	err := r.db(ctx).QueryRow(ctx, query, teamID, clanID, userEventID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count >= 5, nil
}

func (r *postgresRepository) UpdateTeamMemberSixClanMembers(ctx context.Context, teamID, userEventID string, hasSixClanMembers bool) error {
	query := `
		UPDATE team_members
		SET six_clan_members = $1
		WHERE team_id = $2 AND user_event_id = $3
	`

	_, err := r.db(ctx).Exec(ctx, query, hasSixClanMembers, teamID, userEventID)
	return err
}

func (r *postgresRepository) UpdateSixClanMembersForClanInTeam(ctx context.Context, teamID, clanID string, hasSixClanMembers bool) error {
	query := `
		UPDATE team_members tm
		SET six_clan_members = $1
		FROM users u
		WHERE tm.user_event_id = u.user_event_id
		AND tm.team_id = $2
		AND u.clan_id = $3
	`

	_, err := r.db(ctx).Exec(ctx, query, hasSixClanMembers, teamID, clanID)
	return err
}

func (r *postgresRepository) CountClanMembersInTeam(ctx context.Context, teamID, clanID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM team_members tm
		JOIN users u ON u.user_event_id = tm.user_event_id
		WHERE tm.team_id = $1 AND u.clan_id = $2
	`

	var count int
	err := r.db(ctx).QueryRow(ctx, query, teamID, clanID).Scan(&count)
	return count, err
}

func (r *postgresRepository) GetUserIDsByEventID(ctx context.Context, eventID string) ([]string, error) {
	query := `
		SELECT user_id
		FROM users
		WHERE event_id = $1
	`

	rows, err := r.db(ctx).Query(ctx, query, eventID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userIDs := make([]string, 0)
	for rows.Next() {
		var userID string
		err := rows.Scan(&userID)
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, rows.Err()
}

func (r *postgresRepository) GetUnfinishedEventsByUserID(ctx context.Context, userCreateID string) ([]*domain.Event, error) {
	query := `
		SELECT event_id, name, user_create_id, enemy_side_leader_id, user_count, time_start, time_finish, create_time, winner_side, target_game_count, game_count, status, rent_server_sent
		FROM events
		WHERE user_create_id = $1 AND status IN ('pending', 'confirmed', 'in_progress')
	`

	rows, err := r.db(ctx).Query(ctx, query, userCreateID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]*domain.Event, 0)
	for rows.Next() {
		evn, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, &evn)
	}

	return events, rows.Err()
}

func (r *postgresRepository) GetUnfinishedEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error) {
	query := `
		SELECT event_id, name, user_create_id, enemy_side_leader_id, user_count, time_start, time_finish, create_time, winner_side, target_game_count, game_count, status, rent_server_sent
		FROM events
		WHERE name = $1 AND status IN ('pending', 'confirmed', 'in_progress')
	`
	rows, err := r.db(ctx).Query(ctx, query, eventName)
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]domain.Event, 0)
	for rows.Next() {
		evn, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, evn)
	}

	return events, rows.Err()
}

// GetAllUnfinishedEvents отдаёт все ещё активные ивенты (pending/confirmed/
// in_progress — не finished, не declined и не canceled) вне
// зависимости от создателя — используется только при старте сервиса, чтобы
// восстановить цепочки таймеров controlEventTimerDenial, потерянные при
// рестарте (eventTimers — чисто in-memory карта).
func (r *postgresRepository) GetAllUnfinishedEvents(ctx context.Context) ([]domain.Event, error) {
	query := `
		SELECT event_id, name, user_create_id, enemy_side_leader_id, user_count, time_start, time_finish, create_time, winner_side, target_game_count, game_count, status, rent_server_sent
		FROM events
		WHERE status IN ('pending', 'confirmed', 'in_progress')
	`
	rows, err := r.db(ctx).Query(ctx, query)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]domain.Event, 0)
	for rows.Next() {
		evn, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, evn)
	}

	return events, rows.Err()
}

// IncrementEventGameCount возвращает новое значение game_count — сервису
// нужно именно оно, а не прочитанное раньше (оно могло устареть).
func (r *postgresRepository) IncrementEventGameCount(ctx context.Context, eventID string) (int64, error) {
	query := `
		UPDATE events
		SET game_count = game_count + 1
		WHERE event_id = $1
		RETURNING game_count
	`

	var gameCount int64
	err := r.db(ctx).QueryRow(ctx, query, eventID).Scan(&gameCount)
	return gameCount, err
}

// FinishEventDB переводит ивент в статус "finished" и одновременно пишет
// time_finish и итоговый winner_side — атомарно, одним UPDATE'ом (а не
// отдельным вызовом UpdateEventStatus + отдельным UPDATE на эти поля),
// и только из "in_progress", по тем же соображениям идемпотентности, что и
// UpdateEventStatus.
func (r *postgresRepository) FinishEventDB(ctx context.Context, eventID, winnerSide string, timeFinish time.Time) error {
	query := `
		UPDATE events
		SET status = 'finished', time_finish = $1, winner_side = $2
		WHERE event_id = $3 AND status = 'in_progress'
	`

	tag, err := r.db(ctx).Exec(ctx, query, timeFinish, winnerSide, eventID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return domain.FailedPrecondition("cannot finish event %q: it is not in_progress", eventID)
	}
	return nil
}

// MarkRentServerSent помечает, что сигнал об аренде сервера уже отправлен —
// после рестарта сервиса таймер аренды отрабатывает заново, и без этого
// флага rent.server уходил бы повторно.
func (r *postgresRepository) MarkRentServerSent(ctx context.Context, eventID string) error {
	query := `
		UPDATE events
		SET rent_server_sent = true
		WHERE event_id = $1
	`

	_, err := r.db(ctx).Exec(ctx, query, eventID)
	return err
}

func (r *postgresRepository) CreateTeam(ctx context.Context, team domain.Team) error {
	query := `
		INSERT INTO team (team_id, event_id, winner, side_leader_id, game_number, members_count, status, time_start, time_finish, total_kills, total_deaths, total_points, total_revival, total_destroyed_vehicles)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	// time_start/time_finish пишем как настоящий NULL, если ещё не заданы —
	// именно на этом основана логика "пустой time_start = игра не началась".
	var timeStart, timeFinish *time.Time
	if !team.TimeStart.IsZero() {
		timeStart = &team.TimeStart
	}
	if !team.TimeFinish.IsZero() {
		timeFinish = &team.TimeFinish
	}

	status := team.Status
	if status == "" {
		status = domain.TeamStatusPending
	}

	_, err := r.db(ctx).Exec(ctx, query, team.TeamID, team.EventID, team.Winner, team.SideLeaderID, team.GameNumber, team.MembersCount, status, timeStart, timeFinish, team.TotalKills, team.TotalDeaths, team.TotalPoints, team.TotalRevival, team.TotalDestroyedVehicles)
	return err
}

func (r *postgresRepository) GetTeamsByEventID(ctx context.Context, eventID string) ([]domain.Team, error) {
	query := `
		SELECT team_id, event_id, winner, side_leader_id, game_number, members_count, status, time_start, time_finish, total_kills, total_deaths, total_points, total_revival, total_destroyed_vehicles
		FROM team
		WHERE event_id = $1
		ORDER BY game_number ASC
	`

	rows, err := r.db(ctx).Query(ctx, query, eventID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := make([]domain.Team, 0)
	for rows.Next() {
		t, err := scanTeam(rows)
		if err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}

	return teams, rows.Err()
}

func (r *postgresRepository) GetTeamByID(ctx context.Context, teamID string) (domain.Team, error) {
	query := `
		SELECT team_id, event_id, winner, side_leader_id, game_number, members_count, status, time_start, time_finish, total_kills, total_deaths, total_points, total_revival, total_destroyed_vehicles
		FROM team
		WHERE team_id = $1
	`

	t, err := scanTeam(r.db(ctx).QueryRow(ctx, query, teamID))

	// Как и GetEventByID: "не найдено" — пустая структура, сервис сам
	// превращает её в "team not found".
	if err == pgx.ErrNoRows {
		return domain.Team{}, nil
	}
	if err != nil {
		return domain.Team{}, err
	}

	return t, nil
}

// scanTeam — общий разбор строки team. time_start/time_finish читаются через
// указатель: пока игра не началась/не закончилась, в БД там NULL.
func scanTeam(row pgx.Row) (domain.Team, error) {
	var t domain.Team
	var timeStart, timeFinish *time.Time

	err := row.Scan(
		&t.TeamID,
		&t.EventID,
		&t.Winner,
		&t.SideLeaderID,
		&t.GameNumber,
		&t.MembersCount,
		&t.Status,
		&timeStart,
		&timeFinish,
		&t.TotalKills,
		&t.TotalDeaths,
		&t.TotalPoints,
		&t.TotalRevival,
		&t.TotalDestroyedVehicles,
	)
	if err != nil {
		return domain.Team{}, err
	}

	if timeStart != nil {
		t.TimeStart = *timeStart
	}
	if timeFinish != nil {
		t.TimeFinish = *timeFinish
	}

	return t, nil
}

func (r *postgresRepository) JoinUserToTeam(ctx context.Context, teamID, userEventID string, role domain.Role) error {
	query := `
		INSERT INTO team_members (team_id, user_event_id, role)
		VALUES ($1, $2, $3)
	`

	if _, err := r.db(ctx).Exec(ctx, query, teamID, userEventID, role); err != nil {
		if isUniqueViolation(err) {
			return domain.AlreadyExists("user is already in team %q", teamID)
		}
		return err
	}

	updateQuery := `
		UPDATE team
		SET members_count = members_count + 1
		WHERE team_id = $1
	`

	_, err := r.db(ctx).Exec(ctx, updateQuery, teamID)
	return err
}

func (r *postgresRepository) RemoveUserFromTeam(ctx context.Context, teamID, userEventID string) error {
	query := `
		DELETE FROM team_members
		WHERE team_id = $1 AND user_event_id = $2
	`

	tag, err := r.db(ctx).Exec(ctx, query, teamID, userEventID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user is not a member of team %q", teamID)
	}

	updateQuery := `
		UPDATE team
		SET members_count = members_count - 1
		WHERE team_id = $1
	`

	_, err = r.db(ctx).Exec(ctx, updateQuery, teamID)
	return err
}

// RemoveUserFromAllTeams убирает игрока из всех команд ивента и возвращает
// id команд, из которых он реально был удалён (сервис по ним пересчитывает
// six_clan_members). Нужен при выходе из ивента: строку users иначе не
// удалить — на неё ссылается team_members.
func (r *postgresRepository) RemoveUserFromAllTeams(ctx context.Context, eventID, userEventID string) ([]string, error) {
	query := `
		DELETE FROM team_members tm
		USING team t
		WHERE tm.team_id = t.team_id
		AND t.event_id = $1
		AND tm.user_event_id = $2
		RETURNING tm.team_id
	`

	rows, err := r.db(ctx).Query(ctx, query, eventID, userEventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teamIDs := make([]string, 0)
	for rows.Next() {
		var teamID string
		if err := rows.Scan(&teamID); err != nil {
			return nil, err
		}
		teamIDs = append(teamIDs, teamID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(teamIDs) == 0 {
		return teamIDs, nil
	}

	updateQuery := `
		UPDATE team
		SET members_count = members_count - 1
		WHERE team_id = ANY($1)
	`

	if _, err := r.db(ctx).Exec(ctx, updateQuery, teamIDs); err != nil {
		return nil, err
	}

	return teamIDs, nil
}

// StartTeamGame стартует обе команды одной игры одним UPDATE'ом — они всегда
// начинаются вместе. Условие status = 'pending' не даёт стартовать игру
// повторно, даже если два запроса одновременно прошли проверку в сервисе.
func (r *postgresRepository) StartTeamGame(ctx context.Context, team1ID, team2ID string, timeStart time.Time) error {
	query := `
		UPDATE team
		SET time_start = $1, status = 'in_progress'
		WHERE team_id IN ($2, $3) AND status = 'pending'
	`

	tag, err := r.db(ctx).Exec(ctx, query, timeStart, team1ID, team2ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 2 {
		return domain.FailedPrecondition("cannot start game: it is already started or finished")
	}
	return nil
}

// SumTeamMemberStats складывает индивидуальную статистику всех team_members
// этой команды (их заносят по одному через AddTeamMemberStats — ПОСЛЕ того,
// как FinishTeamGame уже закрыл игру для этой команды) — вызывается из
// сервисного слоя после каждого AddTeamMemberStats, чтобы обновить
// Team.total_* и не требовать от клиента самому суммировать все числа.
func (r *postgresRepository) SumTeamMemberStats(ctx context.Context, teamID string) (kills, deaths, points, revival, destroyedVehicles int64, err error) {
	query := `
		SELECT COALESCE(SUM(kills), 0), COALESCE(SUM(deaths), 0), COALESCE(SUM(points), 0), COALESCE(SUM(revival), 0), COALESCE(SUM(destroyed_vehicles), 0)
		FROM team_members
		WHERE team_id = $1
	`

	err = r.db(ctx).QueryRow(ctx, query, teamID).Scan(&kills, &deaths, &points, &revival, &destroyedVehicles)
	return kills, deaths, points, revival, destroyedVehicles, err
}

func (r *postgresRepository) UpdateTeamTotals(ctx context.Context, teamID string, kills, deaths, points, revival, destroyedVehicles int64) error {
	query := `
		UPDATE team
		SET total_kills = $1, total_deaths = $2, total_points = $3, total_revival = $4, total_destroyed_vehicles = $5
		WHERE team_id = $6
	`

	_, err := r.db(ctx).Exec(ctx, query, kills, deaths, points, revival, destroyedVehicles, teamID)
	return err
}

// FinishTeamGame закрывает одну игру: обеим командам сразу проставляет
// time_finish и статус finished, а winner выставляется через сравнение
// team_id с winnerTeamID (сервисный слой уже проверил, что winnerTeamID — это
// team1ID либо team2ID). Закрыть можно только идущую игру — иначе повторный
// вызов перезаписал бы победителя и второй раз увеличил game_count.
func (r *postgresRepository) FinishTeamGame(ctx context.Context, team1ID, team2ID, winnerTeamID string, timeFinish time.Time) error {
	query := `
		UPDATE team
		SET time_finish = $1, winner = (team_id = $2), status = 'finished'
		WHERE team_id IN ($3, $4) AND status = 'in_progress'
	`

	tag, err := r.db(ctx).Exec(ctx, query, timeFinish, winnerTeamID, team1ID, team2ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 2 {
		return domain.FailedPrecondition("cannot finish game: it is not in progress")
	}
	return nil
}

// AddTeamMemberStats обновляет статистику уже существующего участника
// команды (team_members заполняется через JoinUserToTeam до игры) — если
// строки не было, значит игрок не состоял в этой команде, и это ошибка, а не
// молчаливый no-op.
func (r *postgresRepository) AddTeamMemberStats(ctx context.Context, teamID, userEventID string, kills, deaths, points, revival, destroyedVehicles int64) error {
	query := `
		UPDATE team_members
		SET kills = $1, deaths = $2, points = $3, revival = $4, destroyed_vehicles = $5
		WHERE team_id = $6 AND user_event_id = $7
	`

	tag, err := r.db(ctx).Exec(ctx, query, kills, deaths, points, revival, destroyedVehicles, teamID, userEventID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("player is not a member of team %q", teamID)
	}
	return nil
}

func (r *postgresRepository) GetTeamStats(ctx context.Context, teamID string) ([]domain.TeamMember, error) {
	query := `
		SELECT tm.team_id, tm.user_event_id, u.user_id, tm.role, tm.kills, tm.deaths, tm.points, tm.revival, tm.destroyed_vehicles, tm.six_clan_members
		FROM team_members tm
		JOIN users u ON tm.user_event_id = u.user_event_id
		WHERE tm.team_id = $1
	`

	rows, err := r.db(ctx).Query(ctx, query, teamID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make([]domain.TeamMember, 0)
	for rows.Next() {
		var m domain.TeamMember
		err := rows.Scan(
			&m.TeamID,
			&m.UserEventID,
			&m.UserID,
			&m.Role,
			&m.Kills,
			&m.Deaths,
			&m.Points,
			&m.Revival,
			&m.DestroyedVehicles,
			&m.SixClanMembers,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, m)
	}

	return stats, rows.Err()
}

// IsUserInTeam проверяет, состоит ли userEventID в команде teamID. Раньше
// здесь был баг: делали Query() и не читали ни строки результата, ни COUNT —
// функция всегда возвращала (true, nil), из-за чего JoinUserToTeam вообще не
// мог никого добавить (всегда решал, что игрок "уже в команде"). Плюс
// сравнивали с u.user_id, хотя параметром передаётся уже резолвленный
// user_event_id — join с users был не нужен вовсе.
func (r *postgresRepository) IsUserInTeam(ctx context.Context, teamID, userEventID string) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM team_members WHERE team_id = $1 AND user_event_id = $2)
	`

	var exists bool
	if err := r.db(ctx).QueryRow(ctx, query, teamID, userEventID).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

// GetUserEventIDByUserIDAndTeamID резолвит user_id (внешний, из другого
// микросервиса) в user_event_id (внутренний) в контексте конкретной команды.
// users не хранит team_id напрямую — привязку к ивенту (а через него — к
// команде) даёт join с team по event_id.
func (r *postgresRepository) GetUserEventIDByUserIDAndTeamID(ctx context.Context, userID, teamID string) (string, error) {
	query := `
		SELECT u.user_event_id
		FROM users u
		JOIN team t ON t.event_id = u.event_id
		WHERE u.user_id = $1 AND t.team_id = $2
	`

	var result string
	err := r.db(ctx).QueryRow(ctx, query, userID, teamID).Scan(&result)
	if err == pgx.ErrNoRows {
		return "", domain.NotFound("user %q not found in event of team %q", userID, teamID)
	}
	if err != nil {
		return "", err
	}

	return result, nil
}
