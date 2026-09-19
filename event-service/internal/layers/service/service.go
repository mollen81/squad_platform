package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	domain "event-service/internal/core/domain"
	kafka "event-service/internal/core/kafka"

	uuid "github.com/google/uuid"
)

const (
	// Минимальное количество игроков для подтверждения ивента
	MinPlayersRequired = 80
	// За сколько до старта проверяется количество игроков
	CheckTimeBeforeStart = 30 * time.Minute
	// За сколько до старта отправляется сообщение об аренде сервера
	RentServerTimeBeforeStart = 15 * time.Minute
)

type eventService struct {
	eventRepo   EventRepository
	producer    *kafka.Producer
	eventTimers map[string]context.CancelFunc
	timersMutex sync.Mutex
}

func NewEventService(eventRepo EventRepository, producer *kafka.Producer) EventService {
	return &eventService{
		eventRepo:   eventRepo,
		producer:    producer,
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

	if targetGameCount < 1 || targetGameCount > 3 {
		return errors.New("target game count must be 1, 2 or 3")
	}

	event := domain.Event{
		EventID:         uuid.New().String(),
		Name:            eventName,
		UserCreateID:    userCreatorID,
		EnemySideLeader: enemySideLeaderID,
		UserCount:       0,
		TimeStart:       timeStart,
		CreateTime:      time.Now(),
		TargetGameCount: targetGameCount,
		Status:          domain.EventStatusPending,
	}

	controlTime := timeStart.Add(-CheckTimeBeforeStart)
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

	if event.EventID == "" {
		return errors.New("event not found")
	}

	if event.UserCreateID != userCreateID {
		return errors.New("you are not admin")
	}

	// Время можно менять только в статусе pending
	if event.Status != domain.EventStatusPending {
		return fmt.Errorf("cannot change time: event status is %s (must be pending)", event.Status)
	}

	now := time.Now()
	minTime := now.Add(45 * time.Minute)
	maxTime := now.Add(45 * 24 * time.Hour)

	if newTimeStart.Before(minTime) {
		return errors.New("event start time must be at least 45 minutes from now")
	}

	if newTimeStart.After(maxTime) {
		return errors.New("event start time must be within 45 days from now")
	}

	if err = s.eventRepo.UpdateTimeEvent(ctx, eventID, newTimeStart); err != nil {
		return err
	}

	if err = s.producer.PublishEventTimeUpdated(ctx, eventID, userCreateID, newTimeStart); err != nil {
		return err
	}

	controlTime := newTimeStart.Add(-CheckTimeBeforeStart)
	if err := s.controlEventTimerDenial(eventID, controlTime); err != nil {
		return err
	}

	return nil
}

// CancelEvent — ручная отмена ивента создателем: доступна, пока ивент ещё
// pending или confirmed. Статус становится "canceled" (в отличие от
// "declined", в который автоматически уходит ивент, не набравший минимум
// игроков к контрольной точке — см. declineEvent). Сама строка события и вся
// история участников/команд никуда не удаляются — просто больше не активны.
func (s *eventService) CancelEvent(ctx context.Context, eventID, userCreateID string) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}

	if event.EventID == "" {
		return errors.New("event not found")
	}

	if event.UserCreateID != userCreateID {
		return errors.New("you are not admin")
	}

	if event.Status != domain.EventStatusPending && event.Status != domain.EventStatusConfirmed {
		return fmt.Errorf("cannot cancel event: current status is %q", event.Status)
	}

	if err = s.eventRepo.UpdateEventStatus(ctx, eventID, domain.EventStatusCanceled); err != nil {
		return err
	}

	// Иначе, если ивент отменили до controlTime, забытая горутина проснётся
	// позже и попытается прочитать уже отменённый ивент.
	s.cancelEventTimer(eventID)

	if err = s.producer.PublishEventCanceled(ctx, eventID, userCreateID); err != nil {
		return err
	}

	return nil
}

