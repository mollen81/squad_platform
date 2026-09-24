package service

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "event-service/internal/core/domain"
)

func TestCheckMinPlayersConfirmsWhenEnoughPlayers(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	event := repo.events[ev.id]
	event.UserCount = MinPlayersRequired
	repo.events[ev.id] = event

	confirmed, err := svc.checkMinPlayers(context.Background(), ev.id)
	requireNoErr(t, err)

	if !confirmed {
		t.Fatal("при достаточном количестве игроков ивент должен подтверждаться")
	}
	if got := repo.events[ev.id].Status; got != domain.EventStatusConfirmed {
		t.Errorf("статус %s, ожидался confirmed", got)
	}
	if producer.count("event.confirmed") != 1 {
		t.Error("не опубликовано подтверждение ивента")
	}
	if producer.count("event.declined") != 0 {
		t.Error("опубликован отказ, хотя игроков хватило")
	}
}

func TestCheckMinPlayersDeclinesWhenNotEnoughPlayers(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	event := repo.events[ev.id]
	event.UserCount = MinPlayersRequired - 1
	repo.events[ev.id] = event

	confirmed, err := svc.checkMinPlayers(context.Background(), ev.id)
	requireNoErr(t, err)

	if confirmed {
		t.Fatal("ивент не должен подтверждаться без минимума игроков")
	}
	if got := repo.events[ev.id].Status; got != domain.EventStatusDeclined {
		t.Errorf("статус %s, ожидался declined", got)
	}
	if producer.count("event.declined") != 1 {
		t.Error("не опубликован отказ")
	}
}

func TestCheckMinPlayersSkipsEventThatLeftPending(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	// Создатель успел отменить ивент до контрольной точки.
	requireNoErr(t, svc.CancelEvent(context.Background(), ev.id, ev.creator))
	before := len(producer.kinds())

	confirmed, err := svc.checkMinPlayers(context.Background(), ev.id)
	requireNoErr(t, err)

	if confirmed {
		t.Error("отменённый ивент не должен подтверждаться")
	}
	if got := repo.events[ev.id].Status; got != domain.EventStatusCanceled {
		t.Errorf("статус %s, ожидался canceled", got)
	}
	if len(producer.kinds()) != before {
		t.Errorf("по отменённому ивенту ушли лишние сообщения: %v", producer.kinds()[before:])
	}
}

func TestCheckMinPlayersFailsOnMissingEvent(t *testing.T) {
	svc, _, _ := newTestService(t)

	_, err := svc.checkMinPlayers(context.Background(), newID())
	requireKind(t, err, domain.KindNotFound)
}

func TestRentServerPublishesOnlyOnce(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	event := repo.events[ev.id]
	event.Status = domain.EventStatusConfirmed
	repo.events[ev.id] = event

	requireNoErr(t, svc.rentServer(context.Background(), ev.id, ev.timeStart))

	if producer.count("rent.server") != 1 {
		t.Fatalf("сообщений об аренде %d, ожидалось 1", producer.count("rent.server"))
	}
	if !repo.events[ev.id].RentServerSent {
		t.Error("после отправки должен ставиться флаг rent_server_sent")
	}

	// Повторный проход цепочки (например, после рестарта сервиса) второй раз
	// аренду не заказывает.
	requireNoErr(t, svc.rentServer(context.Background(), ev.id, ev.timeStart))
	if producer.count("rent.server") != 1 {
		t.Errorf("после рестарта аренда ушла повторно: %d сообщений", producer.count("rent.server"))
	}
}

func TestRentServerSkipsUnconfirmedEvent(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	// Ивент ещё pending.
	requireNoErr(t, svc.rentServer(context.Background(), ev.id, ev.timeStart))
	if producer.count("rent.server") != 0 {
		t.Error("для неподтверждённого ивента аренда не заказывается")
	}

	requireNoErr(t, svc.CancelEvent(context.Background(), ev.id, ev.creator))
	requireNoErr(t, svc.rentServer(context.Background(), ev.id, ev.timeStart))

	if producer.count("rent.server") != 0 {
		t.Error("для отменённого ивента аренда не заказывается")
	}
	if repo.events[ev.id].RentServerSent {
		t.Error("флаг отправки не должен ставиться без самой отправки")
	}
}

