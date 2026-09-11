package service

import (
	"context"
	"time"
	"errors"
	"sync"

	domain "event-service/internal/core/domain"
	kafka "event-service/internal/core/kafka"

	uuid "github.com/google/uuid"
)

type eventService struct {
	eventRepo EventRepository
	producer  *kafka.Producer
	eventTimers map[string]context.CancelFunc
	timersMutex sync.Mutex
}

func NewEventService(eventRepo EventRepository, producer *kafka.Producer) EventService {
	return &eventService{
		eventRepo: eventRepo,
		producer:  producer,
		eventTimers: make(map[string]context.CancelFunc),
	}
}

func (s *eventService) CreateEvent(ctx context.Context, userCreatorID, creatorClanID, enemySideLeaderID, enemySideLeaderClanID, eventName string, timeStart time.Time, targetGameCount int64) error {
	now := time.Now()
	minTime := now.Add(45 * time.Minute)
	maxTime := now.Add(45 * 24 * time.Hour)

	if timeStart.Before(minTime) {
		return errors.New("event start time must be at least 45 minutes from now")
	}

	if timeStart.After(maxTime) {
		return errors.New("event start time must be within 45 days from now")
	}

	if targetGameCount <= 0 {
		return errors.New("target game count must be positive")
	}

	event := domain.Event{
		EventID: uuid.New().String(),
		Name: eventName,
		UserCreateID: userCreatorID,
		EnemySideLeader: enemySideLeaderID,
		UserCount: 0,
		TimeStart: timeStart,
		CreateTime: time.Now(),
		TargetGameCount: targetGameCount,
	}

	controlTime := timeStart.Add(-30 * time.Minute)
	if err := s.controlEventTimerDenial(event.EventID, controlTime); err != nil {
		return err
	}

	if err := s.eventRepo.CreateEvent(ctx, event); err != nil {
		return err
	}

	enemy := false
	if err := s.JoinToEvent(ctx, event.EventID, userCreatorID, creatorClanID, enemy); err != nil {
		return err
	}

	if err := s.eventRepo.UpdateUserRole(ctx, userCreatorID, event.EventID, domain.RoleSquadLeader); err != nil {
		return err
	}

	if err := s.producer.PublishEventCreated(ctx, event); err != nil {
		return err
	}

	enemy = true
	if err := s.JoinToEvent(ctx, event.EventID, enemySideLeaderID, enemySideLeaderClanID, enemy); err != nil {
		return err
	}

	if err := s.eventRepo.UpdateUserRole(ctx, enemySideLeaderID, event.EventID, domain.RoleSquadLeader); err != nil {
		return err
	}

	// Team.side_leader_id ссылается на users.user_event_id (внутренний id),
	// а не на сырой внешний user_id — поэтому достаём уже присвоенные user_event_id
	// обоих сайд-лидеров после того, как они реально заджойнились в ивент.
	creatorUser, err := s.eventRepo.GetUserByID(ctx, event.EventID, userCreatorID)
	if err != nil {
		return err
	}

	enemyUser, err := s.eventRepo.GetUserByID(ctx, event.EventID, enemySideLeaderID)
	if err != nil {
		return err
	}

	for gameNumber := int64(1); gameNumber <= targetGameCount; gameNumber++ {
		allyTeam := domain.Team{
			TeamID:       uuid.New().String(),
			EventID:      event.EventID,
			SideLeaderID: creatorUser.UserEventID,
			GameNumber:   gameNumber,
		}
		enemyTeam := domain.Team{
			TeamID:       uuid.New().String(),
			EventID:      event.EventID,
			SideLeaderID: enemyUser.UserEventID,
			GameNumber:   gameNumber,
		}

		if err := s.eventRepo.CreateTeam(ctx, allyTeam); err != nil {
			return err
		}

		if err := s.eventRepo.CreateTeam(ctx, enemyTeam); err != nil {
			return err
		}
	}

	return nil
}

func (s *eventService) GetEventsByCreatorId(ctx context.Context, userCreateID string) ([]*domain.Event, error) {
	events, err := s.eventRepo.GetEventsByCreatorId(ctx, userCreateID)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (s *eventService) GetLastEventByCreatorId(ctx context.Context, userCreateID string) (*domain.Event, error) {
	event, err := s.eventRepo.GetLastEventByCreatorId(ctx, userCreateID)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *eventService) GetEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error) {
	event, err := s.eventRepo.GetEventsByEventName(ctx, eventName)
	if err != nil {
		return []domain.Event{}, err
	}

	return event, nil
}

