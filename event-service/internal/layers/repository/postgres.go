package repository

import (
	"context"
	"time"

	"event-service/internal/core/domain"
	service "event-service/internal/layers/service"

	"github.com/jackc/pgx/v5"
	pgxpool "github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) service.EventRepository {
	return &postgresRepository{
		pool: pool,
	}
}

func (r *postgresRepository) CreateEvent(ctx context.Context, event domain.Event) error {
	query := `
	INSERT INTO events (event_id, name, user_create_id, enemy_side_leader_id, time_start, time_finish, create_time, user_count, target_game_count, game_count, is_started, is_finished)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.pool.Exec(ctx, query, event.EventID, event.Name, event.UserCreateID, event.EnemySideLeader, event.TimeStart, event.TimeFinish, event.CreateTime, event.UserCount, event.TargetGameCount, event.GameCount, event.IsStarted, event.IsFinished)
	if err != nil {
		return err
	}

	return nil
}

func (r *postgresRepository) GetEventsByCreatorId(ctx context.Context, userCreateID string) ([]*domain.Event, error) {
	query := `
		SELECT event_id, name, user_create_id, enemy_side_leader_id, user_count, time_start, time_finish, create_time, winner_side, is_started, is_finished, target_game_count, game_count
		FROM events
		WHERE user_create_id = $1
	`

	rows, err := r.pool.Query(ctx, query, userCreateID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]*domain.Event, 0)
	for rows.Next() {
		evn := &domain.Event{}
		err := rows.Scan(
			&evn.EventID,
			&evn.Name,
			&evn.UserCreateID,
			&evn.EnemySideLeader,
			&evn.UserCount,
			&evn.TimeStart,
			&evn.TimeFinish,
			&evn.CreateTime,
			&evn.WinnerSide,
			&evn.IsStarted,
			&evn.IsFinished,
			&evn.TargetGameCount,
			&evn.GameCount,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, evn)
	}

	return events, rows.Err()
}

func (r *postgresRepository) GetLastEventByCreatorId(ctx context.Context, userCreateID string) (*domain.Event, error) {
	query := `
		SELECT event_id, name, user_create_id, enemy_side_leader_id, user_count, time_start, time_finish, create_time, winner_side, is_started, is_finished, target_game_count, game_count
		FROM events
		WHERE user_create_id = $1
		ORDER BY create_time DESC
		LIMIT 1
	`

	evn := &domain.Event{}
	err := r.pool.QueryRow(ctx, query, userCreateID).Scan(
		&evn.EventID,
		&evn.Name,
		&evn.UserCreateID,
		&evn.EnemySideLeader,
		&evn.UserCount,
		&evn.TimeStart,
		&evn.TimeFinish,
		&evn.CreateTime,
		&evn.WinnerSide,
		&evn.IsStarted,
		&evn.IsFinished,
		&evn.TargetGameCount,
		&evn.GameCount,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return evn, nil
}

func (r *postgresRepository) GetEventByID(ctx context.Context, eventID string) (domain.Event, error) {
	query := `
		SELECT event_id, name, user_create_id, enemy_side_leader_id, user_count, time_start, time_finish, create_time, winner_side, is_started, is_finished, target_game_count, game_count
		FROM events
		WHERE event_id = $1
	`
	var event domain.Event
	err := r.pool.QueryRow(ctx, query, eventID).Scan(
		&event.EventID,
		&event.Name,
		&event.UserCreateID,
		&event.EnemySideLeader,
		&event.UserCount,
		&event.TimeStart,
		&event.TimeFinish,
		&event.CreateTime,
		&event.WinnerSide,
		&event.IsStarted,
		&event.IsFinished,
		&event.TargetGameCount,
		&event.GameCount,
	)

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
		SELECT user_event_id, user_id, event_id, clan_id, enemy, role, six_clan_members, join_time
		FROM users
		WHERE event_id = $1
	`

	rows, err := r.pool.Query(ctx, query, eventID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	users := make([]domain.User, 0)

	for rows.Next() {
		usr := domain.User{}
		err := rows.Scan(
			&usr.UserEventID,
			&usr.UserID,
			&usr.EventID,
			&usr.ClanID,
			&usr.Enemy,
			&usr.Role,
			&usr.SixClanMembers,
			&usr.JoinTime,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, usr)
	}

	return users, rows.Err()
}

func (r *postgresRepository) GetEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error) {
	query := `
		SELECT event_id, name, user_create_id, enemy_side_leader_id, user_count, time_start, time_finish, create_time, winner_side, is_started, is_finished, target_game_count, game_count
		FROM events
		WHERE name = $1
	`
	rows, err := r.pool.Query(ctx, query, eventName)
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	events := make([]domain.Event, 0)

	for rows.Next() {
		evn := domain.Event{}
		err := rows.Scan(
			&evn.EventID,
			&evn.Name,
			&evn.UserCreateID,
			&evn.EnemySideLeader,
			&evn.UserCount,
			&evn.TimeStart,
			&evn.TimeFinish,
			&evn.CreateTime,
			&evn.WinnerSide,
			&evn.IsStarted,
			&evn.IsFinished,
			&evn.TargetGameCount,
			&evn.GameCount,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, evn)
	}

	return events, rows.Err()
}

