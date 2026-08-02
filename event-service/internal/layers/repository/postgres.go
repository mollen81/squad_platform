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
	INSERT INTO events (id, name, user_create_id, time_start, time_finish, create_time, user_count, event_team_winner, event_team_loser)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query, event.EventID, event.Name, event.UserCreateID, event.TimeStart, event.TimeFinish, event.CreateTime, event.UserCount, event.Event_team_winner, event.Event_team_loser)
	if err != nil {
		return err
	}

	return nil
}

func (r *postgresRepository) GetEventsByUserID(ctx context.Context, userCreateID string) ([]*domain.Event, error) {
	query := `
		SELECT id, name, user_create_id, user_count, time_start, time_finish, create_time, event_team_winner, event_team_loser
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
			&evn.UserCount,
			&evn.TimeStart,
			&evn.TimeFinish,
			&evn.CreateTime,
			&evn.Event_team_winner,
			&evn.Event_team_loser,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, evn)
	}

	return events, rows.Err()
}

func (r *postgresRepository) GetEventByID(ctx context.Context, eventID string) (domain.Event, error) {
	query := `
		SELECT id, name, user_create_id, user_count, time_start, time_finish, create_time, event_team_winner, event_team_loser
		FROM events
		WHERE id = $1
	`
	var event domain.Event
	err := r.pool.QueryRow(ctx, query, eventID).Scan(
		&event.EventID,
		&event.Name,
		&event.UserCreateID,
		&event.UserCount,
		&event.TimeStart,
		&event.TimeFinish,
		&event.CreateTime,
		&event.Event_team_winner,
		&event.Event_team_loser,
	)

	if err == pgx.ErrNoRows {
		return domain.Event{}, nil
	}

	if err != nil {
		return domain.Event{}, err
	}

	return event, nil
}

func (r *postgresRepository) GetEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error) {
	query := `
		SELECT id, name, user_create_id, user_count, time_start, time_finish, create_time, event_team_winner, event_team_loser
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
			&evn.UserCount,
			&evn.TimeStart,
			&evn.TimeFinish,
			&evn.CreateTime,
			&evn.Event_team_winner,
			&evn.Event_team_loser,
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
		WHERE id = $2
	`

	_, err := r.pool.Exec(ctx, query, newTimeStart, eventID)
	return err
}

func (r *postgresRepository) UpdateTimeFinishEvent(ctx context.Context, eventID string, newTimeFinish time.Time) error {
	query := `
		UPDATE events
		SET time_finish = $1
		WHERE id = $2
	`

	_, err := r.pool.Exec(ctx, query, newTimeFinish, eventID)
	return err
}

func (r *postgresRepository) UpdateEventWinner(ctx context.Context, eventID, winnerTeamID string) error {
	query := `
		UPDATE events
		SET event_team_winner = $1
		WHERE id = $2
	`

	_, err := r.pool.Exec(ctx, query, winnerTeamID, eventID)
	return err
}

func (r *postgresRepository) UpdateEventLoser(ctx context.Context, eventID, loserTeamID string) error {
	query := `
		UPDATE events
		SET event_team_loser = $1
		WHERE id = $2
	`

	_, err := r.pool.Exec(ctx, query, loserTeamID, eventID)
	return err
}

func (r *postgresRepository) RenameEvent(ctx context.Context, eventID, oldName, newName string) error {
	query := `
		UPDATE events
		SET name = $1
		WHERE id = $2 AND name = $3
	`

	_, err := r.pool.Exec(ctx, query, newName, eventID, oldName)
	return err
}

func (r *postgresRepository) DeleteEvent(ctx context.Context, eventID string) error {
	query := `
		DELETE FROM events WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, eventID)
	return err
}

func (r *postgresRepository) JoinToEvent(ctx context.Context, userID, eventID string, joinTime time.Time) error {
	query := `
		INSERT INTO users (id, event_id, join_time)
		VALUES ($1, $2, $3)
	`

	_, err := r.pool.Exec(ctx, query, userID, eventID, joinTime)
	return err
}

func (r *postgresRepository) LeaveEvent(ctx context.Context, userID, eventID string) error {
	query := `
		DELETE FROM users
		WHERE id = $1 AND event_id = $2
	`

	_, err := r.pool.Exec(ctx, query, userID, eventID)
	return err
}

