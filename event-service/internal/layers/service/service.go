package service

import (
	"context"
	"time"
	"errors"

	domain "event-service/internal/core/domain"
	kafka "event-service/internal/core/kafka"

	uuid "github.com/google/uuid"
)

type eventService struct {
	eventRepo EventRepository
	producer  *kafka.Producer
}

func NewEventService(eventRepo EventRepository, producer *kafka.Producer) EventService {
	return &eventService{
		eventRepo: eventRepo,
		producer:  producer,
	}
}

func (s *eventService) CreateEvent(ctx context.Context, userCreateID, enemySideLeader, eventName string, timeStart time.Time) error {
	event := domain.Event{
		EventID: uuid.New().String(),
		Name: eventName,
		UserCreateID: userCreateID,
		EnemySideLeader: enemySideLeader,
		UserCount: 0,
		TimeStart: timeStart,
		CreateTime: time.Now(),
	}

	if err := s.eventRepo.CreateEvent(ctx, event); err != nil {
		return err
	}

	if err := s.eventRepo.JoinToEvent(ctx, userCreateID, event.EventID, time.Now()); err != nil {
		return err
	}

	if err := s.producer.PublishEventCreated(ctx, event); err != nil {
		return err
	}

	if err := s.producer.PublishUserJoinedEvent(ctx, event.EventID, userCreateID, time.Now()); err != nil {
		return err
	}

	return nil
}

func (s *eventService) GetEventsByUserID(ctx context.Context, userCreateID string) ([]*domain.Event, error) {
	events, err := s.eventRepo.GetEventsByUserID(ctx, userCreateID)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (s *eventService) GetEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error) {
	event, err := s.eventRepo.GetEventsByEventName(ctx, eventName)
	if err != nil {
		return []domain.Event{}, err
	}

	return event, nil
}

func (s *eventService) UpdateTimeEvent(ctx context.Context, eventID, userCreateID string, newTimeStart time.Time) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}

	if event.UserCreateID != userCreateID {
		return errors.New("you are not admin")
	}

	if err = s.eventRepo.UpdateTimeEvent(ctx, eventID, newTimeStart); err != nil {
		return err
	}

	if err = s.producer.PublishEventTimeUpdated(ctx, eventID, userCreateID, newTimeStart); err != nil {
		return err
	}

	return nil
}

func (s *eventService) DeleteEvent(ctx context.Context, eventID, userCreateID string) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}

	if event.UserCreateID != userCreateID {
		return errors.New("you are not admin")
	}

	if err = s.eventRepo.DeleteEvent(ctx, eventID); err != nil {
		return err
	}

	if err = s.producer.PublishEventDeleted(ctx, eventID, userCreateID); err != nil {
		return err
	}

	return nil
}

func (s *eventService) JoinToEvent(ctx context.Context, eventID, userID string, joinTime time.Time) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}

	if event.EventID == "" {
		return errors.New("event not found")
	}

	if err = s.eventRepo.JoinToEvent(ctx, userID, eventID, joinTime); err != nil {
		return err
	}

	if err = s.producer.PublishUserJoinedEvent(ctx, eventID, userID, joinTime); err != nil {
		return err
	}

	return nil
}

func (s *eventService) LeaveEvent(ctx context.Context, userID, eventID string) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}

	if event.EventID == "" {
		return errors.New("event not found")
	}

	if event.UserCreateID == userID {
		return errors.New("admin cannot leave event, use DeleteEvent instead")
	}

	if err = s.eventRepo.LeaveEvent(ctx, userID, eventID); err != nil {
		return err
	}

	if err = s.producer.PublishUserLeftEvent(ctx, eventID, userID); err != nil {
		return err
	}

	return nil
}

func (s *eventService) SetRole(ctx context.Context, eventID, sideLeaderID, userID string, role domain.Role) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}

	if event.EventID == "" {
		return errors.New("event not found")
	}

	if event.UserCreateID != sideLeaderID && event.EnemySideLeader != sideLeaderID {
		return errors.New("only side leaders can set roles")
	}

	user, err := s.eventRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.UserID == "" {
		return errors.New("user not found")
	}

	if err = s.eventRepo.UpdateUserRole(ctx, userID, eventID, role); err != nil {
		return err
	}

	return nil
}

func (s *eventService) CreateGame(ctx context.Context, eventID, mapName string, timeStart time.Time) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}

	if event.EventID == "" {
		return errors.New("event not found")
	}

	team1 := domain.Team{
		TeamID: uuid.New().String(),
		GameID: "",
	}

	team2 := domain.Team{
		TeamID: uuid.New().String(),
		GameID: "",
	}

	game := domain.Game{
		GameID:              uuid.New().String(),
		EventID:             eventID,
		UserCreateID:        event.UserCreateID,
		EnemySideLeader:     event.EnemySideLeader,
		Team1ID:             team1.TeamID,
		Team2ID:             team2.TeamID,
		MapName:             mapName,
		Game_team_winner_id: "",
		Game_team_loser_id:  "",
		TimeStart:           timeStart,
		TimeFinish:          time.Time{},
	}

	team1.GameID = game.GameID
	team2.GameID = game.GameID

	if err := s.eventRepo.CreateTeam(ctx, team1); err != nil {
		return err
	}

	if err := s.eventRepo.CreateTeam(ctx, team2); err != nil {
		return err
	}

	if err := s.eventRepo.CreateGame(ctx, game); err != nil {
		return err
	}

	if err := s.producer.PublishGameCreated(ctx, game); err != nil {
		return err
	}

	return nil
}

