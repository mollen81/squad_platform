package service

import (
	"context"
	"time"
	domain "event-service/internal/core/domain"
)

type EventService interface {
	CreateEvent(ctx context.Context, userCreateID int64, timeStart time.Time) error
	GetEventsByUserID(ctx context.Context, userCreateID int64) ([]*domain.Event, error)
	GetEventByName(ctx context.Context, eventName string) (domain.Event, error)
	UpdateTimeEvent(ctx context.Context, userCreateID int64, newTimeStart time.Time) error
	DeleteEvent(ctx context.Context, userCreateID int64) error
}

type EventRepository interface {
	CreateEvent(ctx context.Context, event domain.Event) error
	GetEventsByUserID(ctx context.Context, userCreateID int64) ([]*domain.Event, error)
	GetEventByName(ctx context.Context, eventName string) (domain.Event, error)
	UpdateTimeEvent(ctx context.Context, eventID int64, newTimeStart time.Time) error
	RenameEvent(ctx context.Context, eventID int64, oldName, newName string) error
	DeleteEvent(ctx context.Context, eventID int64) error
	JoinToEvent(ctx context.Context, userID, eventID int64, joinTime time.Time) error
	LeaveEvent(ctx context.Context, userID, eventID int64) error
}