func (r *postgresRepository) CreateGame(ctx context.Context, game domain.Game) error {
	query := `
		INSERT INTO games (id, event_id, map_id, game_team_winner_id, game_team_loser_id, time_start, time_finish)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(ctx, query, game.GameID, game.EventID, game.MapID, game.Game_team_winner_id, game.Game_team_loser_id, game.TimeStart, game.TimeFinish)
	return err
}

func (r *postgresRepository) GetGameByID(ctx context.Context, gameID string) (domain.Game, error) {
	query := `
		SELECT id, event_id, map_id, game_team_winner_id, game_team_loser_id, time_start, time_finish
		FROM games
		WHERE id = $1
	`

	var game domain.Game
	err := r.pool.QueryRow(ctx, query, gameID).Scan(
		&game.GameID,
		&game.EventID,
		&game.MapID,
		&game.Game_team_winner_id,
		&game.Game_team_loser_id,
		&game.TimeStart,
		&game.TimeFinish,
	)

	if err == pgx.ErrNoRows {
		return domain.Game{}, nil
	}

	if err != nil {
		return domain.Game{}, err
	}

	return game, nil
}

func (r *postgresRepository) GetGamesByEventID(ctx context.Context, eventID string) ([]domain.Game, error) {
	query := `
		SELECT id, event_id, map_id, game_team_winner_id, game_team_loser_id, time_start, time_finish
		FROM games
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

	games := make([]domain.Game, 0)
	for rows.Next() {
		var game domain.Game
		err := rows.Scan(
			&game.GameID,
			&game.EventID,
			&game.MapID,
			&game.Game_team_winner_id,
			&game.Game_team_loser_id,
			&game.TimeStart,
			&game.TimeFinish,
		)
		if err != nil {
			return nil, err
		}
		games = append(games, game)
	}

	return games, rows.Err()
}

func (r *postgresRepository) UpdateGameWinner(ctx context.Context, gameID, winnerTeamID string) error {
	query := `
		UPDATE games
		SET game_team_winner_id = $1
		WHERE id = $2
	`

	_, err := r.pool.Exec(ctx, query, winnerTeamID, gameID)
	return err
}

func (r *postgresRepository) UpdateGameLoser(ctx context.Context, gameID, loserTeamID string) error {
	query := `
		UPDATE games
		SET game_team_loser_id = $1
		WHERE id = $2
	`

	_, err := r.pool.Exec(ctx, query, loserTeamID, gameID)
	return err
}

func (r *postgresRepository) FinishGame(ctx context.Context, gameID string, timeFinish time.Time) error {
	query := `
		UPDATE games
		SET time_finish = $1
		WHERE id = $2
	`

	_, err := r.pool.Exec(ctx, query, timeFinish, gameID)
	return err
}

func (r *postgresRepository) DeleteGame(ctx context.Context, gameID string) error {
	query := `
		DELETE FROM games WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, gameID)
	return err
}

func (r *postgresRepository) GetUserByID(ctx context.Context, userID string) (domain.User, error) {
	query := `
		SELECT id, clan_id, team_id, role
		FROM users
		WHERE id = $1
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&user.UserID,
		&user.ClanID,
		&user.TeamID,
		&user.Role,
	)

	if err == pgx.ErrNoRows {
		return domain.User{}, nil
	}

	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (r *postgresRepository) AddUserStatsToGame(ctx context.Context, stats domain.GameUserStats) error {
	query := `
		INSERT INTO game_user_stats (id, game_id, user_id, clan_id, team_id, role, kills, deaths, points)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query, stats.GameUserStatsID, stats.Game.GameID, stats.User.UserID, stats.User.ClanID, stats.User.TeamID, stats.User.Role, stats.Kills, stats.Deaths, stats.Points)
	return err
}

func (r *postgresRepository) GetGameStats(ctx context.Context, gameID string) ([]domain.GameUserStats, error) {
	query := `
		SELECT id, user_id, clan_id, team_id, role, kills, deaths, points
		FROM game_user_stats
		WHERE game_id = $1
	`

	rows, err := r.pool.Query(ctx, query, gameID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make([]domain.GameUserStats, 0)
	for rows.Next() {
		var stat domain.GameUserStats
		err := rows.Scan(
			&stat.GameUserStatsID,
			&stat.User.UserID,
			&stat.User.ClanID,
			&stat.User.TeamID,
			&stat.User.Role,
			&stat.Kills,
			&stat.Deaths,
			&stat.Points,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, rows.Err()
}