func TestRentServerKeepsFlagUnsetWhenPublishFails(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	event := repo.events[ev.id]
	event.Status = domain.EventStatusConfirmed
	repo.events[ev.id] = event
	producer.failOn["rent.server"] = errors.New("kafka недоступна")

	if err := svc.rentServer(context.Background(), ev.id, ev.timeStart); err == nil {
		t.Fatal("ошибка публикации должна возвращаться")
	}

	// Флаг не выставлен — значит после рестарта сервиса аренда уйдёт заново,
	// а не потеряется совсем.
	if repo.events[ev.id].RentServerSent {
		t.Error("флаг rent_server_sent выставлен, хотя сообщение не ушло")
	}
}

func TestStartEventAndGamesStartsFirstGame(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 2)

	event := repo.events[ev.id]
	event.Status = domain.EventStatusConfirmed
	repo.events[ev.id] = event

	requireNoErr(t, svc.startEventAndGames(context.Background(), ev.id))

	if got := repo.events[ev.id].Status; got != domain.EventStatusInProgress {
		t.Errorf("статус ивента %s, ожидался in_progress", got)
	}

	ally1, enemy1 := ev.teams(t, svc, 1)
	for _, teamID := range []string{ally1, enemy1} {
		if got := repo.teams[teamID].Status; got != domain.TeamStatusInProgress {
			t.Errorf("команда первой игры в статусе %s, ожидался in_progress", got)
		}
	}

	ally2, _ := ev.teams(t, svc, 2)
	if got := repo.teams[ally2].Status; got != domain.TeamStatusPending {
		t.Errorf("вторая игра не должна стартовать вместе с ивентом, статус %s", got)
	}

	if producer.count("event.started") != 1 {
		t.Error("не опубликован старт ивента")
	}
	if producer.count("team_game.started") != 2 {
		t.Errorf("сообщений о старте команд %d, ожидалось 2", producer.count("team_game.started"))
	}
}

func TestStartEventAndGamesRequiresConfirmedEvent(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	// Ивент отменили до старта — цепочка таймеров не должна его запустить.
	requireNoErr(t, svc.CancelEvent(context.Background(), ev.id, ev.creator))
	before := len(producer.kinds())

	err := svc.startEventAndGames(context.Background(), ev.id)
	requireKind(t, err, domain.KindFailedPrecondition)

	if got := repo.events[ev.id].Status; got != domain.EventStatusCanceled {
		t.Errorf("статус %s, ожидался canceled", got)
	}

	ally, _ := ev.teams(t, svc, 1)
	if got := repo.teams[ally].Status; got != domain.TeamStatusPending {
		t.Errorf("у отменённого ивента игра ушла в статус %s", got)
	}
	if len(producer.kinds()) != before {
		t.Errorf("по отменённому ивенту ушли лишние сообщения: %v", producer.kinds()[before:])
	}
}

