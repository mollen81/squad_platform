package service

import (
	"context"
	"time"
	domain "event-service/internal/core/domain"
)

type EventService interface {
	CreateEvent(ctx context.Context, userCreateID, eventName string, timeStart time.Time) error
	GetEventsByUserID(ctx context.Context, userCreateID string) ([]*domain.Event, error)
	GetEventByName(ctx context.Context, eventName string) (domain.Event, error)
	UpdateTimeEvent(ctx context.Context, eventID, userCreateID string, newTimeStart time.Time) error
	DeleteEvent(ctx context.Context, eventID, userCreateID string) error
	JoinToEvent(ctx context.Context, eventID, userID string, joinTime time.Time) error
	LeaveEvent(ctx context.Context, userID, eventID string) error
}

type EventRepository interface {
	CreateEvent(ctx context.Context, event domain.Event) error
	GetEventsByUserID(ctx context.Context, userCreateID string) ([]*domain.Event, error)
	GetEventByName(ctx context.Context, eventName string) (domain.Event, error)
	GetEventByID(ctx context.Context, eventID string) (domain.Event, error)
	UpdateTimeEvent(ctx context.Context, eventID string, newTimeStart time.Time) error
	RenameEvent(ctx context.Context, eventID, oldName, newName string) error
	DeleteEvent(ctx context.Context, eventID string) error
	JoinToEvent(ctx context.Context, userID, eventID string, joinTime time.Time) error
	LeaveEvent(ctx context.Context, userID, eventID string) error
}