func (r *postgresRepository) UpdateTimeEvent(ctx context.Context, eventID string, newTimeStart time.Time) error {
	query := `
		UPDATE events
		SET time_start = $1
		WHERE event_id = $2
	`

	_, err := r.pool.Exec(ctx, query, newTimeStart, eventID)
	return err
}

func (r *postgresRepository) UpdateTimeFinishEvent(ctx context.Context, eventID string, newTimeFinish time.Time) error {
	query := `
		UPDATE events
		SET time_finish = $1
		WHERE event_id = $2
	`

	_, err := r.pool.Exec(ctx, query, newTimeFinish, eventID)
	return err
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

	_, err := r.pool.Exec(ctx, query, newName, eventID, oldName)
	return err
}

func (r *postgresRepository) DeleteEvent(ctx context.Context, eventID string) error {
	query := `
		DELETE FROM events WHERE event_id = $1
	`

	_, err := r.pool.Exec(ctx, query, eventID)
	return err
}

func (r *postgresRepository) IsUserInEvent(ctx context.Context, eventID, userID string) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM users WHERE event_id = $1 AND user_id = $2)
	`

	var exists bool
	if err := r.pool.QueryRow(ctx, query, eventID, userID).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *postgresRepository) JoinToEvent(ctx context.Context, userEventID, eventID, userID, clanID string, enemy bool, joinTime time.Time) error {
	query := `
		INSERT INTO users (user_event_id, user_id, event_id, clan_id, enemy, role, six_clan_members, join_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	sixClanMembers := false
	_, err := r.pool.Exec(ctx, query, userEventID, userID, eventID, clanID, enemy, domain.RolePlayer, sixClanMembers, joinTime)
	if err != nil {
		return err
	}

	updateQuery := `
		UPDATE events
		SET user_count = user_count + 1
		WHERE event_id = $1
	`

	_, err = r.pool.Exec(ctx, updateQuery, eventID)
	return err
}

func (r *postgresRepository) LeaveEvent(ctx context.Context, userID, eventID string) error {
	query := `
		DELETE FROM users
		WHERE user_id = $1 AND event_id = $2
	`

	_, err := r.pool.Exec(ctx, query, userID, eventID)
	if err != nil {
		return err
	}

	updateQuery := `
		UPDATE events
		SET user_count = user_count - 1
		WHERE event_id = $1
	`

	_, err = r.pool.Exec(ctx, updateQuery, eventID)
	return err
}

func (r *postgresRepository) UpdateUserRole(ctx context.Context, userID, eventID string, role domain.Role) error {
	query := `
		UPDATE users
		SET role = $1
		WHERE user_id = $2 AND event_id = $3
	`

	_, err := r.pool.Exec(ctx, query, string(role), userID, eventID)
	return err
}

