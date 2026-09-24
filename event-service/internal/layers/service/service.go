package service

import (
	"context"
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
	// Максимум игроков в одной команде (сторона в одной игре)
	MaxTeamMembers = 50
	// Сколько максимум ждать Kafka при публикации после коммита
	publishTimeout = 30 * time.Second
)

// Конкурентность устроена так:
//   - каждая изменяющая операция над ивентом (ручка или шаг таймера) берёт
//     мьютекс этого ивента (eventLocks, см. event_lock.go) и держит его до
//     конца, включая публикацию в Kafka;
//   - вся работа с БД внутри операции — одна SERIALIZABLE-транзакция
//     (eventRepo.WithTx), первым шагом блокирующая строку ивента
//     (lockEventRow). Все проверки статуса/состояния делаются уже внутри неё;
//   - в Kafka публикуем только после коммита: транзакция может быть
//     перезапущена при конфликте сериализации, а сообщение о несостоявшемся
//     изменении отправлять нельзя.
type eventService struct {
	eventRepo   EventRepository
	producer    *kafka.Producer
	eventLocks  *eventLocker
	eventTimers map[string]*eventTimer
	timersMutex sync.Mutex
}

// eventTimer — хэндл цепочки таймеров одного ивента (контроль → аренда
// сервера → старт). По указателю finishEventTimer отличает "свою" цепочку
// от новой, поставленной поверх неё.
type eventTimer struct {
	cancel context.CancelFunc
}

func NewEventService(eventRepo EventRepository, producer *kafka.Producer) EventService {
	return &eventService{
		eventRepo:   eventRepo,
		producer:    producer,
		eventLocks:  newEventLocker(),
		eventTimers: make(map[string]*eventTimer),
	}
}

// publishCtx — контекст для публикации в Kafka после коммита: изменение уже
// в БД, поэтому отмена или дедлайн клиентского запроса не должны терять
// сообщение, но и ждать Kafka бесконечно (держа мьютекс ивента) тоже нельзя.
func publishCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), publishTimeout)
}

// lockEventRow — первый шаг любой пишущей транзакции по ивенту: блокирует
// строку events (SELECT ... FOR UPDATE) до конца транзакции и отдаёт её
// актуальное состояние.
func (s *eventService) lockEventRow(ctx context.Context, eventID string) (domain.Event, error) {
	event, err := s.eventRepo.GetEventByIDForUpdate(ctx, eventID)
	if err != nil {
		return domain.Event{}, err
	}

	if event.EventID == "" {
		return domain.Event{}, domain.NotFound("event not found")
	}

	return event, nil
}

// teamEventID — event_id команды, чтобы знать, какой мьютекс ивента брать.
// У команды он никогда не меняется, поэтому его можно прочитать до захвата
// мьютекса; остальное состояние команды перечитывается уже под ним.
func (s *eventService) teamEventID(ctx context.Context, teamID string) (string, error) {
	team, err := s.GetTeamByID(ctx, teamID)
	if err != nil {
		return "", err
	}

	return team.EventID, nil
}