// cancelEventTimer отменяет запланированный controlEventTimerDenial для
// eventID, если он ещё не сработал, и убирает запись из мапы. Безопасно
// вызывать, даже если таймера для этого eventID нет.
func (s *eventService) cancelEventTimer(eventID string) {
	s.timersMutex.Lock()
	defer s.timersMutex.Unlock()

	if cancelFunc, exists := s.eventTimers[eventID]; exists {
		cancelFunc()
		delete(s.eventTimers, eventID)
	}
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

	// Присоединяться можно только к pending/confirmed ивентам
	if event.Status != domain.EventStatusPending && event.Status != domain.EventStatusConfirmed {
		return fmt.Errorf("cannot join: event status is %s", event.Status)
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
		return errors.New("admin cannot leave event, use CancelEvent instead")
	}

	// Выходить можно только из pending/confirmed ивентов
	if event.Status != domain.EventStatusPending && event.Status != domain.EventStatusConfirmed {
		return fmt.Errorf("cannot leave: event status is %s", event.Status)
	}

	// Проверяем, что уходящий вообще состоит в этом ивенте — иначе просто
	// молча не удалится ни одной строки, а ошибки не будет.
	if _, err := s.eventRepo.GetUserByID(ctx, eventID, userID); err != nil {
		return err
	}

	if err = s.eventRepo.LeaveEvent(ctx, userID, eventID); err != nil {
		return err
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

// startEventAndGames запускается только изнутри confirmEvent (по таймеру, в
// момент event.TimeStart), а не по ручке — публичного StartEvent RPC больше
// нет. Меняет статус на in_progress и стартует первую по счёту игру
// (game_number = 1) сразу для обеих сторон одним вызовом StartTeamGame;
// остальные игры стартуют по мере того, как заканчиваются предыдущие (через
// публичный StartTeamGame(team1, team2) для каждой следующей пары).
func (s *eventService) startEventAndGames(ctx context.Context, eventID string) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}

	if event.EventID == "" {
		return errors.New("event not found")
	}

	if err := s.eventRepo.UpdateEventStatus(ctx, eventID, domain.EventStatusInProgress); err != nil {
		return err
	}

	if err := s.producer.PublishEventStarted(ctx, eventID, event.TimeStart); err != nil {
		return err
	}

	teams, err := s.eventRepo.GetTeamsByEventID(ctx, eventID)
	if err != nil {
		return err
	}

	var team1ID, team2ID string
	for _, team := range teams {
		if team.GameNumber != 1 {
			continue
		}
		if team1ID == "" {
			team1ID = team.TeamID
		} else {
			team2ID = team.TeamID
		}
	}

	if team1ID == "" || team2ID == "" {
		return fmt.Errorf("event %q: expected 2 teams for game 1, found team1=%q team2=%q", eventID, team1ID, team2ID)
	}

	now := time.Now()
	if err := s.eventRepo.StartTeamGame(ctx, team1ID, team2ID, now); err != nil {
		return err
	}

	if err := s.producer.PublishTeamGameStarted(ctx, team1ID, 1, now); err != nil {
		return err
	}
	if err := s.producer.PublishTeamGameStarted(ctx, team2ID, 1, now); err != nil {
		return err
	}

	return nil
}

// controlEventTimerDenial — единственный источник правды по времени старта
// ивента: за CheckTimeBeforeStart (2 минуты) до event.TimeStart проверяет,
// набралось ли MinPlayersRequired (4 игрока), и либо подтверждает ивент
// (confirmEvent), либо отклоняет его (declineEvent). Повторный вызов для того
// же eventID (например, из UpdateTimeEvent) отменяет предыдущий таймер и ставит
// новый на его место.
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
		case <-sleepChan(duration):
			event, err := s.eventRepo.GetEventByID(context.Background(), eventID)
			if err != nil {
				log.Printf("controlEventTimerDenial: get event %q: %v", eventID, err)
				return
			}

			if event.EventID == "" {
				// Не должно происходить (ивенты больше не удаляются, только
				// переводятся в другой статус), но на всякий случай.
				return
			}

			if event.Status != domain.EventStatusPending {
				// Ивент уже не pending — например, создатель успел отменить
				// его вручную (CancelEvent) до срабатывания таймера.
				// cancelEventTimer уже должен был остановить эту горутину, но
				// на случай гонки (таймер сработал одновременно с
				// CancelEvent) на всякий случай ничего больше не делаем.
				return
			}

			if event.UserCount >= MinPlayersRequired {
				if err := s.confirmEvent(context.Background(), eventID); err != nil {
					log.Printf("confirmEvent(%q): %v", eventID, err)
				}
			} else {
				if err := s.declineEvent(context.Background(), eventID); err != nil {
					log.Printf("declineEvent(%q): %v", eventID, err)
				}
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

// sleepChan — как time.After(d), но для d <= 0 отдаёт уже закрытый канал,
// то есть срабатывает немедленно, а не через (по факту) случайный интервал,
// который получился бы из отрицательной длительности. Нужен, потому что
// после рестарта сервиса (RecoverPendingEvents) или после переноса времени
// слишком близко к текущему моменту controlTime/rentServerTime/TimeStart
// вполне могут оказаться уже в прошлом — и тогда действие должно произойти
// сразу, а не быть потеряно.
func sleepChan(d time.Duration) <-chan time.Time {
	if d <= 0 {
		ch := make(chan time.Time, 1)
		ch <- time.Now()
		return ch
	}
	return time.After(d)
}

// confirmEvent — набрали минимум игроков к контрольной точке. Меняет статус
// на confirmed, оповещает Kafka и расставляет два оставшихся таймера: за
// RentServerTimeBeforeStart (1 минуту) до старта — сигнал "пора арендовать
// сервер", и в момент старта — фактический запуск ивента и первой игры.
func (s *eventService) confirmEvent(ctx context.Context, eventID string) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}

	if event.EventID == "" {
		return errors.New("event not found")
	}

	if err := s.eventRepo.UpdateEventStatus(ctx, eventID, domain.EventStatusConfirmed); err != nil {
		return err
	}

	if err := s.producer.PublishEventConfirmed(ctx, eventID); err != nil {
		return err
	}

	s.scheduleRentAndStart(eventID, event.TimeStart)

	return nil
}