func TestRecoverPendingEventsReschedulesTimers(t *testing.T) {
	svc, repo, _ := newTestService(t)

	pending := createTestEvent(t, svc, 1)
	confirmed := createTestEvent(t, svc, 1)
	running := createTestEvent(t, svc, 1)
	finished := createTestEvent(t, svc, 1)

	setStatus := func(ev testEvent, status domain.EventStatus) {
		event := repo.events[ev.id]
		event.Status = status
		repo.events[ev.id] = event
	}
	setStatus(confirmed, domain.EventStatusConfirmed)
	setStatus(running, domain.EventStatusInProgress)
	setStatus(finished, domain.EventStatusFinished)

	// Рестарт процесса: все таймеры в памяти потеряны.
	for _, ev := range []testEvent{pending, confirmed, running, finished} {
		svc.cancelEventTimer(ev.id)
	}
	if len(svc.eventTimers) != 0 {
		t.Fatalf("перед восстановлением таймеров быть не должно, есть %d", len(svc.eventTimers))
	}

	requireNoErr(t, svc.RecoverPendingEvents(context.Background()))

	if _, ok := svc.eventTimers[pending.id]; !ok {
		t.Error("для pending-ивента не восстановлена проверка минимума игроков")
	}
	if _, ok := svc.eventTimers[confirmed.id]; !ok {
		t.Error("для confirmed-ивента не восстановлены аренда и старт")
	}
	if _, ok := svc.eventTimers[running.id]; ok {
		t.Error("для идущего ивента таймеры не нужны")
	}
	if _, ok := svc.eventTimers[finished.id]; ok {
		t.Error("для завершённого ивента таймеры не нужны")
	}
}

func TestRecoverPendingEventsReportsRepositoryFailure(t *testing.T) {
	svc, repo, _ := newTestService(t)
	repo.failOn["GetAllUnfinishedEvents"] = errors.New("база недоступна")

	if err := svc.RecoverPendingEvents(context.Background()); err == nil {
		t.Fatal("ошибка чтения незавершённых ивентов должна возвращаться")
	}
}

func TestEventTimerChainReplacedNotDuplicated(t *testing.T) {
	svc, _, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	first := svc.eventTimers[ev.id]
	svc.controlEventTimerDenial(ev.id, time.Now().Add(4*time.Hour))
	second := svc.eventTimers[ev.id]

	if first == second {
		t.Fatal("повторная постановка должна заменять цепочку")
	}

	// Старая цепочка, завершаясь, не должна снести новую.
	svc.finishEventTimer(ev.id, first)
	if svc.eventTimers[ev.id] != second {
		t.Error("завершение старой цепочки стёрло новую")
	}

	svc.cancelEventTimer(ev.id)
	if _, ok := svc.eventTimers[ev.id]; ok {
		t.Error("после отмены запись о цепочке должна исчезнуть")
	}
}

func TestWaitUntil(t *testing.T) {
	t.Run("момент в прошлом срабатывает сразу", func(t *testing.T) {
		if !waitUntil(context.Background(), time.Now().Add(-time.Hour)) {
			t.Error("пропущенный момент должен срабатывать немедленно, а не теряться")
		}
	})

	t.Run("отменённая цепочка не срабатывает", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if waitUntil(ctx, time.Now().Add(-time.Hour)) {
			t.Error("для отменённой цепочки шаг выполняться не должен")
		}
	})

	t.Run("отмена во время ожидания", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(20 * time.Millisecond)
			cancel()
		}()

		done := make(chan bool, 1)
		go func() { done <- waitUntil(ctx, time.Now().Add(time.Hour)) }()

		select {
		case got := <-done:
			if got {
				t.Error("ожидание прервано отменой — шаг выполняться не должен")
			}
		case <-time.After(2 * time.Second):
			t.Error("waitUntil не среагировал на отмену")
		}
	})
}

func TestPublishCtxSurvivesRequestCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	pubCtx, release := publishCtx(ctx)
	defer release()

	// Клиент отвалился уже после коммита — сообщение всё равно должно уйти.
	cancel()

	if err := pubCtx.Err(); err != nil {
		t.Errorf("контекст публикации отменился вместе с запросом: %v", err)
	}
	if _, ok := pubCtx.Deadline(); !ok {
		t.Error("у публикации должен быть свой предел ожидания")
	}
}

