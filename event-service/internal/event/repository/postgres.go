package repository

import (
	"context"
	"time"

	"event-service/internal/core/domain"
	service "event-service/internal/event/service"

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
	INSERT INTO events (id, name, user_create_id, time_start, create_time, user_count)
		VALUE ($1, $2, $3, $4, $5, 0)
	`

	_, err := r.pool.Exec(ctx, query, event.ID, event.Name, event.UserCreateID, event.TimeStart, event.CreateTime)
	if err != nil {
		return err
	}

	return nil
}

func (r *postgresRepository) GetEventsByUserID(ctx context.Context, userCreateID int64) ([]*domain.Event, error) {
	query := `
		SELECT id, name, user_create_id, time_start, create_time, user_count
		FROM event
		WHERE user_create_id = &1
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
			&evn.ID,
			&evn.Name,
			&evn.UserCreateID,
			&evn.UserCount,
			&evn.TimeStart,
			&evn.CreateTime,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, evn)
	}

	return events, rows.Err()
}

func (r *postgresRepository) GetEventByName(ctx context.Context, eventName string) (domain.Event, error) {
	query := `
		SELECT id, name, user_create_id, time_start, create_time, user_count
		FROM event
		WHERE name = &1
	`
	var event domain.Event
	err := r.pool.QueryRow(ctx, query, eventName).Scan(
		&event.ID,
		&event.Name,
		&event.UserCreateID,
		&event.UserCount,
		&event.TimeStart,
		&event.CreateTime,
	)

	if err == pgx.ErrNoRows {
		return domain.Event{}, nil
	}

	if err != nil {
		return domain.Event{}, err
	}

	return event, nil
}

func (r *postgresRepository) UpdateTimeEvent(ctx context.Context, eventID int64, newTimeStart time.Time) error {
	query := `
		UPDATE events
		SET time_start = $1
		WHERE id = $2
	`

	_, err := r.pool.Exec(ctx, query, newTimeStart, eventID)
	return err
}

func (r *postgresRepository) RenameEvent(ctx context.Context, eventID int64, oldName, newName string) error {
	query := `
		UPDATE events
		SET name = $1
		WHERE id = $2 AND name = $3
	`

	_, err := r.pool.Exec(ctx, query, newName, eventID, oldName)
	return err
}

func (r *postgresRepository) DeleteEvent(ctx context.Context, eventID int64) error {
	query := `
		DELETE FROM events WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, eventID)
	return err
}

func (r *postgresRepository) JoinToEvent(ctx context.Context, userID, eventID int64, joinTime time.Time) error {
	query := `
		INSERT INTO users (id, event_id, join_time)
		VALUE ($1, $2, $3)
	`

	_, err := r.pool.Exec(ctx, query, userID, eventID, joinTime)
	return err
}

func (r *postgresRepository) LeaveEvent(ctx context.Context, userID, eventID int64) error {
	query := `
		DELETE FROM users
		WHERE id = $1 AND event_id = $2
	`

	_, err := r.pool.Exec(ctx, query, userID, eventID)
	return err
}