func (s *eventService) GetEventMembersList(ctx context.Context, eventID string) ([]domain.User, error) {
	events, err := s.eventRepo.GetEventMembersList(ctx, eventID)
	if err != nil {
		return []domain.User{}, err
	}

	return events, nil
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

	controlTime := newTimeStart.Add(-30 * time.Minute)
	if err := s.controlEventTimerDenial(eventID, controlTime); err != nil {
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

func (s *eventService) JoinToEvent(ctx context.Context, eventID, userID, clanID string, enemy bool) error {
	if eventID == "" {
		return errors.New("must have event id")
	}

	if userID == "" {
		return errors.New("must have user id")
	}

	if clanID == "" {
		return errors.New("must have clanID")
	}

	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}

	if event.EventID == "" {
		return errors.New("event not found")
	}

	alreadyJoined, err := s.eventRepo.IsUserInEvent(ctx, eventID, userID)
	if err != nil {
		return err
	}

	if alreadyJoined {
		return errors.New("user already joined this event")
	}

	userEventID := uuid.New().String()
	joinTime := time.Now()

	if err = s.eventRepo.JoinToEvent(ctx, userEventID, eventID, userID, clanID, enemy, joinTime); err != nil {
		return err
	}

	// Раньше это проверялось батчем в StartEvent по всем team.Members разом;
	// команды (Team) в новой схеме такого ростера не хранят, так что проверяем
	// сразу на джойне. hasSixClanMembers считает "других" участников этого же
	// клана в этом же ивенте (CheckSixClanMembers исключает самого себя) — если
	// их уже 5 и больше, значит вместе с только что зашедшим — 6+.
	hasSixClanMembers, err := s.eventRepo.CheckSixClanMembers(ctx, eventID, userID, clanID)
	if err != nil {
		return err
	}

	if hasSixClanMembers {
		// Флаг относится ко всей группе из одного клана в этом ивенте, а не
		// только к тому, кто зашёл последним и перевалил порог — иначе первые
		// 5 человек так и останутся с six_clan_members=false навсегда.
		if err := s.eventRepo.UpdateSixClanMembersForClan(ctx, eventID, clanID, true); err != nil {
			return err
		}
	} else {
		if err := s.eventRepo.UpdateUserSixClanMembers(ctx, userID, eventID, false); err != nil {
			return err
		}
	}

	if err = s.producer.PublishUserJoinedEvent(ctx, eventID, userID, clanID, enemy, joinTime); err != nil {
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

	leavingUser, err := s.eventRepo.GetUserByID(ctx, eventID, userID)
	if err != nil {
		return err
	}

	if err = s.eventRepo.LeaveEvent(ctx, userID, eventID); err != nil {
		return err
	}

	// Симметрично JoinToEvent: после выхода из клана в этом ивенте может стать
	// меньше 6 человек — если так, снимаем six_clan_members у оставшихся,
	// иначе флаг так и останется true уже после того, как условие перестало
	// выполняться.
	if leavingUser.ClanID != "" {
		clanMembersLeft, err := s.eventRepo.CountClanMembersInEvent(ctx, eventID, leavingUser.ClanID)
		if err != nil {
			return err
		}

		if clanMembersLeft < 6 {
			if err := s.eventRepo.UpdateSixClanMembersForClan(ctx, eventID, leavingUser.ClanID, false); err != nil {
				return err
			}
		}
	}

	if err = s.producer.PublishUserLeftEvent(ctx, eventID, userID); err != nil {
		return err
	}

	return nil
}

func (s *eventService) SetRole(ctx context.Context, eventID, yourID, userID string, role domain.Role) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}

	if event.EventID == "" {
		return errors.New("event not found")
	}

	if event.UserCreateID != yourID && event.EnemySideLeader != yourID {
		return errors.New("only side leaders can set roles")
	}

	if role == domain.RoleSquadLeader {
		return errors.New("squad_leader role cannot be assigned via SetRole")
	}

	if role != domain.RoleSideLeader && role != domain.RolePlayer {
		return errors.New("invalid role: only side_leader or player can be set via SetRole")
	}

	user, err := s.eventRepo.GetUserByID(ctx, eventID, userID)
	if err != nil {
		return err
	}

	if user.UserID == "" {
		return errors.New("user not found")
	}

	if err = s.eventRepo.UpdateUserRole(ctx, userID, eventID, role); err != nil {
		return err
	}

	if err = s.producer.PublishUserRoleChanged(ctx, eventID, userID, role); err != nil {
		return err
	}

	return nil
}

