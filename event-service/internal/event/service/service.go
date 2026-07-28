package service

import (
	"context"
	"time"
	"errors"

	domain "event-service/internal/core/domain"

	uuid "github.com/google/uuid"
)

type eventService struct {
	eventRepo EventRepository
}

func NewEventService(eventRepo EventRepository) EventService {
	return &eventService{
		eventRepo: eventRepo,
	}
}

func (s *eventService) CreateEvent(ctx context.Context, userCreateID, eventName string, timeStart time.Time) error {
	event := domain.Event{
		ID: uuid.New().String(),
		Name: eventName,
		UserCreateID: userCreateID,
		UserCount: 0,
		TimeStart: timeStart,
		CreateTime: time.Now(),
	}

	_, err := s.eventRepo.GetEventByName(ctx, eventName)
	if err != nil {
		/*kafka*/
		return err
	}

	if err = s.eventRepo.CreateEvent(ctx, event); err != nil {
		/*kafka*/
		return err
	}

	if err = s.eventRepo.JoinToEvent(ctx, userCreateID, event.ID, time.Now()); err != nil {
		/*kafka*/
		return err
	}

	/*kafka*/
	return nil
}

func (s *eventService) GetEventsByUserID(ctx context.Context, userCreateID string) ([]*domain.Event, error) {
	events, err := s.eventRepo.GetEventsByUserID(ctx, userCreateID)
	if err != nil {
		/*kafka*/
		return nil, err
	}

	/*kafka*/
	return events, nil
}

func (s *eventService) GetEventByName(ctx context.Context, eventName string) (domain.Event, error) {
	event, err := s.eventRepo.GetEventByName(ctx, eventName)
	if err != nil {
		/*kafka*/
		return domain.Event{}, err
	}

	/*kafka*/
	return event, nil
}

func (s *eventService) UpdateTimeEvent(ctx context.Context, eventID, userCreateID string, newTimeStart time.Time) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		/*kafka*/
		return err
	}

	if event.UserCreateID != userCreateID {
		/*kafka*/
		return errors.New("you are not admin")
	}

	if err = s.eventRepo.UpdateTimeEvent(ctx, eventID, newTimeStart); err != nil {
		/*kafka*/
		return err
	}

	/*kafka*/
	return nil
}

func (s *eventService) DeleteEvent(ctx context.Context, eventID, userCreateID string) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		/*kafka*/
		return err
	}

	if event.UserCreateID != userCreateID {
		/*kafka*/
		return errors.New("you are not admin")
	}

	if err = s.eventRepo.DeleteEvent(ctx, eventID); err != nil {
		/*kafka*/
		return err
	}

	/*kafka*/
	return nil
}

func (s *eventService) JoinToEvent(ctx context.Context, eventID, userID string, joinTime time.Time) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		/*kafka*/
		return err
	}

	if event.ID == "" {
		/*kafka*/
		return errors.New("event not found")
	}

	if err = s.eventRepo.JoinToEvent(ctx, userID, eventID, joinTime); err != nil {
		/*kafka*/
		return err
	}

	/*kafka*/
	return nil
}

func (s *eventService) LeaveEvent(ctx context.Context, userID, eventID string) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		/*kafka*/
		return err
	}

	if event.ID == "" {
		/*kafka*/
		return errors.New("event not found")
	}

	if event.UserCreateID == userID {
		/*kafka*/
		return errors.New("admin cannot leave event, use DeleteEvent instead")
	}

	if err = s.eventRepo.LeaveEvent(ctx, userID, eventID); err != nil {
		/*kafka*/
		return err
	}

	/*kafka*/
	return nil
}