func (s *eventService) CreateEvent(ctx context.Context, userCreatorID, creatorClanID, enemySideLeaderID, enemySideLeaderClanID, eventName string, timeStart time.Time, targetGameCount int64) error {
	if err := validateID("user_creator_id", userCreatorID); err != nil {
		return err
	}

	if err := validateID("creator_clan_id", creatorClanID); err != nil {
		return err
	}

	if err := validateID("enemy_side_leader_id", enemySideLeaderID); err != nil {
		return err
	}

	if err := validateID("enemy_side_leader_clan_id", enemySideLeaderClanID); err != nil {
		return err
	}

	if userCreatorID == enemySideLeaderID {
		return domain.InvalidArgument("creator and enemy side leader must be different users")
	}

	eventName, err := normalizeEventName(eventName)
	if err != nil {
		return err
	}

	now := time.Now()
	minTime := now.Add(45 * time.Minute)
	maxTime := now.Add(45 * 24 * time.Hour)

	if timeStart.Before(minTime) {
		return domain.InvalidArgument("event start time must be at least 45 minutes from now")
	}

	if timeStart.After(maxTime) {
		return domain.InvalidArgument("event start time must be within 45 days from now")
	}

	if targetGameCount < 1 || targetGameCount > 3 {
		return domain.InvalidArgument("target game count must be 1, 2 or 3")
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

	// До коммита ивент никому не виден, но мьютекс берём всё равно: иначе
	// JoinToEvent, нашедший ивент сразу после коммита, мог бы опубликовать
	// user.joined_event раньше, чем отсюда уйдёт event.created.
	unlock, err := s.eventLocks.lock(ctx, event.EventID)
	if err != nil {
		return err
	}
	defer unlock()

	// Ивент, оба сайд-лидера и команды на все игры создаются одной
	// транзакцией: при ошибке на любом шаге не остаётся "полуивента" без
	// второй стороны или без команд.
	var creatorUser, enemyUser domain.User
	err = s.eventRepo.WithTx(ctx, func(ctx context.Context) error {
		if err := s.eventRepo.CreateEvent(ctx, event); err != nil {
			return err
		}

		var err error
		creatorUser, err = s.addUserToEvent(ctx, event.EventID, userCreatorID, creatorClanID, false)
		if err != nil {
			return err
		}

		enemyUser, err = s.addUserToEvent(ctx, event.EventID, enemySideLeaderID, enemySideLeaderClanID, true)
		if err != nil {
			return err
		}

		// Team.side_leader_id ссылается на users.user_event_id (внутренний id),
		// а не на сырой внешний user_id.
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

			// Сайд-лидер обязан играть каждую игру, поэтому сразу попадает в
			// состав своей команды на все игры: убрать его оттуда нельзя
			// (RemoveUserFromTeam), выйти из ивента — тоже (LeaveEvent).
			// squad_leader — главный в команде, через SetRole эта роль не выдаётся.
			if err := s.eventRepo.JoinUserToTeam(ctx, allyTeam.TeamID, creatorUser.UserEventID, domain.RoleSquadLeader); err != nil {
				return err
			}

			if err := s.eventRepo.JoinUserToTeam(ctx, enemyTeam.TeamID, enemyUser.UserEventID, domain.RoleSquadLeader); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Таймер — только после коммита: раньше он ставился до вставки ивента и
	// оставался висеть, если создание дальше падало.
	s.controlEventTimerDenial(event.EventID, event.TimeStart)

	pubCtx, cancel := publishCtx(ctx)
	defer cancel()

	if err := s.producer.PublishEventCreated(pubCtx, event); err != nil {
		return err
	}

	if err := s.producer.PublishUserJoinedEvent(pubCtx, event.EventID, creatorUser.UserID, creatorUser.ClanID, creatorUser.Enemy, creatorUser.JoinTime); err != nil {
		return err
	}

	return s.producer.PublishUserJoinedEvent(pubCtx, event.EventID, enemyUser.UserID, enemyUser.ClanID, enemyUser.Enemy, enemyUser.JoinTime)
}

func (s *eventService) GetEventsByCreatorId(ctx context.Context, userCreateID string) ([]*domain.Event, error) {
	if err := validateID("user_create_id", userCreateID); err != nil {
		return nil, err
	}

	events, err := s.eventRepo.GetEventsByCreatorId(ctx, userCreateID)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (s *eventService) GetLastEventByCreatorId(ctx context.Context, userCreateID string) (*domain.Event, error) {
	if err := validateID("user_create_id", userCreateID); err != nil {
		return nil, err
	}

	event, err := s.eventRepo.GetLastEventByCreatorId(ctx, userCreateID)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *eventService) GetEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error) {
	if err := validateSearchName(eventName); err != nil {
		return []domain.Event{}, err
	}

	event, err := s.eventRepo.GetEventsByEventName(ctx, eventName)
	if err != nil {
		return []domain.Event{}, err
	}

	return event, nil
}

func (s *eventService) GetEventMembersList(ctx context.Context, eventID string) ([]domain.User, error) {
	if err := validateID("event_id", eventID); err != nil {
		return []domain.User{}, err
	}

	events, err := s.eventRepo.GetEventMembersList(ctx, eventID)
	if err != nil {
		return []domain.User{}, err
	}

	return events, nil
}

func (s *eventService) UpdateTimeEvent(ctx context.Context, eventID, userCreateID string, newTimeStart time.Time) error {
	if err := validateID("event_id", eventID); err != nil {
		return err
	}

	if err := validateID("user_create_id", userCreateID); err != nil {
		return err
	}

	unlock, err := s.eventLocks.lock(ctx, eventID)
	if err != nil {
		return err
	}
	defer unlock()

	err = s.eventRepo.WithTx(ctx, func(ctx context.Context) error {
		event, err := s.lockEventRow(ctx, eventID)
		if err != nil {
			return err
		}

		if event.UserCreateID != userCreateID {
			return domain.PermissionDenied("you are not admin")
		}

		// Время можно менять только в статусе pending
		if event.Status != domain.EventStatusPending {
			return domain.FailedPrecondition("cannot change time: event status is %s (must be pending)", event.Status)
		}

		now := time.Now()
		minTime := now.Add(45 * time.Minute)
		maxTime := now.Add(45 * 24 * time.Hour)

		if newTimeStart.Before(minTime) {
			return domain.InvalidArgument("event start time must be at least 45 minutes from now")
		}

		if newTimeStart.After(maxTime) {
			return domain.InvalidArgument("event start time must be within 45 days from now")
		}

		return s.eventRepo.UpdateTimeEvent(ctx, eventID, newTimeStart)
	})
	if err != nil {
		return err
	}

	// Перезапуск таймера — под тем же мьютексом, что и смена времени, и до
	// публикации: старая цепочка отменяется раньше, чем её шаг контроля
	// успеет взять мьютекс, и не подтвердит/отклонит ивент по старому
	// времени. Раньше при ошибке публикации таймер не переставлялся вовсе.
	s.controlEventTimerDenial(eventID, newTimeStart)

	pubCtx, cancel := publishCtx(ctx)
	defer cancel()

	return s.producer.PublishEventTimeUpdated(pubCtx, eventID, userCreateID, newTimeStart)
}

// CancelEvent — ручная отмена ивента создателем: доступна, пока ивент ещё
// pending или confirmed. Статус становится "canceled" (в отличие от
// "declined", в который автоматически уходит ивент, не набравший минимум
// игроков к контрольной точке — см. checkMinPlayers). Сама строка события и
// вся история участников/команд никуда не удаляются — просто больше не активны.
func (s *eventService) CancelEvent(ctx context.Context, eventID, userCreateID string) error {
	if err := validateID("event_id", eventID); err != nil {
		return err
	}

	if err := validateID("user_create_id", userCreateID); err != nil {
		return err
	}

	unlock, err := s.eventLocks.lock(ctx, eventID)
	if err != nil {
		return err
	}
	defer unlock()

	err = s.eventRepo.WithTx(ctx, func(ctx context.Context) error {
		event, err := s.lockEventRow(ctx, eventID)
		if err != nil {
			return err
		}

		if event.UserCreateID != userCreateID {
			return domain.PermissionDenied("you are not admin")
		}

		if event.Status != domain.EventStatusPending && event.Status != domain.EventStatusConfirmed {
			return domain.FailedPrecondition("cannot cancel event: current status is %q", event.Status)
		}

		return s.eventRepo.UpdateEventStatus(ctx, eventID, domain.EventStatusCanceled)
	})
	if err != nil {
		return err
	}

	// Останавливаем всю цепочку таймеров ивента: и контроль, и (если ивент
	// уже confirmed) аренду сервера со стартом. Раньше последние две
	// горутины не отменялись, и отменённый ивент в момент старта всё равно
	// получал rent.server, event.started и запущенную первую игру.
	s.cancelEventTimer(eventID)

	pubCtx, cancel := publishCtx(ctx)
	defer cancel()

	return s.producer.PublishEventCanceled(pubCtx, eventID, userCreateID)
}

func (s *eventService) JoinToEvent(ctx context.Context, eventID, userID, clanID string, enemy bool) error {
	if err := validateID("event_id", eventID); err != nil {
		return err
	}

	if err := validateID("user_id", userID); err != nil {
		return err
	}

	if err := validateID("clan_id", clanID); err != nil {
		return err
	}

	unlock, err := s.eventLocks.lock(ctx, eventID)
	if err != nil {
		return err
	}
	defer unlock()

	var user domain.User
	err = s.eventRepo.WithTx(ctx, func(ctx context.Context) error {
		event, err := s.lockEventRow(ctx, eventID)
		if err != nil {
			return err
		}

		// Присоединиться можно только к pending-ивенту: после подтверждения
		// состав заморожен — иначе он мог бы упасть ниже минимума уже после
		// проверки, а список игроков в rent.server устареть.
		if event.Status != domain.EventStatusPending {
			return domain.FailedPrecondition("cannot join: event status is %s (must be pending)", event.Status)
		}

		user, err = s.addUserToEvent(ctx, eventID, userID, clanID, enemy)
		return err
	})
	if err != nil {
		return err
	}

	pubCtx, cancel := publishCtx(ctx)
	defer cancel()

	return s.producer.PublishUserJoinedEvent(pubCtx, eventID, userID, clanID, enemy, user.JoinTime)
}

// addUserToEvent — общая часть CreateEvent и JoinToEvent. Вызывается внутри
// транзакции, в которой строка ивента уже заблокирована (или только что
// вставлена этой же транзакцией). Роль здесь не назначается: она есть только
// внутри команды (team_members), см. SetRole.
func (s *eventService) addUserToEvent(ctx context.Context, eventID, userID, clanID string, enemy bool) (domain.User, error) {
	alreadyJoined, err := s.eventRepo.IsUserInEvent(ctx, eventID, userID)
	if err != nil {
		return domain.User{}, err
	}

	if alreadyJoined {
		return domain.User{}, domain.AlreadyExists("user already joined this event")
	}

	user := domain.User{
		UserEventID: uuid.New().String(),
		UserID:      userID,
		EventID:     eventID,
		ClanID:      clanID,
		Enemy:       enemy,
		JoinTime:    time.Now(),
	}

	if err := s.eventRepo.JoinToEvent(ctx, user.UserEventID, eventID, userID, clanID, enemy, user.JoinTime); err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (s *eventService) LeaveEvent(ctx context.Context, userID, eventID string) error {
	if err := validateID("event_id", eventID); err != nil {
		return err
	}

	if err := validateID("user_id", userID); err != nil {
		return err
	}

	unlock, err := s.eventLocks.lock(ctx, eventID)
	if err != nil {
		return err
	}
	defer unlock()

	err = s.eventRepo.WithTx(ctx, func(ctx context.Context) error {
		event, err := s.lockEventRow(ctx, eventID)
		if err != nil {
			return err
		}

		// Сайд-лидеры обязаны играть каждую игру, поэтому выйти из ивента не
		// могут ни создатель, ни лидер второй стороны: создателю остаётся
		// CancelEvent, второму — договариваться с ним.
		if event.UserCreateID == userID {
			return domain.PermissionDenied("side leader cannot leave event, use CancelEvent instead")
		}

		if event.EnemySideLeader == userID {
			return domain.PermissionDenied("side leader cannot leave event")
		}

		// Выходить можно только из pending-ивента: после подтверждения состав
		// заморожен (симметрично JoinToEvent).
		if event.Status != domain.EventStatusPending {
			return domain.FailedPrecondition("cannot leave: event status is %s (must be pending)", event.Status)
		}

		// Проверяем, что уходящий вообще состоит в этом ивенте — иначе просто
		// молча не удалится ни одной строки, а ошибки не будет.
		user, err := s.eventRepo.GetUserByID(ctx, eventID, userID)
		if err != nil {
			return err
		}

		// Сначала из всех команд ивента, иначе строку users не удалить — на
		// неё ссылается team_members. Игры ещё не начинались (ивент pending),
		// так что терять нечего.
		teamIDs, err := s.eventRepo.RemoveUserFromAllTeams(ctx, eventID, user.UserEventID)
		if err != nil {
			return err
		}

		// Там, где однокланников осталось меньше порога, снимаем флаг —
		// как в RemoveUserFromTeam.
		for _, teamID := range teamIDs {
			if err := s.refreshSixClanMembers(ctx, teamID, user.ClanID); err != nil {
				return err
			}
		}

		return s.eventRepo.LeaveEvent(ctx, userID, eventID)
	})
	if err != nil {
		return err
	}

	pubCtx, cancel := publishCtx(ctx)
	defer cancel()

	return s.producer.PublishUserLeftEvent(pubCtx, eventID, userID)
}

// SetRole назначает роль внутри ОДНОЙ команды, то есть на одну игру: роли
// уровня всего ивента больше нет (колонка users.role убрана). Менять роли
// может только сайд-лидер этой же команды — до чужой стороны он не
// дотягивается, — и только пока игра не началась.
func (s *eventService) SetRole(ctx context.Context, teamID, yourID, userID string, role domain.Role) error {
	if err := validateID("team_id", teamID); err != nil {
		return err
	}

	if err := validateID("side_leader_id", yourID); err != nil {
		return err
	}

	if err := validateID("user_id", userID); err != nil {
		return err
	}

	if err := validateAssignableRole(role); err != nil {
		return err
	}

	eventID, err := s.teamEventID(ctx, teamID)
	if err != nil {
		return err
	}

	unlock, err := s.eventLocks.lock(ctx, eventID)
	if err != nil {
		return err
	}
	defer unlock()

	err = s.eventRepo.WithTx(ctx, func(ctx context.Context) error {
		if _, err := s.lockEventRow(ctx, eventID); err != nil {
			return err
		}

		team, err := s.eventRepo.GetTeamByID(ctx, teamID)
		if err != nil {
			return err
		}

		if team.Status != domain.TeamStatusPending {
			return domain.FailedPrecondition("cannot set role: game already started (team status is %s)", team.Status)
		}

		caller, err := s.eventRepo.GetUserByID(ctx, eventID, yourID)
		if err != nil {
			return err
		}

		// Сайд-лидер команды — единственный, кто раздаёт в ней роли; чужую
		// сторону это автоматически закрывает.
		if team.SideLeaderID != caller.UserEventID {
			return domain.PermissionDenied("only the side leader of this team can set roles")
		}

		user, err := s.eventRepo.GetUserByID(ctx, eventID, userID)
		if err != nil {
			return err
		}

		// Сам сайд-лидер остаётся squad_leader — понизить его (в том числе
		// самому себе) нельзя.
		if team.SideLeaderID == user.UserEventID {
			return domain.FailedPrecondition("cannot change the role of the team side leader")
		}

		return s.eventRepo.UpdateTeamMemberRole(ctx, teamID, user.UserEventID, role)
	})
	if err != nil {
		return err
	}

	pubCtx, cancel := publishCtx(ctx)
	defer cancel()

	return s.producer.PublishUserRoleChanged(pubCtx, eventID, teamID, userID, role)
}

// ── Таймеры ─────────────────────────────────────────────────────────────────
//
// На каждый активный ивент — одна цепочка таймеров в одной горутине:
// контроль минимума игроков → аренда сервера → старт ивента и первой игры.
// Цепочка зарегистрирована в eventTimers и отменяется целиком
// (CancelEvent) или заменяется новой (UpdateTimeEvent). Каждый её шаг
// выполняется под мьютексом ивента, как и ручки, и перед работой проверяет,
// что цепочку не отменили, пока он ждал мьютекс.

// startEventTimer регистрирует новую цепочку таймеров ивента, отменяя
// предыдущую, если она была.
func (s *eventService) startEventTimer(eventID string) (context.Context, *eventTimer) {
	s.timersMutex.Lock()
	defer s.timersMutex.Unlock()

	if old, exists := s.eventTimers[eventID]; exists {
		old.cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	t := &eventTimer{cancel: cancel}
	s.eventTimers[eventID] = t

	return ctx, t
}

// finishEventTimer вызывается цепочкой, когда она отработала или была
// отменена. Удаляет запись из мапы, только если там всё ещё эта же
// цепочка — иначе стёр бы новую, которую успел поставить UpdateTimeEvent.
func (s *eventService) finishEventTimer(eventID string, t *eventTimer) {
	s.timersMutex.Lock()
	defer s.timersMutex.Unlock()

	if current, exists := s.eventTimers[eventID]; exists && current == t {
		delete(s.eventTimers, eventID)
	}
	t.cancel()
}

// cancelEventTimer отменяет цепочку таймеров ивента, если она есть, и
// убирает запись из мапы. Безопасно вызывать, даже если цепочки нет.
func (s *eventService) cancelEventTimer(eventID string) {
	s.timersMutex.Lock()
	defer s.timersMutex.Unlock()

	if t, exists := s.eventTimers[eventID]; exists {
		t.cancel()
		delete(s.eventTimers, eventID)
	}
}

// lockForTimerStep захватывает мьютекс ивента для шага цепочки таймеров.
// false — цепочку отменили (CancelEvent/UpdateTimeEvent), пока шаг ждал
// мьютекс: выполнять его уже нельзя.
func (s *eventService) lockForTimerStep(ctx context.Context, eventID string) (func(), bool) {
	unlock, err := s.eventLocks.lock(ctx, eventID)
	if err != nil {
		return nil, false
	}

	// select внутри lock мог выбрать освободившийся мьютекс, даже если ctx
	// к этому моменту уже отменён.
	if ctx.Err() != nil {
		unlock()
		return nil, false
	}

	return unlock, true
}

// waitUntil ждёт момента t или отмены ctx; true — дождались, и цепочка не
// отменена. Момент в прошлом срабатывает сразу, а не теряется: после
// рестарта сервиса (RecoverPendingEvents) controlTime/rentServerTime/
// TimeStart вполне могут оказаться уже позади.
func waitUntil(ctx context.Context, t time.Time) bool {
	if d := time.Until(t); d > 0 {
		timer := time.NewTimer(d)
		defer timer.Stop()

		select {
		case <-timer.C:
		case <-ctx.Done():
			return false
		}
	}

	return ctx.Err() == nil
}

// controlEventTimerDenial ставит цепочку таймеров ивента — единственный
// источник правды по времени старта: за CheckTimeBeforeStart до timeStart
// проверяет, набралось ли MinPlayersRequired, и либо подтверждает ивент (и
// дальше та же цепочка ждёт аренду сервера и старт), либо отклоняет его.
// Повторный вызов для того же eventID (UpdateTimeEvent) отменяет предыдущую
// цепочку и ставит новую на её место.
func (s *eventService) controlEventTimerDenial(eventID string, timeStart time.Time) {
	ctx, t := s.startEventTimer(eventID)

	go func() {
		defer s.finishEventTimer(eventID, t)

		if !waitUntil(ctx, timeStart.Add(-CheckTimeBeforeStart)) {
			return
		}

		confirmed, err := s.checkMinPlayers(ctx, eventID)
		if err != nil {
			log.Printf("checkMinPlayers(%q): %v", eventID, err)
		}

		if confirmed {
			s.runRentAndStart(ctx, eventID, timeStart)
		}
	}()
}

// scheduleRentAndStart ставит только вторую половину цепочки (аренда
// сервера и старт) — для RecoverPendingEvents, чтобы после рестарта
// процесса не публиковать event.confirmed повторно для уже подтверждённого
// ивента.
func (s *eventService) scheduleRentAndStart(eventID string, timeStart time.Time) {
	ctx, t := s.startEventTimer(eventID)

	go func() {
		defer s.finishEventTimer(eventID, t)

		s.runRentAndStart(ctx, eventID, timeStart)
	}()
}

// checkMinPlayers — контрольная точка: pending-ивент, набравший
// MinPlayersRequired, становится confirmed, иначе — declined (не удаляется:
// история участников/команд остаётся, просто ивент больше не активен).
// Возвращает true, если ивент подтверждён и цепочке нужно идти дальше, к
// аренде сервера и старту — даже если публикация в Kafka не удалась.
func (s *eventService) checkMinPlayers(ctx context.Context, eventID string) (bool, error) {
	unlock, ok := s.lockForTimerStep(ctx, eventID)
	if !ok {
		return false, nil
	}
	defer unlock()

	var newStatus domain.EventStatus
	err := s.eventRepo.WithTx(ctx, func(ctx context.Context) error {
		newStatus = ""

		event, err := s.lockEventRow(ctx, eventID)
		if err != nil {
			return err
		}

		// Ивент уже не pending — например, создатель отменил его вручную
		// (CancelEvent) до срабатывания таймера.
		if event.Status != domain.EventStatusPending {
			return nil
		}

		newStatus = domain.EventStatusDeclined
		if event.UserCount >= MinPlayersRequired {
			newStatus = domain.EventStatusConfirmed
		}

		return s.eventRepo.UpdateEventStatus(ctx, eventID, newStatus)
	})
	if err != nil {
		return false, err
	}

	pubCtx, cancel := publishCtx(ctx)
	defer cancel()

	switch newStatus {
	case domain.EventStatusConfirmed:
		return true, s.producer.PublishEventConfirmed(pubCtx, eventID)
	case domain.EventStatusDeclined:
		return false, s.producer.PublishEventDeclined(pubCtx, eventID)
	}

	return false, nil
}

// runRentAndStart — вторая половина цепочки подтверждённого ивента: за
// RentServerTimeBeforeStart до старта — сигнал "пора арендовать сервер", в
// момент старта — запуск ивента и первой игры.
func (s *eventService) runRentAndStart(ctx context.Context, eventID string, timeStart time.Time) {
	if !waitUntil(ctx, timeStart.Add(-RentServerTimeBeforeStart)) {
		return
	}

	if err := s.rentServer(ctx, eventID, timeStart); err != nil {
		log.Printf("rentServer(%q): %v", eventID, err)
	}

	if !waitUntil(ctx, timeStart) {
		return
	}

	if err := s.startEventAndGames(ctx, eventID); err != nil {
		log.Printf("startEventAndGames(%q): %v", eventID, err)
	}
}

// rentServer публикует rent.server со списком игроков, если ивент всё ещё
// confirmed (его могли отменить в последний момент).
func (s *eventService) rentServer(ctx context.Context, eventID string, timeStart time.Time) error {
	unlock, ok := s.lockForTimerStep(ctx, eventID)
	if !ok {
		return nil
	}
	defer unlock()

	var needRent bool
	var playersList []string
	// Статус, флаг отправки и список игроков — из одного снимка БД.
	err := s.eventRepo.WithTx(ctx, func(ctx context.Context) error {
		event, err := s.eventRepo.GetEventByID(ctx, eventID)
		if err != nil {
			return err
		}

		needRent = event.Status == domain.EventStatusConfirmed && !event.RentServerSent
		if !needRent {
			return nil
		}

		playersList, err = s.eventRepo.GetUserIDsByEventID(ctx, eventID)
		return err
	})
	if err != nil || !needRent {
		return err
	}

	pubCtx, cancel := publishCtx(ctx)
	defer cancel()

	if err := s.producer.PublishRentServer(pubCtx, eventID, playersList, timeStart); err != nil {
		return err
	}

	// Флаг ставим только после успешной публикации: если Kafka недоступна, он
	// останется снятым, и после рестарта сервиса аренда уйдёт заново, а не
	// потеряется совсем.
	return s.eventRepo.MarkRentServerSent(pubCtx, eventID)
}

// startEventAndGames запускается только из цепочки таймеров (в момент
// event.TimeStart), а не по ручке — публичного StartEvent RPC больше нет.
// Меняет статус на in_progress и стартует первую по счёту игру
// (game_number = 1) сразу для обеих сторон; остальные игры стартуют по мере
// того, как заканчиваются предыдущие (через публичный StartTeamGame).
func (s *eventService) startEventAndGames(ctx context.Context, eventID string) error {
	unlock, ok := s.lockForTimerStep(ctx, eventID)
	if !ok {
		return nil
	}
	defer unlock()

	var event domain.Event
	var team1ID, team2ID string
	var now time.Time
	// Статус in_progress и старт первой игры — одной транзакцией: раньше при
	// ошибке между ними ивент оставался in_progress без запущенной игры.
	err := s.eventRepo.WithTx(ctx, func(ctx context.Context) error {
		var err error
		event, err = s.lockEventRow(ctx, eventID)
		if err != nil {
			return err
		}

		if event.Status != domain.EventStatusConfirmed {
			return domain.FailedPrecondition("cannot start event: status is %q (must be confirmed)", event.Status)
		}

		if err := s.eventRepo.UpdateEventStatus(ctx, eventID, domain.EventStatusInProgress); err != nil {
			return err
		}

		teams, err := s.eventRepo.GetTeamsByEventID(ctx, eventID)
		if err != nil {
			return err
		}

		team1ID, team2ID = "", ""
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

		now = time.Now()
		return s.eventRepo.StartTeamGame(ctx, team1ID, team2ID, now)
	})
	if err != nil {
		return err
	}

	pubCtx, cancel := publishCtx(ctx)
	defer cancel()

	if err := s.producer.PublishEventStarted(pubCtx, eventID, event.TimeStart); err != nil {
		return err
	}

	if err := s.producer.PublishTeamGameStarted(pubCtx, team1ID, 1, now); err != nil {
		return err
	}

	return s.producer.PublishTeamGameStarted(pubCtx, team2ID, 1, now)
}

// RecoverPendingEvents переживает рестарт процесса: eventTimers — чисто
// in-memory карта цепочек таймеров, поэтому после любого рестарта все
// запланированные проверки на минимум игроков, аренды сервера и старты
// терялись, и незавершённый ивент мог зависнуть навсегда. Вызывается один
// раз при старте сервиса, до того как gRPC-сервер начнёт принимать запросы
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
			// него), waitUntil сработает немедленно.
			s.controlEventTimerDenial(event.EventID, event.TimeStart)
		}
	}

	return nil
}

func (s *eventService) GetUnfinishedEventsByUserID(ctx context.Context, userCreateID string) ([]*domain.Event, error) {
	if err := validateID("user_create_id", userCreateID); err != nil {
		return nil, err
	}

	events, err := s.eventRepo.GetUnfinishedEventsByUserID(ctx, userCreateID)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (s *eventService) GetUnfinishedEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error) {
	if err := validateSearchName(eventName); err != nil {
		return []domain.Event{}, err
	}

	events, err := s.eventRepo.GetUnfinishedEventsByEventName(ctx, eventName)
	if err != nil {
		return []domain.Event{}, err
	}

	return events, nil
}

// FinishEvent как отдельный метод убран: ивент теперь завершается автоматически
// внутри FinishTeamGame, когда game_count после этой игры достигает target_game_count.

func (s *eventService) GetTeamsByEventID(ctx context.Context, eventID string) ([]domain.Team, error) {
	if err := validateID("event_id", eventID); err != nil {
		return nil, err
	}

	return s.eventRepo.GetTeamsByEventID(ctx, eventID)
}

func (s *eventService) GetTeamByID(ctx context.Context, teamID string) (domain.Team, error) {
	if err := validateID("team_id", teamID); err != nil {
		return domain.Team{}, err
	}

	team, err := s.eventRepo.GetTeamByID(ctx, teamID)
	if err != nil {
		return domain.Team{}, err
	}

	if team.TeamID == "" {
		return domain.Team{}, domain.NotFound("team not found")
	}

	return team, nil
}

func (s *eventService) JoinUserToTeam(ctx context.Context, teamID, userID string, role domain.Role) error {
	if err := validateID("team_id", teamID); err != nil {
		return err
	}

	if err := validateID("user_id", userID); err != nil {
		return err
	}

	// squad_leader в команде один — сайд-лидер стороны, его ставит CreateEvent.
	if err := validateAssignableRole(role); err != nil {
		return err
	}

	eventID, err := s.teamEventID(ctx, teamID)
	if err != nil {
		return err
	}

	unlock, err := s.eventLocks.lock(ctx, eventID)
	if err != nil {
		return err
	}
	defer unlock()

	var user domain.User
	err = s.eventRepo.WithTx(ctx, func(ctx context.Context) error {
		event, err := s.lockEventRow(ctx, eventID)
		if err != nil {
			return err
		}

		if event.Status == domain.EventStatusFinished || event.Status == domain.EventStatusCanceled || event.Status == domain.EventStatusDeclined {
			return domain.FailedPrecondition("cannot join team: event status is %s", event.Status)
		}

		team, err := s.eventRepo.GetTeamByID(ctx, teamID)
		if err != nil {
			return err
		}

		// Состав набирается до игры: после старта команда зафиксирована.
		if team.Status != domain.TeamStatusPending {
			return domain.FailedPrecondition("cannot join team: game already started (team status is %s)", team.Status)
		}

		if team.MembersCount >= MaxTeamMembers {
			return domain.FailedPrecondition("team is full: %d players max", MaxTeamMembers)
		}

		user, err = s.eventRepo.GetUserByID(ctx, eventID, userID)
		if err != nil {
			return err
		}

		// Играть можно только за ту сторону, за которую человек вошёл в
		// ивент: сторона команды — это сторона её сайд-лидера.
		sideLeader, err := s.eventRepo.GetUserByUserEventID(ctx, team.SideLeaderID)
		if err != nil {
			return err
		}

		if user.Enemy != sideLeader.Enemy {
			return domain.FailedPrecondition("user cannot join a team of the opposite side")
		}

		res, err := s.eventRepo.IsUserInTeam(ctx, teamID, user.UserEventID)
		if err != nil {
			return err
		}

		if res == true {
			return domain.AlreadyExists("user is already in this team")
		}

		if err := s.eventRepo.JoinUserToTeam(ctx, teamID, user.UserEventID, role); err != nil {
			return err
		}

		// six_clan_members считается на уровне КОМАНДЫ (team_members), а не
		// всего ивента: другой микросервис перед стартом конкретной игры
		// смотрит именно на этот флаг у участника команды и решает, звать его
		// в игру или нет (true — зовёт, false — игнорирует). hasSixClanMembers
		// считает "других" участников этого же клана в этой же команде
		// (CheckSixClanMembers исключает самого себя). Подсчёт и обновление
		// флагов — в той же транзакции, что и вставка: иначе два
		// одновременных входа однокланников не видели бы друг друга, и порог
		// "проскакивал" бы.
		hasSixClanMembers, err := s.eventRepo.CheckSixClanMembers(ctx, teamID, user.UserEventID, user.ClanID)
		if err != nil {
			return err
		}

		if hasSixClanMembers {
			// Флаг относится ко всей группе из одного клана в этой команде, а не
			// только к тому, кто зашёл последним и перевалил порог — иначе
			// первые (порог-1) человек так и останутся с six_clan_members=false
			// навсегда.
			return s.eventRepo.UpdateSixClanMembersForClanInTeam(ctx, teamID, user.ClanID, true)
		}

		return s.eventRepo.UpdateTeamMemberSixClanMembers(ctx, teamID, user.UserEventID, false)
	})
	if err != nil {
		return err
	}

	pubCtx, cancel := publishCtx(ctx)
	defer cancel()

	return s.producer.PublishUserJoinedTeam(pubCtx, teamID, userID, user.UserEventID, role)
}

func (s *eventService) RemoveUserFromTeam(ctx context.Context, teamID, userID string) error {
	if err := validateID("team_id", teamID); err != nil {
		return err
	}

	if err := validateID("user_id", userID); err != nil {
		return err
	}

	eventID, err := s.teamEventID(ctx, teamID)
	if err != nil {
		return err
	}

	unlock, err := s.eventLocks.lock(ctx, eventID)
	if err != nil {
		return err
	}
	defer unlock()

	var user domain.User
	err = s.eventRepo.WithTx(ctx, func(ctx context.Context) error {
		if _, err := s.lockEventRow(ctx, eventID); err != nil {
			return err
		}

		team, err := s.eventRepo.GetTeamByID(ctx, teamID)
		if err != nil {
			return err
		}

		// Убирать из состава можно только до старта игры: потом у игрока уже
		// есть участие в ней и (после её конца) статистика.
		if team.Status != domain.TeamStatusPending {
			return domain.FailedPrecondition("cannot remove from team: game already started (team status is %s)", team.Status)
		}

		user, err = s.eventRepo.GetUserByID(ctx, eventID, userID)
		if err != nil {
			return err
		}

		// Сайд-лидер обязан играть каждую игру.
		if team.SideLeaderID == user.UserEventID {
			return domain.FailedPrecondition("side leader cannot be removed from team")
		}

		res, err := s.eventRepo.IsUserInTeam(ctx, teamID, user.UserEventID)
		if err != nil {
			return err
		}

		if res == false {
			return domain.NotFound("user is not in this team")
		}

		if err := s.eventRepo.RemoveUserFromTeam(ctx, teamID, user.UserEventID); err != nil {
			return err
		}

		return s.refreshSixClanMembers(ctx, teamID, user.ClanID)
	})
	if err != nil {
		return err
	}

	pubCtx, cancel := publishCtx(ctx)
	defer cancel()

	return s.producer.PublishUserLeftTeam(pubCtx, teamID, userID, user.UserEventID)
}

// refreshSixClanMembers — симметрично JoinUserToTeam: после ухода игрока
// однокланников в команде может остаться меньше порога, и тогда флаг
// six_clan_members надо снять всей группе, иначе он останется true уже после
// того, как условие перестало выполняться. Вызывается внутри транзакции.
func (s *eventService) refreshSixClanMembers(ctx context.Context, teamID, clanID string) error {
	if clanID == "" {
		return nil
	}

	clanMembersLeft, err := s.eventRepo.CountClanMembersInTeam(ctx, teamID, clanID)
	if err != nil {
		return err
	}

	if clanMembersLeft >= 6 {
		return nil
	}

	return s.eventRepo.UpdateSixClanMembersForClanInTeam(ctx, teamID, clanID, false)
}

// StartTeamGame стартует одну игру (game_number) целиком — обе её команды
// сразу одним вызовом, поскольку они всегда начинаются вместе.
func (s *eventService) StartTeamGame(ctx context.Context, team1ID, team2ID string) error {
	if err := validateID("team1_id", team1ID); err != nil {
		return err
	}

	if err := validateID("team2_id", team2ID); err != nil {
		return err
	}

	if team1ID == team2ID {
		return domain.InvalidArgument("team1_id and team2_id must be different")
	}

	eventID, err := s.teamEventID(ctx, team1ID)
	if err != nil {
		return err
	}

	unlock, err := s.eventLocks.lock(ctx, eventID)
	if err != nil {
		return err
	}
	defer unlock()

	var gameNumber int64
	var now time.Time
	err = s.eventRepo.WithTx(ctx, func(ctx context.Context) error {
		event, err := s.lockEventRow(ctx, eventID)
		if err != nil {
			return err
		}

		// Команды читаем уже под блокировкой ивента: прочитанное до неё могло
		// устареть (игру мог только что стартовать параллельный запрос).
		team1, err := s.eventRepo.GetTeamByID(ctx, team1ID)
		if err != nil {
			return err
		}
		if team1.TeamID == "" {
			return domain.NotFound("team1 not found")
		}

		team2, err := s.eventRepo.GetTeamByID(ctx, team2ID)
		if err != nil {
			return err
		}
		if team2.TeamID == "" {
			return domain.NotFound("team2 not found")
		}

		if team1.EventID != team2.EventID {
			return domain.InvalidArgument("team1 and team2 belong to different events")
		}
		if team1.GameNumber != team2.GameNumber {
			return domain.InvalidArgument("team1 and team2 belong to different games")
		}

		if team1.Status != domain.TeamStatusPending || team2.Status != domain.TeamStatusPending {
			return domain.FailedPrecondition("game already started (team status is %s)", team1.Status)
		}

		if event.Status != domain.EventStatusInProgress {
			return domain.FailedPrecondition("cannot start game: event status is %s (must be in_progress)", event.Status)
		}

		if team1.GameNumber > 1 {
			teams, err := s.eventRepo.GetTeamsByEventID(ctx, eventID)
			if err != nil {
				return err
			}

			for _, t := range teams {
				if t.GameNumber == team1.GameNumber-1 && t.Status != domain.TeamStatusFinished {
					return domain.FailedPrecondition("previous game is not finished yet")
				}
			}
		}

		gameNumber = team1.GameNumber
		now = time.Now()
		return s.eventRepo.StartTeamGame(ctx, team1ID, team2ID, now)
	})
	if err != nil {
		return err
	}

	pubCtx, cancel := publishCtx(ctx)
	defer cancel()

	if err := s.producer.PublishTeamGameStarted(pubCtx, team1ID, gameNumber, now); err != nil {
		return err
	}

	return s.producer.PublishTeamGameStarted(pubCtx, team2ID, gameNumber, now)
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
// отдельный FinishEvent RPC). Закрытие игры, инкремент game_count и
// завершение ивента — одна транзакция: два одновременных FinishTeamGame
// раньше оба проходили проверку "ещё не закончена" и дважды увеличивали
// game_count.
func (s *eventService) FinishTeamGame(ctx context.Context, team1ID, team2ID, teamWinnerID string) error {
	if err := validateID("team1_id", team1ID); err != nil {
		return err
	}

	if err := validateID("team2_id", team2ID); err != nil {
		return err
	}

	if err := validateID("team_winner_id", teamWinnerID); err != nil {
		return err
	}

	if team1ID == team2ID {
		return domain.InvalidArgument("team1_id and team2_id must be different")
	}

	if teamWinnerID != team1ID && teamWinnerID != team2ID {
		return domain.InvalidArgument("team_winner_id must be either team1_id or team2_id")
	}

	eventID, err := s.teamEventID(ctx, team1ID)
	if err != nil {
		return err
	}

	unlock, err := s.eventLocks.lock(ctx, eventID)
	if err != nil {
		return err
	}
	defer unlock()

	var now time.Time
	var eventFinished bool
	var winnerSide string
	err = s.eventRepo.WithTx(ctx, func(ctx context.Context) error {
		eventFinished, winnerSide = false, ""

		event, err := s.lockEventRow(ctx, eventID)
		if err != nil {
			return err
		}

		team1, err := s.eventRepo.GetTeamByID(ctx, team1ID)
		if err != nil {
			return err
		}
		if team1.TeamID == "" {
			return domain.NotFound("team1 not found")
		}

		team2, err := s.eventRepo.GetTeamByID(ctx, team2ID)
		if err != nil {
			return err
		}
		if team2.TeamID == "" {
			return domain.NotFound("team2 not found")
		}

		if team1.EventID != team2.EventID {
			return domain.InvalidArgument("team1 and team2 belong to different events")
		}
		if team1.GameNumber != team2.GameNumber {
			return domain.InvalidArgument("team1 and team2 belong to different games")
		}

		if team1.Status == domain.TeamStatusFinished || team2.Status == domain.TeamStatusFinished {
			return domain.FailedPrecondition("game already finished")
		}

		if team1.Status != domain.TeamStatusInProgress || team2.Status != domain.TeamStatusInProgress {
			return domain.FailedPrecondition("game hasn't started yet")
		}

		if event.Status != domain.EventStatusInProgress {
			return domain.FailedPrecondition("cannot finish game: event status is %s (must be in_progress)", event.Status)
		}

		now = time.Now()
		if err := s.eventRepo.FinishTeamGame(ctx, team1ID, team2ID, teamWinnerID, now); err != nil {
			return err
		}

		newGameCount, err := s.eventRepo.IncrementEventGameCount(ctx, eventID)
		if err != nil {
			return err
		}

		// Считаем счёт по всем сыгранным играм. team.side_leader_id — это
		// users.user_event_id, а не сырой event.user_create_id/
		// enemy_side_leader_id, поэтому сначала достаём internal id обоих
		// сайд-лидеров, чтобы было с чем сравнивать team.Winner.
		teams, err := s.eventRepo.GetTeamsByEventID(ctx, eventID)
		if err != nil {
			return err
		}

		creatorUser, err := s.eventRepo.GetUserByID(ctx, eventID, event.UserCreateID)
		if err != nil {
			return err
		}

		enemyUser, err := s.eventRepo.GetUserByID(ctx, eventID, event.EnemySideLeader)
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

		// Серия идёт до большинства побед: при 2:0 из трёх игр третью играть
		// уже незачем — ивент завершается досрочно, оставшаяся игра не
		// стартует (статус ивента больше не in_progress).
		decided := allyWins*2 > int(event.TargetGameCount) || enemyWins*2 > int(event.TargetGameCount)
		if !decided && newGameCount < event.TargetGameCount {
			return nil
		}

		// При чётном target_game_count (2) счёт побед вполне может сойтись
		// 1-1 — это настоящая ничья, а не "ещё не решено", поэтому пишем "draw",
		// а не оставляем пустую строку (target_game_count 1 и 3 ничьей не дают —
		// там всегда есть большинство).
		winnerSide = "draw"
		if allyWins > enemyWins {
			winnerSide = "ally"
		} else if enemyWins > allyWins {
			winnerSide = "enemy"
		}

		eventFinished = true
		return s.eventRepo.FinishEventDB(ctx, eventID, winnerSide, now)
	})
	if err != nil {
		return err
	}

	pubCtx, cancel := publishCtx(ctx)
	defer cancel()

	if err := s.producer.PublishTeamGameFinished(pubCtx, team1ID, team1ID == teamWinnerID, now); err != nil {
		return err
	}
	if err := s.producer.PublishTeamGameFinished(pubCtx, team2ID, team2ID == teamWinnerID, now); err != nil {
		return err
	}

	if !eventFinished {
		return nil
	}

	return s.producer.PublishEventFinished(pubCtx, eventID, winnerSide, now)
}

// AddTeamMemberStats заносит статистику ОДНОГО игрока за уже сыгранную игру —
// вызывается один раз на игрока, ПОСЛЕ того как FinishTeamGame закрыл игру
// для его команды (а не до/вместо неё, как раньше). user_id — сырой внешний
// id, как и везде в API; сервис сам резолвит его в user_event_id по паре
// (user_id, team_id), клиенту знать user_event_id не нужно.
func (s *eventService) AddTeamMemberStats(ctx context.Context, teamID, userID string, kills, deaths, points, revival, destroyedVehicles int64) error {
	if err := validateID("team_id", teamID); err != nil {
		return err
	}

	if err := validateID("user_id", userID); err != nil {
		return err
	}

	// Статистика уходит в суммарные total_* команды, поэтому мусор в ней
	// портит итоги всей команды, а не только одну строку.
	for _, stat := range []struct {
		field string
		value int64
	}{
		{"kills", kills},
		{"deaths", deaths},
		{"points", points},
		{"revival", revival},
		{"destroyed_vehicles", destroyedVehicles},
	} {
		if err := validateStat(stat.field, stat.value); err != nil {
			return err
		}
	}

	eventID, err := s.teamEventID(ctx, teamID)
	if err != nil {
		return err
	}

	unlock, err := s.eventLocks.lock(ctx, eventID)
	if err != nil {
		return err
	}
	defer unlock()

	var userEventID string
	err = s.eventRepo.WithTx(ctx, func(ctx context.Context) error {
		if _, err := s.lockEventRow(ctx, eventID); err != nil {
			return err
		}

		team, err := s.eventRepo.GetTeamByID(ctx, teamID)
		if err != nil {
			return err
		}

		if team.Status != domain.TeamStatusFinished {
			return domain.FailedPrecondition("game hasn't finished yet: stats can only be added after FinishTeamGame")
		}

		userEventID, err = s.eventRepo.GetUserEventIDByUserIDAndTeamID(ctx, userID, teamID)
		if err != nil {
			return err
		}

		if err := s.eventRepo.AddTeamMemberStats(ctx, teamID, userEventID, kills, deaths, points, revival, destroyedVehicles); err != nil {
			return err
		}

		// team.total_* — не то, что прислал именно этот вызов, а сумма по всем
		// team_members на данный момент, поэтому пересчитываем целиком, а не
		// прибавляем дельту (иначе повторный вызов на того же игрока с
		// исправленными числами задвоил бы итог). Сумма и запись итогов — в
		// той же транзакции, что и запись игрока: иначе параллельный вызов
		// мог записать итог, посчитанный до чужой записи, поверх более
		// свежего.
		sumKills, sumDeaths, sumPoints, sumRevival, sumDestroyed, err := s.eventRepo.SumTeamMemberStats(ctx, teamID)
		if err != nil {
			return err
		}

		return s.eventRepo.UpdateTeamTotals(ctx, teamID, sumKills, sumDeaths, sumPoints, sumRevival, sumDestroyed)
	})
	if err != nil {
		return err
	}

	pubCtx, cancel := publishCtx(ctx)
	defer cancel()

	return s.producer.PublishTeamMemberStatsAdded(pubCtx, teamID, userEventID, kills, deaths, points, revival, destroyedVehicles)
}

// GetTeamStats возвращает и статистику по каждому игроку, и итоговые Total*-поля
// команды — они не приходят от клиента одним числом, а каждый раз пересчитываются
// на сервере как сумма TeamMember (см. AddTeamMemberStats), чтобы клиенту не
// пришлось самому суммировать список, если нужен только общий итог.
func (s *eventService) GetTeamStats(ctx context.Context, teamID string) (domain.Team, []domain.TeamMember, error) {
	if err := validateID("team_id", teamID); err != nil {
		return domain.Team{}, nil, err
	}

	var team domain.Team
	var stats []domain.TeamMember
	// Итоги команды и статистика игроков — из одного снимка БД, иначе между
	// двумя запросами мог пройти AddTeamMemberStats, и total_* не сошлись бы
	// с суммой по игрокам.
	err := s.eventRepo.WithTx(ctx, func(ctx context.Context) error {
		var err error
		team, err = s.eventRepo.GetTeamByID(ctx, teamID)
		if err != nil {
			return err
		}

		if team.TeamID == "" {
			return domain.NotFound("team not found")
		}

		stats, err = s.eventRepo.GetTeamStats(ctx, teamID)
		return err
	})
	if err != nil {
		return domain.Team{}, nil, err
	}

	return team, stats, nil
}