// scheduleRentAndStart ставит две отложенные горутины: rent_server в Kafka за
// RentServerTimeBeforeStart (1 минуту) до event.TimeStart, и запуск ивента в
// момент event.TimeStart. Вынесено отдельно от confirmEvent, чтобы
// RecoverPendingEvents могло переставить их заново после рестарта процесса, не
// публикуя повторно event_confirmed для уже подтверждённого ивента.
func (s *eventService) scheduleRentAndStart(eventID string, timeStart time.Time) {
	rentServerTime := time.Until(timeStart.Add(-RentServerTimeBeforeStart))
	go func() {
		<-sleepChan(rentServerTime)

		playersList, err := s.eventRepo.GetUserIDsByEventID(context.Background(), eventID)
		if err != nil {
			log.Printf("confirmEvent(%q): get players for rent_server: %v", eventID, err)
			return
		}

		if err := s.producer.PublishRentServer(context.Background(), eventID, playersList, timeStart); err != nil {
			log.Printf("confirmEvent(%q): publish rent_server: %v", eventID, err)
		}
	}()

	startDuration := time.Until(timeStart)
	go func() {
		<-sleepChan(startDuration)
		if err := s.startEventAndGames(context.Background(), eventID); err != nil {
			log.Printf("startEventAndGames(%q): %v", eventID, err)
		}
	}()
}

// declineEvent — к контрольной точке не набралось минимум игроков, ивент
// переходит в статус "declined" (не удаляется — история участников/команд
// остаётся, просто ивент больше не активен).
func (s *eventService) declineEvent(ctx context.Context, eventID string) error {
	event, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}

	if event.EventID == "" {
		return errors.New("event not found")
	}

	if err := s.eventRepo.UpdateEventStatus(ctx, eventID, domain.EventStatusDeclined); err != nil {
		return err
	}

	if err := s.producer.PublishEventDeclined(ctx, eventID); err != nil {
		return err
	}

	return nil
}