// startEventAndGames запускается только изнутри controlEventTimerDenial/confirmEvent80
// (по таймеру, в момент event.TimeStart), а не по ручке — публичного StartEvent
// RPC больше нет. Помечает ивент начатым и стартует первую по счёту игру
// (game_number = 1) сразу для обеих сторон; остальные игры стартуют по мере
// того, как заканчиваются предыдущие (через StartTeamGame по каждой team отдельно).
func (s *eventService) startEventAndGames(ctx context.Context, eventID string) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}

	if event.EventID == "" {
		return errors.New("event not found")
	}

	if err := s.eventRepo.StartEventDB(ctx, eventID); err != nil {
		return err
	}

	if err := s.producer.PublishEventStarted(ctx, eventID, event.TimeStart); err != nil {
		return err
	}

	teams, err := s.eventRepo.GetTeamsByEventID(ctx, eventID)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, team := range teams {
		if team.GameNumber != 1 {
			continue
		}

		if err := s.eventRepo.StartTeamGame(ctx, team.TeamID, now); err != nil {
			return err
		}

		if err := s.producer.PublishTeamGameStarted(ctx, team.TeamID, team.GameNumber, now); err != nil {
			return err
		}
	}

	return nil
}

func (s *eventService) controlEventTimerDenial(eventID string, controlTime time.Time) error {
	duration := time.Until(controlTime)

	s.timersMutex.Lock()
	if cancelFunc, exists := s.eventTimers[eventID]; exists {
		cancelFunc()
	}

	timerCtx, cancel := context.WithCancel(context.Background())
	s.eventTimers[eventID] = cancel
	s.timersMutex.Unlock()

	go func() {
		select {
		case <-time.After(duration):
			event, err := s.eventRepo.GetEventByID(context.Background(), eventID)
			if err != nil {
				return
			}

			if event.UserCount >= 80 {
				s.confirmEvent80(context.Background(), eventID)
			} else {
				s.declineEvent80(context.Background(), eventID)
			}

			s.timersMutex.Lock()
			delete(s.eventTimers, eventID)
			s.timersMutex.Unlock()
		case <-timerCtx.Done():
			return
		}
	}()

	return nil
}

func (s *eventService) confirmEvent80(ctx context.Context, eventID string) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}

	if event.EventID == "" {
		return errors.New("event not found")
	}

	if err := s.producer.PublishEventConfirmed(ctx, eventID); err != nil {
		return err
	}

	rentServerTime := time.Until(event.TimeStart.Add(-15 * time.Minute))
	if rentServerTime > 0 {
		go func() {
			time.Sleep(rentServerTime)

			playersList, err := s.eventRepo.GetUserIDsByEventID(context.Background(), eventID)
			if err != nil {
				return
			}

			if err := s.producer.PublishRentServer(context.Background(), eventID, playersList, event.TimeStart); err != nil {
				return
			}
		}()
	}

	duration := time.Until(event.TimeStart)
	if duration > 0 {
		go func() {
			time.Sleep(duration)
			s.startEventAndGames(context.Background(), eventID)
		}()
	}

	return nil
}

func (s *eventService) declineEvent80(ctx context.Context, eventID string) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}

	if event.EventID == "" {
		return errors.New("event not found")
	}

	// Не набрали 80 человек к контрольному времени — ивент реально отменяется
	// (удаляется), а не просто помечается флагом.
	if err := s.eventRepo.DeleteEvent(ctx, eventID); err != nil {
		return err
	}

	if err := s.producer.PublishEventDeclined(ctx, eventID); err != nil {
		return err
	}

	return nil
}

func (s *eventService) GetUnfinishedEventsByUserID(ctx context.Context, userCreateID string) ([]*domain.Event, error) {
	events, err := s.eventRepo.GetUnfinishedEventsByUserID(ctx, userCreateID)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (s *eventService) GetUnfinishedEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error) {
	events, err := s.eventRepo.GetUnfinishedEventsByEventName(ctx, eventName)
	if err != nil {
		return []domain.Event{}, err
	}

	return events, nil
}

// FinishEvent как отдельный метод убран: ивент теперь завершается автоматически
// внутри FinishTeamGame, когда game_count после этой игры достигает target_game_count.