// Контрольная точка уже позади (сервис лежал дольше, чем оставалось до неё) —
// цепочка должна отработать сразу, а не потерять ивент.
func TestTimerChainDeclinesEventWithoutPlayers(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	svc.controlEventTimerDenial(ev.id, time.Now().Add(-time.Hour))

	waitFor(t, func() bool {
		return eventStatus(repo, ev.id) == domain.EventStatusDeclined
	}, "ивент без игроков должен быть отклонён")

	waitFor(t, func() bool {
		return producer.count("event.declined") == 1
	}, "отказ должен быть опубликован")

	if producer.count("rent.server") != 0 || producer.count("event.started") != 0 {
		t.Error("у отклонённого ивента не должно быть ни аренды, ни старта")
	}
}

// Полный проход цепочки: контроль → аренда сервера → старт ивента и первой игры.
func TestTimerChainRunsWholeLifecycle(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 2)

	setEventFields(repo, ev.id, func(event *domain.Event) {
		event.UserCount = MinPlayersRequired
	})

	svc.controlEventTimerDenial(ev.id, time.Now().Add(-time.Hour))

	waitFor(t, func() bool {
		return eventStatus(repo, ev.id) == domain.EventStatusInProgress
	}, "ивент должен пройти путь до in_progress")

	waitFor(t, func() bool {
		return producer.count("event.started") == 1
	}, "старт ивента должен быть опубликован")

	for _, kind := range []string{"event.confirmed", "rent.server", "event.started"} {
		if producer.count(kind) != 1 {
			t.Errorf("сообщений %s: %d, ожидалось 1", kind, producer.count(kind))
		}
	}

	repo.mu.Lock()
	rentSent := repo.events[ev.id].RentServerSent
	repo.mu.Unlock()
	if !rentSent {
		t.Error("после отправки аренды должен стоять флаг rent_server_sent")
	}

	ally1, enemy1 := ev.teams(t, svc, 1)
	for _, teamID := range []string{ally1, enemy1} {
		repo.mu.Lock()
		status := repo.teams[teamID].Status
		repo.mu.Unlock()

		if status != domain.TeamStatusInProgress {
			t.Errorf("команда первой игры в статусе %s, ожидался in_progress", status)
		}
	}
}

// Вторая половина цепочки после рестарта: подтверждение не публикуется заново.
func TestScheduleRentAndStartResumesConfirmedEvent(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	setEventFields(repo, ev.id, func(event *domain.Event) {
		event.Status = domain.EventStatusConfirmed
		event.UserCount = MinPlayersRequired
	})

	svc.scheduleRentAndStart(ev.id, time.Now().Add(-time.Hour))

	waitFor(t, func() bool {
		return eventStatus(repo, ev.id) == domain.EventStatusInProgress
	}, "подтверждённый ивент должен стартовать после восстановления")

	if producer.count("event.confirmed") != 0 {
		t.Error("для уже подтверждённого ивента подтверждение не публикуется повторно")
	}
	if producer.count("rent.server") != 1 {
		t.Errorf("сообщений об аренде %d, ожидалось 1", producer.count("rent.server"))
	}
}

// Отмена ивента гасит всю цепочку: ни аренды, ни старта быть не должно.
func TestCancelStopsTimerChainBeforeStart(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	setEventFields(repo, ev.id, func(event *domain.Event) {
		event.Status = domain.EventStatusConfirmed
		event.UserCount = MinPlayersRequired
	})

	// Цепочка ждёт момента старта, который наступит нескоро.
	svc.scheduleRentAndStart(ev.id, time.Now().Add(time.Hour))
	requireNoErr(t, svc.CancelEvent(context.Background(), ev.id, ev.creator))

	time.Sleep(50 * time.Millisecond)

	if got := eventStatus(repo, ev.id); got != domain.EventStatusCanceled {
		t.Errorf("статус %s, ожидался canceled", got)
	}
	if producer.count("event.started") != 0 {
		t.Error("отменённый ивент не должен стартовать")
	}

	ally, _ := ev.teams(t, svc, 1)
	repo.mu.Lock()
	status := repo.teams[ally].Status
	repo.mu.Unlock()
	if status != domain.TeamStatusPending {
		t.Errorf("у отменённого ивента игра ушла в статус %s", status)
	}
}