func (r *postgresRepository) GetUserByID(ctx context.Context, eventID, userID string) (domain.User, error) {
	query := `
		SELECT user_event_id, user_id, event_id, clan_id, enemy, role, six_clan_members, join_time
		FROM users
		WHERE user_id = $1 AND event_id = $2
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, userID, eventID).Scan(
		&user.UserEventID,
		&user.UserID,
		&user.EventID,
		&user.ClanID,
		&user.Enemy,
		&user.Role,
		&user.SixClanMembers,
		&user.JoinTime,
	)

	if err == pgx.ErrNoRows {
		return domain.User{}, nil
	}

	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (r *postgresRepository) CheckSixClanMembers(ctx context.Context, eventID, userID, clanID string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM users
		WHERE event_id = $1
		AND clan_id = $2
		AND clan_id != ''
		AND user_id != $3
	`

	var count int
	err := r.pool.QueryRow(ctx, query, eventID, clanID, userID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count >= 5, nil
}

func (r *postgresRepository) UpdateUserSixClanMembers(ctx context.Context, userID, eventID string, hasSixClanMembers bool) error {
	query := `
		UPDATE users
		SET six_clan_members = $1
		WHERE user_id = $2 AND event_id = $3
	`

	_, err := r.pool.Exec(ctx, query, hasSixClanMembers, userID, eventID)
	return err
}

func (r *postgresRepository) UpdateSixClanMembersForClan(ctx context.Context, eventID, clanID string, hasSixClanMembers bool) error {
	query := `
		UPDATE users
		SET six_clan_members = $1
		WHERE event_id = $2 AND clan_id = $3
	`

	_, err := r.pool.Exec(ctx, query, hasSixClanMembers, eventID, clanID)
	return err
}

func (r *postgresRepository) CountClanMembersInEvent(ctx context.Context, eventID, clanID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM users
		WHERE event_id = $1 AND clan_id = $2
	`

	var count int
	err := r.pool.QueryRow(ctx, query, eventID, clanID).Scan(&count)
	return count, err
}

func (r *postgresRepository) GetUserIDsByEventID(ctx context.Context, eventID string) ([]string, error) {
	query := `
		SELECT user_id
		FROM users
		WHERE event_id = $1
	`

	rows, err := r.pool.Query(ctx, query, eventID)
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

// events.is_confirmed убрана из схемы (её не было и на новой диаграмме) —
// подтверждение/отклонение по достижению 80% теперь не персистится отдельным
// флагом: отклонение = реальное DeleteEvent, подтверждение — просто Kafka-событие
// из сервисного слоя.

func (r *postgresRepository) GetUnfinishedEventsByUserID(ctx context.Context, userCreateID string) ([]*domain.Event, error) {
	query := `
		SELECT event_id, name, user_create_id, enemy_side_leader_id, user_count, time_start, time_finish, create_time, winner_side, is_started, is_finished, target_game_count, game_count
		FROM events
		WHERE user_create_id = $1 AND is_finished = false
	`

	rows, err := r.pool.Query(ctx, query, userCreateID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]*domain.Event, 0)
	for rows.Next() {
		evn := &domain.Event{}
		err := rows.Scan(
			&evn.EventID,
			&evn.Name,
			&evn.UserCreateID,
			&evn.EnemySideLeader,
			&evn.UserCount,
			&evn.TimeStart,
			&evn.TimeFinish,
			&evn.CreateTime,
			&evn.WinnerSide,
			&evn.IsStarted,
			&evn.IsFinished,
			&evn.TargetGameCount,
			&evn.GameCount,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, evn)
	}

	return events, rows.Err()
}

func (r *postgresRepository) GetUnfinishedEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error) {
	query := `
		SELECT event_id, name, user_create_id, enemy_side_leader_id, user_count, time_start, time_finish, create_time, winner_side, is_started, is_finished, target_game_count, game_count
		FROM events
		WHERE name = $1 AND is_finished = false
	`
	rows, err := r.pool.Query(ctx, query, eventName)
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	events := make([]domain.Event, 0)

	for rows.Next() {
		evn := domain.Event{}
		err := rows.Scan(
			&evn.EventID,
			&evn.Name,
			&evn.UserCreateID,
			&evn.EnemySideLeader,
			&evn.UserCount,
			&evn.TimeStart,
			&evn.TimeFinish,
			&evn.CreateTime,
			&evn.WinnerSide,
			&evn.IsStarted,
			&evn.IsFinished,
			&evn.TargetGameCount,
			&evn.GameCount,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, evn)
	}

	return events, rows.Err()
}

func (r *postgresRepository) StartEventDB(ctx context.Context, eventID string) error {
	query := `
		UPDATE events
		SET is_started = true
		WHERE event_id = $1
	`

	_, err := r.pool.Exec(ctx, query, eventID)
	return err
}

func (r *postgresRepository) IncrementEventGameCount(ctx context.Context, eventID string) error {
	query := `
		UPDATE events
		SET game_count = game_count + 1
		WHERE event_id = $1
	`

	_, err := r.pool.Exec(ctx, query, eventID)
	return err
}

func (r *postgresRepository) FinishEventDB(ctx context.Context, eventID, winnerSide string) error {
	query := `
		UPDATE events
		SET is_finished = true, time_finish = $1, winner_side = $2
		WHERE event_id = $3
	`

	_, err := r.pool.Exec(ctx, query, time.Now(), winnerSide, eventID)
	return err
}

func (r *postgresRepository) CreateTeam(ctx context.Context, team domain.Team) error {
	query := `
		INSERT INTO team (team_id, event_id, winner, side_leader_id, game_number, members_count, time_start, time_finish, kills, deaths, revival, equipment_destroyed)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
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

	_, err := r.pool.Exec(ctx, query, team.TeamID, team.EventID, team.Winner, team.SideLeaderID, team.GameNumber, team.MembersCount, timeStart, timeFinish, team.Kills, team.Deaths, team.Revival, team.EquipmentDestroyed)
	return err
}

func (r *postgresRepository) GetTeamsByEventID(ctx context.Context, eventID string) ([]domain.Team, error) {
	query := `
		SELECT team_id, event_id, winner, side_leader_id, game_number, members_count, time_start, time_finish, kills, deaths, revival, equipment_destroyed
		FROM team
		WHERE event_id = $1
		ORDER BY game_number ASC
	`

	rows, err := r.pool.Query(ctx, query, eventID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := make([]domain.Team, 0)
	for rows.Next() {
		var t domain.Team
		var timeStart, timeFinish *time.Time
		err := rows.Scan(
			&t.TeamID,
			&t.EventID,
			&t.Winner,
			&t.SideLeaderID,
			&t.GameNumber,
			&t.MembersCount,
			&timeStart,
			&timeFinish,
			&t.Kills,
			&t.Deaths,
			&t.Revival,
			&t.EquipmentDestroyed,
		)
		if err != nil {
			return nil, err
		}
		if timeStart != nil {
			t.TimeStart = *timeStart
		}
		if timeFinish != nil {
			t.TimeFinish = *timeFinish
		}
		teams = append(teams, t)
	}

	return teams, rows.Err()
}

func (r *postgresRepository) GetTeamByID(ctx context.Context, teamID string) (domain.Team, error) {
	query := `
		SELECT team_id, event_id, winner, side_leader_id, game_number, members_count, time_start, time_finish, kills, deaths, revival, equipment_destroyed
		FROM team
		WHERE team_id = $1
	`

	var t domain.Team
	var timeStart, timeFinish *time.Time
	err := r.pool.QueryRow(ctx, query, teamID).Scan(
		&t.TeamID,
		&t.EventID,
		&t.Winner,
		&t.SideLeaderID,
		&t.GameNumber,
		&t.MembersCount,
		&timeStart,
		&timeFinish,
		&t.Kills,
		&t.Deaths,
		&t.Revival,
		&t.EquipmentDestroyed,
	)

	if err == pgx.ErrNoRows {
		return domain.Team{}, nil
	}
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

func (r *postgresRepository) AddUserToTeam(ctx context.Context, teamID, userEventID string, role domain.Role) error {
	query := `
		INSERT INTO team_members (team_id, user_event_id, role)
		VALUES ($1, $2, $3)
	`

	if _, err := r.pool.Exec(ctx, query, teamID, userEventID, role); err != nil {
		return err
	}

	updateQuery := `
		UPDATE team
		SET members_count = members_count + 1
		WHERE team_id = $1
	`

	_, err := r.pool.Exec(ctx, updateQuery, teamID)
	return err
}

func (r *postgresRepository) RemoveUserFromTeam(ctx context.Context, teamID, userEventID string) error {
	query := `
		DELETE FROM team_members
		WHERE team_id = $1 AND user_event_id = $2
	`

	if _, err := r.pool.Exec(ctx, query, teamID, userEventID); err != nil {
		return err
	}

	updateQuery := `
		UPDATE team
		SET members_count = members_count - 1
		WHERE team_id = $1
	`

	_, err := r.pool.Exec(ctx, updateQuery, teamID)
	return err
}

func (r *postgresRepository) StartTeamGame(ctx context.Context, teamID string, timeStart time.Time) error {
	query := `
		UPDATE team
		SET time_start = $1
		WHERE team_id = $2
	`

	_, err := r.pool.Exec(ctx, query, timeStart, teamID)
	return err
}

func (r *postgresRepository) FinishTeamGame(ctx context.Context, teamID string, timeFinish time.Time, winner bool, kills, deaths, revival, equipmentDestroyed int64) error {
	query := `
		UPDATE team
		SET time_finish = $1, winner = $2, kills = $3, deaths = $4, revival = $5, equipment_destroyed = $6
		WHERE team_id = $7
	`

	_, err := r.pool.Exec(ctx, query, timeFinish, winner, kills, deaths, revival, equipmentDestroyed, teamID)
	return err
}

func (r *postgresRepository) AddTeamMemberStats(ctx context.Context, teamID, userEventID string, kills, deaths, points int64) error {
	query := `
		UPDATE team_members
		SET kills = $1, deaths = $2, points = $3
		WHERE team_id = $4 AND user_event_id = $5
	`

	_, err := r.pool.Exec(ctx, query, kills, deaths, points, teamID, userEventID)
	return err
}

func (r *postgresRepository) GetTeamStats(ctx context.Context, teamID string) ([]domain.TeamMember, error) {
	query := `
		SELECT team_id, user_event_id, role, kills, deaths, points
		FROM team_members
		WHERE team_id = $1
	`

	rows, err := r.pool.Query(ctx, query, teamID)
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
			&m.Role,
			&m.Kills,
			&m.Deaths,
			&m.Points,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, m)
	}

	return stats, rows.Err()
}