func (s *eventService) GetTeamsByEventID(ctx context.Context, eventID string) ([]domain.Team, error) {
	return s.eventRepo.GetTeamsByEventID(ctx, eventID)
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

func (s *eventService) AddUserToTeam(ctx context.Context, teamID, userEventID string, role domain.Role) error {
	team, err := s.eventRepo.GetTeamByID(ctx, teamID)
	if err != nil {
		return err
	}

	if team.TeamID == "" {
		return errors.New("team not found")
	}

	return s.eventRepo.AddUserToTeam(ctx, teamID, userEventID, role)
}

func (s *eventService) RemoveUserFromTeam(ctx context.Context, teamID, userEventID string) error {
	team, err := s.eventRepo.GetTeamByID(ctx, teamID)
	if err != nil {
		return err
	}

	if team.TeamID == "" {
		return errors.New("team not found")
	}

	return s.eventRepo.RemoveUserFromTeam(ctx, teamID, userEventID)
}

func (s *eventService) StartTeamGame(ctx context.Context, teamID string) error {
	team, err := s.eventRepo.GetTeamByID(ctx, teamID)
	if err != nil {
		return err
	}

	if team.TeamID == "" {
		return errors.New("team not found")
	}

	if !team.TimeStart.IsZero() {
		return errors.New("game already started")
	}

	now := time.Now()
	if err := s.eventRepo.StartTeamGame(ctx, teamID, now); err != nil {
		return err
	}

	return s.producer.PublishTeamGameStarted(ctx, teamID, team.GameNumber, now)
}

// FinishTeamGame закрывает игру для одной стороны (team_id — это одна из двух
// команд, играющих под одним team.game_number). Когда закрыты обе стороны этой
// игры, засчитывается game_count ивента; когда game_count доходит до
// target_game_count — ивент считается завершённым (это заменяет старый
// отдельный FinishEvent RPC).
func (s *eventService) FinishTeamGame(ctx context.Context, teamID string, winner bool, kills, deaths, revival, equipmentDestroyed int64) error {
	team, err := s.eventRepo.GetTeamByID(ctx, teamID)
	if err != nil {
		return err
	}

	if team.TeamID == "" {
		return errors.New("team not found")
	}

	if !team.TimeFinish.IsZero() {
		return errors.New("game already finished")
	}

	if err := s.eventRepo.FinishTeamGame(ctx, teamID, time.Now(), winner, kills, deaths, revival, equipmentDestroyed); err != nil {
		return err
	}

	if err := s.producer.PublishTeamGameFinished(ctx, teamID, winner, time.Now()); err != nil {
		return err
	}

	event, err := s.eventRepo.GetEventByID(ctx, team.EventID)
	if err != nil {
		return err
	}

	teams, err := s.eventRepo.GetTeamsByEventID(ctx, team.EventID)
	if err != nil {
		return err
	}

	// ищем "второй" team этой же игры (тот же game_number, другой team_id)
	var sibling *domain.Team
	for i := range teams {
		if teams[i].GameNumber == team.GameNumber && teams[i].TeamID != teamID {
			sibling = &teams[i]
			break
		}
	}

	siblingFinished := sibling != nil && !sibling.TimeFinish.IsZero()
	if !siblingFinished {
		// вторая сторона ещё не закрыла игру — ждём её, засчитывать game_count рано
		return nil
	}

	if err := s.eventRepo.IncrementEventGameCount(ctx, team.EventID); err != nil {
		return err
	}

	newGameCount := event.GameCount + 1
	if newGameCount < event.TargetGameCount {
		return nil
	}

	// это была последняя запланированная игра — считаем итог по всем играм
	// и завершаем ивент целиком. team.side_leader_id — это users.user_event_id,
	// а не сырой event.user_create_id/enemy_side_leader_id, поэтому сначала
	// достаём internal id обоих сайд-лидеров, чтобы было с чем сравнивать team.Winner.
	creatorUser, err := s.eventRepo.GetUserByID(ctx, event.EventID, event.UserCreateID)
	if err != nil {
		return err
	}

	enemyUser, err := s.eventRepo.GetUserByID(ctx, event.EventID, event.EnemySideLeader)
	if err != nil {
		return err
	}

	allyWins, enemyWins := 0, 0
	for _, t := range teams {
		if !t.Winner {
			continue
		}
		switch t.SideLeaderID {
		case creatorUser.UserEventID:
			allyWins++
		case enemyUser.UserEventID:
			enemyWins++
		}
	}

	winnerSide := ""
	if allyWins > enemyWins {
		winnerSide = "ally"
	} else if enemyWins > allyWins {
		winnerSide = "enemy"
	}

	if err := s.eventRepo.FinishEventDB(ctx, team.EventID, winnerSide); err != nil {
		return err
	}

	if err := s.producer.PublishEventFinished(ctx, team.EventID, winnerSide, time.Now()); err != nil {
		return err
	}

	return nil
}

func (s *eventService) AddTeamMemberStats(ctx context.Context, teamID, userEventID string, kills, deaths, points int64) error {
	team, err := s.eventRepo.GetTeamByID(ctx, teamID)
	if err != nil {
		return err
	}

	if team.TeamID == "" {
		return errors.New("team not found")
	}

	if err := s.eventRepo.AddTeamMemberStats(ctx, teamID, userEventID, kills, deaths, points); err != nil {
		return err
	}

	return s.producer.PublishTeamMemberStatsAdded(ctx, teamID, userEventID, kills, deaths, points)
}

func (s *eventService) GetTeamStats(ctx context.Context, teamID string) ([]domain.TeamMember, error) {
	return s.eventRepo.GetTeamStats(ctx, teamID)
}