// Шаг цепочки, проснувшийся после отмены, не должен ничего делать.
func TestTimerStepsSkipCanceledChain(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	setEventFields(repo, ev.id, func(event *domain.Event) {
		event.Status = domain.EventStatusConfirmed
		event.UserCount = MinPlayersRequired
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	confirmed, err := svc.checkMinPlayers(ctx, ev.id)
	requireNoErr(t, err)
	if confirmed {
		t.Error("отменённая цепочка не должна подтверждать ивент")
	}

	requireNoErr(t, svc.rentServer(ctx, ev.id, ev.timeStart))
	requireNoErr(t, svc.startEventAndGames(ctx, ev.id))

	if got := eventStatus(repo, ev.id); got != domain.EventStatusConfirmed {
		t.Errorf("статус изменился на %s, хотя цепочка отменена", got)
	}
	if len(producer.kinds()) != 3 {
		t.Errorf("по отменённой цепочке ушли лишние сообщения: %v", producer.kinds())
	}
}

func TestRunRentAndStartStopsWhenChainCanceled(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	setEventFields(repo, ev.id, func(event *domain.Event) {
		event.Status = domain.EventStatusConfirmed
		event.UserCount = MinPlayersRequired
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	svc.runRentAndStart(ctx, ev.id, time.Now().Add(-time.Hour))

	if got := eventStatus(repo, ev.id); got != domain.EventStatusConfirmed {
		t.Errorf("статус %s, ожидался confirmed: отменённая цепочка ничего не делает", got)
	}
	if producer.count("rent.server") != 0 || producer.count("event.started") != 0 {
		t.Errorf("отменённая цепочка опубликовала %v", producer.kinds())
	}
}

func TestRunRentAndStartContinuesAfterRentFailure(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	setEventFields(repo, ev.id, func(event *domain.Event) {
		event.Status = domain.EventStatusConfirmed
		event.UserCount = MinPlayersRequired
	})
	producer.failOn["rent.server"] = errors.New("kafka недоступна")

	// Аренда не ушла, но ивент всё равно должен стартовать.
	svc.runRentAndStart(context.Background(), ev.id, time.Now().Add(-time.Hour))

	if got := eventStatus(repo, ev.id); got != domain.EventStatusInProgress {
		t.Errorf("статус %s, ожидался in_progress несмотря на сбой аренды", got)
	}
	if producer.count("event.started") != 1 {
		t.Error("старт ивента должен быть опубликован")
	}
}

func TestStartEventAndGamesFailsWithoutTeams(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	setEventFields(repo, ev.id, func(event *domain.Event) {
		event.Status = domain.EventStatusConfirmed
	})

	// Команд нет — стартовать нечего, и статус меняться не должен.
	repo.mu.Lock()
	for teamID, team := range repo.teams {
		if team.EventID == ev.id {
			delete(repo.teams, teamID)
		}
	}
	repo.mu.Unlock()

	if err := svc.startEventAndGames(context.Background(), ev.id); err == nil {
		t.Fatal("без команд ивент стартовать не может")
	}

	if got := eventStatus(repo, ev.id); got != domain.EventStatusConfirmed {
		t.Errorf("статус %s, ожидался confirmed после отката", got)
	}
}

func TestRunRentAndStartLogsStartFailure(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	setEventFields(repo, ev.id, func(event *domain.Event) {
		event.Status = domain.EventStatusConfirmed
	})
	repo.failOn["StartTeamGame"] = errDB

	// Сбой старта не должен ронять цепочку — он только логируется.
	svc.runRentAndStart(context.Background(), ev.id, time.Now().Add(-time.Hour))

	if producer.count("rent.server") != 1 {
		t.Errorf("аренда должна была уйти до сбоя старта, сообщений %d", producer.count("rent.server"))
	}
	if got := eventStatus(repo, ev.id); got != domain.EventStatusConfirmed {
		t.Errorf("статус %s, ожидался confirmed", got)
	}
}

func TestStartEventAndGamesFailsOnStatusUpdate(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	setEventFields(repo, ev.id, func(event *domain.Event) {
		event.Status = domain.EventStatusConfirmed
	})
	repo.failOn["UpdateEventStatus"] = errDB

	requireInternal(t, svc.startEventAndGames(context.Background(), ev.id))
}

func TestWaitUntilWaitsForFutureMoment(t *testing.T) {
	start := time.Now()

	if !waitUntil(context.Background(), start.Add(30*time.Millisecond)) {
		t.Fatal("ожидание должно завершиться срабатыванием, а не отменой")
	}

	if elapsed := time.Since(start); elapsed < 25*time.Millisecond {
		t.Errorf("сработало раньше срока: прошло %v", elapsed)
	}
}

func TestRunRentAndStartStopsBetweenRentAndStart(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	setEventFields(repo, ev.id, func(event *domain.Event) {
		event.Status = domain.EventStatusConfirmed
		event.UserCount = MinPlayersRequired
	})

	// Аренда — почти сразу, старт — нескоро: цепочку отменяем в промежутке.
	timeStart := time.Now().Add(RentServerTimeBeforeStart + 50*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		svc.runRentAndStart(ctx, ev.id, timeStart)
		close(done)
	}()

	waitFor(t, func() bool { return producer.count("rent.server") == 1 }, "аренда должна уйти")
	cancel()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("цепочка не остановилась после отмены")
	}

	if got := eventStatus(repo, ev.id); got != domain.EventStatusConfirmed {
		t.Errorf("статус %s: отменённая цепочка не должна стартовать ивент", got)
	}
	if producer.count("event.started") != 0 {
		t.Error("отменённая цепочка опубликовала старт ивента")
	}
}

func TestTimerStepsFailOnRepositoryErrors(t *testing.T) {
	t.Run("startEventAndGames: старт первой игры", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		setEventFields(repo, ev.id, func(event *domain.Event) {
			event.Status = domain.EventStatusConfirmed
		})
		repo.failOn["StartTeamGame"] = errDB

		requireInternal(t, svc.startEventAndGames(context.Background(), ev.id))

		// Статус и старт игры меняются одной транзакцией: ивент не должен
		// остаться in_progress без запущенной игры.
		if got := eventStatus(repo, ev.id); got != domain.EventStatusConfirmed {
			t.Errorf("статус ивента %s, ожидался confirmed после отката", got)
		}
	})

	t.Run("rentServer: список игроков", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		setEventFields(repo, ev.id, func(event *domain.Event) {
			event.Status = domain.EventStatusConfirmed
		})
		repo.failOn["GetUserIDsByEventID"] = errDB

		requireInternal(t, svc.rentServer(context.Background(), ev.id, ev.timeStart))
	})
}

func TestTimerStepsPublishFailuresAreReported(t *testing.T) {
	t.Run("startEventAndGames", func(t *testing.T) {
		svc, repo, producer := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		setEventFields(repo, ev.id, func(event *domain.Event) {
			event.Status = domain.EventStatusConfirmed
		})
		producer.failOn["event.started"] = errKafka

		if err := svc.startEventAndGames(context.Background(), ev.id); err == nil {
			t.Error("ошибка публикации должна возвращаться")
		}
	})

	t.Run("checkMinPlayers", func(t *testing.T) {
		svc, repo, producer := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		setEventFields(repo, ev.id, func(event *domain.Event) {
			event.UserCount = MinPlayersRequired
		})
		producer.failOn["event.confirmed"] = errKafka

		// Подтверждение уже записано в базу, поэтому цепочка обязана идти
		// дальше даже при сбое публикации.
		confirmed, err := svc.checkMinPlayers(context.Background(), ev.id)
		if err == nil {
			t.Error("ошибка публикации должна возвращаться")
		}
		if !confirmed {
			t.Error("при сбое публикации цепочка всё равно должна продолжиться: ивент подтверждён")
		}
	})
}