// RecoverPendingEvents переживает рестарт процесса: eventTimers — чисто
// in-memory карта goroutine-cancel-функций, поэтому после любого рестарта все
// запланированные проверки на минимум игроков, все "за 1 минуту — арендовать
// сервер" и все "стартовать ивент" безвозвратно терялись, и незавершённый
// ивент мог зависнуть навсегда. Вызывается один раз при старте сервиса
// (см. main.go).
func (s *eventService) RecoverPendingEvents(ctx context.Context) error {
	events, err := s.eventRepo.GetAllUnfinishedEvents(ctx)
	if err != nil {
		return err
	}

	for _, event := range events {
		switch event.Status {
		case domain.EventStatusInProgress:
			// Игры уже идут — своих таймеров это не требует, дальше события
			// ведутся явными вызовами StartTeamGame/FinishTeamGame.
		case domain.EventStatusConfirmed:
			// Проверка на минимум игроков уже пройдена раньше — просто
			// переставляем оставшиеся таймеры (rent_server / старт), не
			// публикуя event_confirmed ещё раз.
			s.scheduleRentAndStart(event.EventID, event.TimeStart)
		case domain.EventStatusPending:
			// Проверка ещё не проходила — ставим её заново. Если controlTime
			// уже в прошлом (сервис был недоступен дольше, чем оставалось до
			// него), sleepChan сработает немедленно.
			controlTime := event.TimeStart.Add(-CheckTimeBeforeStart)
			if err := s.controlEventTimerDenial(event.EventID, controlTime); err != nil {
				log.Printf("RecoverPendingEvents: reschedule %q: %v", event.EventID, err)
			}
		}
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

func (s *eventService) JoinUserToTeam(ctx context.Context, teamID, userID string, role domain.Role) error {
	team, err := s.eventRepo.GetTeamByID(ctx, teamID)
	if err != nil {
		return err
	}

	if team.TeamID == "" {
		return errors.New("team not found")
	}

	user, err := s.eventRepo.GetUserByID(ctx, team.EventID, userID)
	if err != nil {
		return errors.New("user not found")
	}

	res, err := s.eventRepo.IsUserInTeam(ctx, teamID, user.UserEventID)
	if err != nil {
		return err
	}

	if res == true {
		return errors.New("user already im team")
	}

	if err := s.eventRepo.JoinUserToTeam(ctx, teamID, user.UserEventID, role); err != nil {
		return err
	}

	// six_clan_members теперь считается на уровне КОМАНДЫ (team_members), а не
	// всего ивента: другой микросервис перед стартом конкретной игры смотрит
	// именно на этот флаг у участника команды и решает, звать его в игру или
	// нет (true — зовёт, false — игнорирует). hasSixClanMembers считает
	// "других" участников этого же клана в этой же команде (CheckSixClanMembers
	// исключает самого себя).
	hasSixClanMembers, err := s.eventRepo.CheckSixClanMembers(ctx, teamID, user.UserEventID, user.ClanID)
	if err != nil {
		return err
	}

	if hasSixClanMembers {
		// Флаг относится ко всей группе из одного клана в этой команде, а не
		// только к тому, кто зашёл последним и перевалил порог — иначе
		// первые (порог-1) человек так и останутся с six_clan_members=false
		// навсегда.
		if err := s.eventRepo.UpdateSixClanMembersForClanInTeam(ctx, teamID, user.ClanID, true); err != nil {
			return err
		}
	} else {
		if err := s.eventRepo.UpdateTeamMemberSixClanMembers(ctx, teamID, user.UserEventID, false); err != nil {
			return err
		}
	}

	if err := s.producer.PublishUserJoinedTeam(ctx, teamID, userID, user.UserEventID, role); err != nil {
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

	user, err := s.eventRepo.GetUserByID(ctx, team.EventID, userID)
	if err != nil {
		return errors.New("user not found")
	}

	res, err := s.eventRepo.IsUserInTeam(ctx, teamID, user.UserEventID)
	if err != nil {
		return err
	}

	if res == false {
		return errors.New("user not in team")
	}

	if err := s.eventRepo.RemoveUserFromTeam(ctx, teamID, user.UserEventID); err != nil {
		return err
	}

	// Симметрично JoinUserToTeam: после выхода из команды однокланников в ней
	// может остаться меньше порога — если так, снимаем six_clan_members
	// оставшимся, иначе флаг так и останется true уже после того, как условие
	// перестало выполняться.
	if user.ClanID != "" {
		clanMembersLeft, err := s.eventRepo.CountClanMembersInTeam(ctx, teamID, user.ClanID)
		if err != nil {
			return err
		}

		if clanMembersLeft < 6 {
			if err := s.eventRepo.UpdateSixClanMembersForClanInTeam(ctx, teamID, user.ClanID, false); err != nil {
				return err
			}
		}
	}

	if err := s.producer.PublishUserLeftTeam(ctx, teamID, userID, user.UserEventID); err != nil {
		return err
	}

	return nil
}

// StartTeamGame стартует одну игру (game_number) целиком — обе её команды
// сразу одним вызовом, поскольку они всегда начинаются вместе.
func (s *eventService) StartTeamGame(ctx context.Context, team1ID, team2ID string) error {
	if team1ID == "" || team2ID == "" {
		return errors.New("must have both team ids")
	}
	if team1ID == team2ID {
		return errors.New("team1_id and team2_id must be different")
	}

	team1, err := s.eventRepo.GetTeamByID(ctx, team1ID)
	if err != nil {
		return err
	}
	if team1.TeamID == "" {
		return errors.New("team1 not found")
	}

	team2, err := s.eventRepo.GetTeamByID(ctx, team2ID)
	if err != nil {
		return err
	}
	if team2.TeamID == "" {
		return errors.New("team2 not found")
	}

	if team1.EventID != team2.EventID {
		return errors.New("team1 and team2 belong to different events")
	}
	if team1.GameNumber != team2.GameNumber {
		return errors.New("team1 and team2 belong to different games")
	}

	if !team1.TimeStart.IsZero() || !team2.TimeStart.IsZero() {
		return errors.New("game already started")
	}

	event, err := s.eventRepo.GetEventByID(ctx, team1.EventID)
	if err != nil {
		return err
	}

	if event.EventID == "" {
		return errors.New("event not found")
	}

	if event.Status != domain.EventStatusInProgress {
		return errors.New("event hasn't started yet")
	}

	if team1.GameNumber > 1 {
		teams, err := s.eventRepo.GetTeamsByEventID(ctx, team1.EventID)
		if err != nil {
			return err
		}

		for _, t := range teams {
			if t.GameNumber == team1.GameNumber-1 && t.TimeFinish.IsZero() {
				return errors.New("previous game is not finished yet")
			}
		}
	}

	now := time.Now()
	if err := s.eventRepo.StartTeamGame(ctx, team1ID, team2ID, now); err != nil {
		return err
	}

	if err := s.producer.PublishTeamGameStarted(ctx, team1ID, team1.GameNumber, now); err != nil {
		return err
	}
	return s.producer.PublishTeamGameStarted(ctx, team2ID, team2.GameNumber, now)
}

// FinishTeamGame закрывает ровно одну игру (game_number) — сразу обе её
// команды, с явным указанием команды-победителя. Статистику игроков эта
// ручка не трогает: она заносится позже, по одному игроку за раз, через
// AddTeamMemberStats (см. её комментарий) — здесь просто нечего суммировать,
// раз статистика ещё не внесена. Поскольку обе команды закрываются одним
// вызовом, ждать "вторую сторону" отдельным флагом больше не нужно —
// game_count инкрементится сразу; когда он доходит до target_game_count
// (1, 2 или 3 — см. CreateEvent), ивент автоматически считается завершённым,
// и по итогам всех игр определяется общий победитель (это заменяет старый
// отдельный FinishEvent RPC).
func (s *eventService) FinishTeamGame(ctx context.Context, team1ID, team2ID, teamWinnerID string) error {
	if team1ID == "" || team2ID == "" {
		return errors.New("must have both team ids")
	}
	if team1ID == team2ID {
		return errors.New("team1_id and team2_id must be different")
	}
	if teamWinnerID != team1ID && teamWinnerID != team2ID {
		return errors.New("team_winner_id must be either team1_id or team2_id")
	}

	team1, err := s.eventRepo.GetTeamByID(ctx, team1ID)
	if err != nil {
		return err
	}
	if team1.TeamID == "" {
		return errors.New("team1 not found")
	}

	team2, err := s.eventRepo.GetTeamByID(ctx, team2ID)
	if err != nil {
		return err
	}
	if team2.TeamID == "" {
		return errors.New("team2 not found")
	}

	if team1.EventID != team2.EventID {
		return errors.New("team1 and team2 belong to different events")
	}
	if team1.GameNumber != team2.GameNumber {
		return errors.New("team1 and team2 belong to different games")
	}

	if team1.TimeStart.IsZero() || team2.TimeStart.IsZero() {
		return errors.New("game hasn't started yet")
	}

	if !team1.TimeFinish.IsZero() || !team2.TimeFinish.IsZero() {
		return errors.New("game already finished")
	}

	event, err := s.eventRepo.GetEventByID(ctx, team1.EventID)
	if err != nil {
		return err
	}

	if event.EventID == "" {
		return errors.New("event not found")
	}

	if event.Status != domain.EventStatusInProgress {
		return errors.New("event hasn't started yet")
	}

	now := time.Now()
	if err := s.eventRepo.FinishTeamGame(ctx, team1ID, team2ID, teamWinnerID, now); err != nil {
		return err
	}

	if err := s.producer.PublishTeamGameFinished(ctx, team1ID, team1ID == teamWinnerID, now); err != nil {
		return err
	}
	if err := s.producer.PublishTeamGameFinished(ctx, team2ID, team2ID == teamWinnerID, now); err != nil {
		return err
	}

	if err := s.eventRepo.IncrementEventGameCount(ctx, team1.EventID); err != nil {
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
	teams, err := s.eventRepo.GetTeamsByEventID(ctx, team1.EventID)
	if err != nil {
		return err
	}

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

	// При чётном target_game_count (2) счёт побед вполне может сойтись
	// 1-1 — это настоящая ничья, а не "ещё не решено", поэтому пишем "draw",
	// а не оставляем пустую строку (target_game_count 1 и 3 ничьей не дают —
	// там всегда есть большинство).
	winnerSide := "draw"
	if allyWins > enemyWins {
		winnerSide = "ally"
	} else if enemyWins > allyWins {
		winnerSide = "enemy"
	}

	if err := s.eventRepo.FinishEventDB(ctx, team1.EventID, winnerSide); err != nil {
		return err
	}

	if err := s.producer.PublishEventFinished(ctx, team1.EventID, winnerSide, now); err != nil {
		return err
	}

	return nil
}

// AddTeamMemberStats заносит статистику ОДНОГО игрока за уже сыгранную игру —
// вызывается один раз на игрока, ПОСЛЕ того как FinishTeamGame закрыл игру
// для его команды (а не до/вместо неё, как раньше). user_id — сырой внешний
// id, как и везде в API; сервис сам резолвит его в user_event_id по паре
// (user_id, team_id), клиенту знать user_event_id не нужно.
func (s *eventService) AddTeamMemberStats(ctx context.Context, teamID, userID string, kills, deaths, points, revival, destroyedVehicles int64) error {
	team, err := s.eventRepo.GetTeamByID(ctx, teamID)
	if err != nil {
		return err
	}

	if team.TeamID == "" {
		return errors.New("team not found")
	}

	if team.TimeFinish.IsZero() {
		return errors.New("game hasn't finished yet: stats can only be added after FinishTeamGame")
	}

	userEventID, err := s.eventRepo.GetUserEventIDByUserIDAndTeamID(ctx, userID, teamID)
	if err != nil {
		return errors.New("user not found")
	}

	if err := s.eventRepo.AddTeamMemberStats(ctx, teamID, userEventID, kills, deaths, points, revival, destroyedVehicles); err != nil {
		return err
	}

	// team.total_* — не то, что прислал именно этот вызов, а сумма по всем
	// team_members на данный момент, поэтому пересчитываем целиком, а не
	// прибавляем дельту (иначе повторный вызов на того же игрока с
	// исправленными числами задвоил бы итог).
	sumKills, sumDeaths, sumPoints, sumRevival, sumDestroyed, err := s.eventRepo.SumTeamMemberStats(ctx, teamID)
	if err != nil {
		return err
	}

	if err := s.eventRepo.UpdateTeamTotals(ctx, teamID, sumKills, sumDeaths, sumPoints, sumRevival, sumDestroyed); err != nil {
		return err
	}

	return s.producer.PublishTeamMemberStatsAdded(ctx, teamID, userEventID, kills, deaths, points, revival, destroyedVehicles)
}

// GetTeamStats возвращает и статистику по каждому игроку, и итоговые Total*-поля
// команды — они не приходят от клиента одним числом, а каждый раз пересчитываются
// на сервере как сумма TeamMember (см. AddTeamMemberStats), чтобы клиенту не
// пришлось самому суммировать список, если нужен только общий итог.
func (s *eventService) GetTeamStats(ctx context.Context, teamID string) (domain.Team, []domain.TeamMember, error) {
	team, err := s.eventRepo.GetTeamByID(ctx, teamID)
	if err != nil {
		return domain.Team{}, nil, err
	}

	if team.TeamID == "" {
		return domain.Team{}, nil, errors.New("team not found")
	}

	stats, err := s.eventRepo.GetTeamStats(ctx, teamID)
	if err != nil {
		return domain.Team{}, nil, err
	}

	return team, stats, nil
}