func (s *eventService) GetGameByID(ctx context.Context, gameID string) (domain.Game, error) {
	game, err := s.eventRepo.GetGameByID(ctx, gameID)
	if err != nil {
		return domain.Game{}, err
	}

	if game.GameID == "" {
		return domain.Game{}, errors.New("game not found")
	}

	return game, nil
}

func (s *eventService) GetGamesByEventID(ctx context.Context, eventID string) ([]domain.Game, error) {
	games, err := s.eventRepo.GetGamesByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	return games, nil
}

func (s *eventService) UpdateGameWinner(ctx context.Context, gameID, winnerTeamID string) error {
	game, err := s.eventRepo.GetGameByID(ctx, gameID)
	if err != nil {
		return err
	}

	if game.GameID == "" {
		return errors.New("game not found")
	}

	if err = s.eventRepo.UpdateGameWinner(ctx, gameID, winnerTeamID); err != nil {
		return err
	}

	if err = s.producer.PublishGameWinnerUpdated(ctx, gameID, winnerTeamID); err != nil {
		return err
	}

	return nil
}

func (s *eventService) UpdateGameLoser(ctx context.Context, gameID, loserTeamID string) error {
	game, err := s.eventRepo.GetGameByID(ctx, gameID)
	if err != nil {
		return err
	}

	if game.GameID == "" {
		return errors.New("game not found")
	}

	if err = s.eventRepo.UpdateGameLoser(ctx, gameID, loserTeamID); err != nil {
		return err
	}

	if err = s.producer.PublishGameLoserUpdated(ctx, gameID, loserTeamID); err != nil {
		return err
	}

	return nil
}

func (s *eventService) FinishGame(ctx context.Context, gameID string, timeFinish time.Time) error {
	game, err := s.eventRepo.GetGameByID(ctx, gameID)
	if err != nil {
		return err
	}

	if game.GameID == "" {
		return errors.New("game not found")
	}

	if err = s.eventRepo.FinishGame(ctx, gameID, timeFinish); err != nil {
		return err
	}

	if err = s.producer.PublishGameFinished(ctx, gameID, timeFinish); err != nil {
		return err
	}

	return nil
}

func (s *eventService) DeleteGame(ctx context.Context, gameID string) error {
	game, err := s.eventRepo.GetGameByID(ctx, gameID)
	if err != nil {
		return err
	}

	if game.GameID == "" {
		return errors.New("game not found")
	}

	if err = s.eventRepo.DeleteGame(ctx, gameID); err != nil {
		return err
	}

	if err = s.producer.PublishGameDeleted(ctx, gameID); err != nil {
		return err
	}

	return nil
}

func (s *eventService) AddUserStatsToGame(ctx context.Context, gameID, userID string, kills, deaths, points int64) error {
	game, err := s.eventRepo.GetGameByID(ctx, gameID)
	if err != nil {
		return err
	}

	if game.GameID == "" {
		return errors.New("game not found")
	}

	user, err := s.eventRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.UserID == "" {
		return errors.New("user not found")
	}

	stats := domain.GameUserStats{
		GameUserStatsID: uuid.New().String(),
		Game:   game,
		User:   user,
		Kills:  kills,
		Deaths: deaths,
		Points: points,
	}

	if err := s.eventRepo.AddUserStatsToGame(ctx, stats); err != nil {
		return err
	}

	if err := s.producer.PublishUserStatsAdded(ctx, stats); err != nil {
		return err
	}

	return nil
}

func (s *eventService) GetGameStats(ctx context.Context, gameID string) ([]domain.GameUserStats, error) {
	game, err := s.eventRepo.GetGameByID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if game.GameID == "" {
		return nil, errors.New("game not found")
	}

	stats, err := s.eventRepo.GetGameStats(ctx, gameID)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *eventService) GetTeamByID(ctx context.Context, teamID string) (domain.Team, error) {
	team, err := s.eventRepo.GetTeamByID(ctx, teamID)
	if err != nil {
		return domain.Team{}, err
	}

	if team.TeamID == "" {
		return domain.Team{}, errors.New("team not found")
	}

	return team, nil
}

func (s *eventService) AddUserToTeam(ctx context.Context, teamID, userID, clanID string, role domain.Role) error {
	team, err := s.eventRepo.GetTeamByID(ctx, teamID)
	if err != nil {
		return err
	}

	if team.TeamID == "" {
		return errors.New("team not found")
	}

	if err := s.eventRepo.AddUserToTeam(ctx, teamID, userID, clanID, role); err != nil {
		return err
	}

	return nil
}

func (s *eventService) RemoveUserFromTeam(ctx context.Context, teamID, userID string) error {
	team, err := s.eventRepo.GetTeamByID(ctx, teamID)
	if err != nil {
		return err
	}

	if team.TeamID == "" {
		return errors.New("team not found")
	}

	if err := s.eventRepo.RemoveUserFromTeam(ctx, teamID, userID); err != nil {
		return err
	}

	return nil
}
