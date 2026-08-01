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

	if err := s.eventRepo.CreateEvent(ctx, event); err != nil {
		/*kafka*/
		return err
	}

	if err := s.eventRepo.JoinToEvent(ctx, userCreateID, event.ID, time.Now()); err != nil {
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

func (s *eventService) GetEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error) {
	event, err := s.eventRepo.GetEventsByEventName(ctx, eventName)
	if err != nil {
		/*kafka*/
		return []domain.Event{}, err
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

func (s *eventService) CreateGame(ctx context.Context, eventID, mapID string, timeStart time.Time) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		/*kafka*/
		return err
	}

	if event.ID == "" {
		/*kafka*/
		return errors.New("event not found")
	}

	game := domain.Game{
		ID:                   uuid.New().String(),
		EventID:              eventID,
		MapID:                mapID,
		Game_team_winner_id:  "",
		Game_team_loser_id:   "",
		TimeStart:            timeStart,
		TimeFinish:           time.Time{},
	}

	if err := s.eventRepo.CreateGame(ctx, game); err != nil {
		/*kafka*/
		return err
	}

	/*kafka*/
	return nil
}

func (s *eventService) GetGameByID(ctx context.Context, gameID string) (domain.Game, error) {
	game, err := s.eventRepo.GetGameByID(ctx, gameID)
	if err != nil {
		/*kafka*/
		return domain.Game{}, err
	}

	if game.ID == "" {
		/*kafka*/
		return domain.Game{}, errors.New("game not found")
	}

	/*kafka*/
	return game, nil
}

func (s *eventService) GetGamesByEventID(ctx context.Context, eventID string) ([]domain.Game, error) {
	games, err := s.eventRepo.GetGamesByEventID(ctx, eventID)
	if err != nil {
		/*kafka*/
		return nil, err
	}

	/*kafka*/
	return games, nil
}

func (s *eventService) UpdateGameWinner(ctx context.Context, gameID, winnerTeamID string) error {
	game, err := s.eventRepo.GetGameByID(ctx, gameID)
	if err != nil {
		/*kafka*/
		return err
	}

	if game.ID == "" {
		/*kafka*/
		return errors.New("game not found")
	}

	if err = s.eventRepo.UpdateGameWinner(ctx, gameID, winnerTeamID); err != nil {
		/*kafka*/
		return err
	}

	/*kafka*/
	return nil
}

func (s *eventService) UpdateGameLoser(ctx context.Context, gameID, loserTeamID string) error {
	game, err := s.eventRepo.GetGameByID(ctx, gameID)
	if err != nil {
		/*kafka*/
		return err
	}

	if game.ID == "" {
		/*kafka*/
		return errors.New("game not found")
	}

	if err = s.eventRepo.UpdateGameLoser(ctx, gameID, loserTeamID); err != nil {
		/*kafka*/
		return err
	}

	/*kafka*/
	return nil
}

func (s *eventService) FinishGame(ctx context.Context, gameID string, timeFinish time.Time) error {
	game, err := s.eventRepo.GetGameByID(ctx, gameID)
	if err != nil {
		/*kafka*/
		return err
	}

	if game.ID == "" {
		/*kafka*/
		return errors.New("game not found")
	}

	if err = s.eventRepo.FinishGame(ctx, gameID, timeFinish); err != nil {
		/*kafka*/
		return err
	}

	/*kafka*/
	return nil
}

func (s *eventService) DeleteGame(ctx context.Context, gameID string) error {
	game, err := s.eventRepo.GetGameByID(ctx, gameID)
	if err != nil {
		/*kafka*/
		return err
	}

	if game.ID == "" {
		/*kafka*/
		return errors.New("game not found")
	}

	if err = s.eventRepo.DeleteGame(ctx, gameID); err != nil {
		/*kafka*/
		return err
	}

	/*kafka*/
	return nil
}

func (s *eventService) AddUserStatsToGame(ctx context.Context, gameID, userID string, kills, deaths, points int64) error {
	game, err := s.eventRepo.GetGameByID(ctx, gameID)
	if err != nil {
		/*kafka*/
		return err
	}

	if game.ID == "" {
		/*kafka*/
		return errors.New("game not found")
	}

	user, err := s.eventRepo.GetUserByID(ctx, userID)
	if err != nil {
		/*kafka*/
		return err
	}

	if user.ID == "" {
		/*kafka*/
		return errors.New("user not found")
	}

	stats := domain.GameUserStats{
		ID:     uuid.New().String(),
		User:   user,
		Kills:  kills,
		Deaths: deaths,
		Points: points,
	}

	if err := s.eventRepo.AddUserStatsToGame(ctx, stats); err != nil {
		/*kafka*/
		return err
	}

	/*kafka*/
	return nil
}

func (s *eventService) GetGameStats(ctx context.Context, gameID string) ([]domain.GameUserStats, error) {
	game, err := s.eventRepo.GetGameByID(ctx, gameID)
	if err != nil {
		/*kafka*/
		return nil, err
	}

	if game.ID == "" {
		/*kafka*/
		return nil, errors.New("game not found")
	}

	stats, err := s.eventRepo.GetGameStats(ctx, gameID)
	if err != nil {
		/*kafka*/
		return nil, err
	}

	/*kafka*/
	return stats, nil
}

