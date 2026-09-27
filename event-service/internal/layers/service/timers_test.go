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

	if producer.count("server.rent") != 1 {
		t.Fatalf("сообщений об аренде %d, ожидалось 1", producer.count("server.rent"))
	}
	if !repo.events[ev.id].RentServerSent {
		t.Error("после отправки должен ставиться флаг rent_server_sent")
	}

	// Повторный проход цепочки (например, после рестарта сервиса) второй раз
	// аренду не заказывает.
	requireNoErr(t, svc.rentServer(context.Background(), ev.id, ev.timeStart))
	if producer.count("server.rent") != 1 {
		t.Errorf("после рестарта аренда ушла повторно: %d сообщений", producer.count("server.rent"))
	}
}

func TestRentServerSkipsUnconfirmedEvent(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	// Ивент ещё pending.
	requireNoErr(t, svc.rentServer(context.Background(), ev.id, ev.timeStart))
	if producer.count("server.rent") != 0 {
		t.Error("для неподтверждённого ивента аренда не заказывается")
	}

	requireNoErr(t, svc.CancelEvent(context.Background(), ev.id, ev.creator))
	requireNoErr(t, svc.rentServer(context.Background(), ev.id, ev.timeStart))

	if producer.count("server.rent") != 0 {
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
	producer.fail("server.rent", errors.New("kafka недоступна"))

	if err := svc.rentServer(context.Background(), ev.id, ev.timeStart); err == nil {
		t.Fatal("ошибка публикации должна возвращаться")
	}

	// Флаг не выставлен — значит после рестарта сервиса аренда уйдёт заново,
	// а не потеряется совсем.
	if repo.events[ev.id].RentServerSent {
		t.Error("флаг rent_server_sent выставлен, хотя сообщение не ушло")
	}
}

func TestRecoverPendingEventsReportsRepositoryFailure(t *testing.T) {
	svc, repo, _ := newTestService(t)
	repo.fail("GetAllUnfinishedEvents", errors.New("база недоступна"))

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

	if producer.count("server.rent") != 0 || producer.count("event.started") != 0 {
		t.Error("у отклонённого ивента не должно быть ни аренды, ни старта")
	}
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

// ── цепочка: аренда сервера и ожидание его деплоя ───────────────────────────

func TestRunRentAndWaitServerOrdersServerAndWaits(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	setEventFields(repo, ev.id, func(event *domain.Event) {
		event.Status = domain.EventStatusConfirmed
		event.UserCount = MinPlayersRequired
		// Старт по расписанию давно прошёл, сервера так и нет.
		event.TimeStart = time.Now().Add(-domain.ServerDeployTimeout - time.Hour)
	})

	svc.runRentAndWaitServer(context.Background(), ev.id, time.Now().Add(-domain.ServerDeployTimeout-time.Hour))

	if producer.count("server.rent") != 1 {
		t.Errorf("сообщений об аренде %d, ожидалось 1", producer.count("server.rent"))
	}

	// Сервер не приехал — ивент отменён, а не завис в confirmed.
	if got := eventStatus(repo, ev.id); got != domain.EventStatusCanceled {
		t.Errorf("статус %s, ожидался canceled", got)
	}
	if producer.count("event.canceled") != 1 {
		t.Error("не опубликована отмена ивента без сервера")
	}
	if producer.count("event.started") != 0 {
		t.Error("ивент без сервера стартовать не должен")
	}
}

func TestRunRentAndWaitServerKeepsEventWhenServerArrived(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	setEventFields(repo, ev.id, func(event *domain.Event) {
		// Сервер уже доложил о себе (статус deployed) — отменять нечего.
		event.Status = domain.EventStatusDeployed
		event.UserCount = MinPlayersRequired
		event.ServerDeployedAt = time.Now()
	})

	svc.runRentAndWaitServer(context.Background(), ev.id, time.Now().Add(-domain.ServerDeployTimeout-time.Hour))

	if got := eventStatus(repo, ev.id); got != domain.EventStatusDeployed {
		t.Errorf("статус %s, ожидался deployed: сервер приехал, отменять нечего", got)
	}
	if producer.count("event.canceled") != 0 {
		t.Error("ивент с сервером отменён по таймауту деплоя")
	}
}

func TestRunRentAndWaitServerContinuesAfterRentFailure(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	setEventFields(repo, ev.id, func(event *domain.Event) {
		event.Status = domain.EventStatusConfirmed
		event.UserCount = MinPlayersRequired
	})
	producer.fail("server.rent", errKafka)

	// Аренда не ушла, но цепочка обязана дойти до таймаута деплоя.
	svc.runRentAndWaitServer(context.Background(), ev.id, time.Now().Add(-domain.ServerDeployTimeout-time.Hour))

	if got := eventStatus(repo, ev.id); got != domain.EventStatusCanceled {
		t.Errorf("статус %s, ожидался canceled", got)
	}
}

func TestRunRentAndWaitServerStopsWhenChainCanceled(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	setEventFields(repo, ev.id, func(event *domain.Event) {
		event.Status = domain.EventStatusConfirmed
		event.UserCount = MinPlayersRequired
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	svc.runRentAndWaitServer(ctx, ev.id, time.Now().Add(-time.Hour))

	if producer.count("server.rent") != 0 {
		t.Error("отменённая цепочка заказала сервер")
	}
	if got := eventStatus(repo, ev.id); got != domain.EventStatusConfirmed {
		t.Errorf("статус %s, ожидался confirmed", got)
	}
}

func TestScheduleRentAndWaitServerResumesConfirmedEvent(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	setEventFields(repo, ev.id, func(event *domain.Event) {
		event.Status = domain.EventStatusConfirmed
		event.UserCount = MinPlayersRequired
	})

	svc.scheduleRentAndWaitServer(ev.id, time.Now().Add(-time.Hour))

	waitFor(t, func() bool { return producer.count("server.rent") == 1 }, "аренда должна уйти после восстановления")

	if producer.count("event.confirmed") != 0 {
		t.Error("для уже подтверждённого ивента подтверждение не публикуется повторно")
	}
}

// ── цепочка: гейт StartEvent и ожидание нажатий ─────────────────────────────

func TestServerGateChainOpensStartEvent(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	setEventFields(repo, ev.id, func(event *domain.Event) {
		event.Status = domain.EventStatusDeployed
		event.ServerDeployedAt = time.Now().Add(-domain.ServerReadyDelay - time.Minute)
	})

	// Гейт уже должен быть открыт, а до автостарта ещё есть время.
	svc.scheduleServerGate(ev.id, time.Now().Add(-time.Minute))

	waitFor(t, func() bool { return eventStatus(repo, ev.id) == domain.EventStatusReady }, "ивент должен перейти в ready")

	if producer.count("event.ready") != 1 {
		t.Errorf("сообщений event.ready %d, ожидалось 1", producer.count("event.ready"))
	}
}

func TestOpenStartGateSkipsCanceledEvent(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	setEventFields(repo, ev.id, func(event *domain.Event) {
		event.ServerDeployedAt = time.Now()
	})
	requireNoErr(t, svc.CancelEvent(context.Background(), ev.id, ev.creator))
	// Отменённый ивент гейт не открывает.

	requireNoErr(t, svc.openStartGate(context.Background(), ev.id, ev.timeStart))

	if got := eventStatus(repo, ev.id); got != domain.EventStatusCanceled {
		t.Errorf("статус %s, ожидался canceled", got)
	}
	if producer.count("event.ready") != 0 {
		t.Error("для отменённого ивента StartEvent открываться не должен")
	}
}

func TestResolveStartVoteStartsWithOneSideReady(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 2)
	confirmEvent(t, svc, repo, ev)
	deployServer(t, svc, repo, ev)

	// Нажал только создатель — второго лидера ждать не будем.
	requireNoErr(t, svc.StartEvent(context.Background(), ev.id, ev.creator))
	if got := eventStatus(repo, ev.id); got != domain.EventStatusReady {
		t.Fatalf("после одного нажатия статус %s, ожидался ready", got)
	}

	requireNoErr(t, svc.resolveStartVote(context.Background(), ev.id))

	if got := eventStatus(repo, ev.id); got != domain.EventStatusInProgress {
		t.Errorf("статус %s, ожидался in_progress", got)
	}
	if producer.count("event.started") != 1 {
		t.Error("не опубликован старт ивента")
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

func TestResolveStartVoteCancelsWhenNobodyReady(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	confirmEvent(t, svc, repo, ev)
	deployServer(t, svc, repo, ev)

	requireNoErr(t, svc.resolveStartVote(context.Background(), ev.id))

	if got := eventStatus(repo, ev.id); got != domain.EventStatusCanceled {
		t.Errorf("статус %s, ожидался canceled: играть некому", got)
	}
	if producer.count("event.canceled") != 1 {
		t.Error("не опубликована отмена ивента")
	}
	if producer.count("event.started") != 0 {
		t.Error("ивент без готовых сторон стартовать не должен")
	}
}

func TestResolveStartVoteIgnoresAlreadyStartedEvent(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	startEvent(t, svc, repo, ev)

	before := len(producer.kinds())
	requireNoErr(t, svc.resolveStartVote(context.Background(), ev.id))

	if got := eventStatus(repo, ev.id); got != domain.EventStatusInProgress {
		t.Errorf("статус %s, ожидался in_progress", got)
	}
	if len(producer.kinds()) != before {
		t.Errorf("по уже стартовавшему ивенту ушли лишние сообщения: %v", producer.kinds()[before:])
	}
}

func TestStartVoteChainStartsEventOnTimeout(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	confirmEvent(t, svc, repo, ev)
	deployServer(t, svc, repo, ev)
	requireNoErr(t, svc.StartEvent(context.Background(), ev.id, ev.enemy))

	// Дедлайн голосования уже прошёл — цепочка должна стартовать сама.
	svc.scheduleStartVote(ev.id, time.Now().Add(-time.Minute))

	waitFor(t, func() bool { return eventStatus(repo, ev.id) == domain.EventStatusInProgress }, "ивент должен стартовать по таймауту")

	if producer.count("event.started") != 1 {
		t.Errorf("сообщений о старте %d, ожидалось 1", producer.count("event.started"))
	}
}

func TestTimerStepsSkipCanceledChain(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	setEventFields(repo, ev.id, func(event *domain.Event) {
		event.Status = domain.EventStatusConfirmed
		event.UserCount = MinPlayersRequired
		event.ServerDeployedAt = time.Now().Add(-domain.ServerReadyDelay - time.Minute)
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	confirmed, err := svc.checkMinPlayers(ctx, ev.id)
	requireNoErr(t, err)
	if confirmed {
		t.Error("отменённая цепочка не должна подтверждать ивент")
	}

	requireNoErr(t, svc.rentServer(ctx, ev.id, ev.timeStart))
	requireNoErr(t, svc.openStartGate(ctx, ev.id, ev.timeStart))
	requireNoErr(t, svc.resolveStartVote(ctx, ev.id))
	requireNoErr(t, svc.cancelEventWithoutServer(ctx, ev.id))

	if got := eventStatus(repo, ev.id); got != domain.EventStatusConfirmed {
		t.Errorf("статус изменился на %s, хотя цепочка отменена", got)
	}
	if len(producer.kinds()) != 3 {
		t.Errorf("по отменённой цепочке ушли лишние сообщения: %v", producer.kinds())
	}
}

func TestRecoverPendingEventsReschedulesTimers(t *testing.T) {
	svc, repo, _ := newTestService(t)

	pending := createTestEvent(t, svc, 1)
	awaitingServer := createTestEvent(t, svc, 1)
	deployed := createTestEvent(t, svc, 1)
	ready := createTestEvent(t, svc, 1)
	running := createTestEvent(t, svc, 1)
	finished := createTestEvent(t, svc, 1)

	setEventFields(repo, awaitingServer.id, func(event *domain.Event) {
		event.Status = domain.EventStatusConfirmed
	})
	setEventFields(repo, deployed.id, func(event *domain.Event) {
		event.Status = domain.EventStatusConfirmed
		event.ServerDeployedAt = time.Now().Add(time.Hour)
	})
	setEventFields(repo, ready.id, func(event *domain.Event) {
		event.Status = domain.EventStatusReady
		event.ServerDeployedAt = time.Now().Add(time.Hour)
	})
	setEventFields(repo, running.id, func(event *domain.Event) {
		event.Status = domain.EventStatusInProgress
	})
	setEventFields(repo, finished.id, func(event *domain.Event) {
		event.Status = domain.EventStatusFinished
	})

	// Рестарт процесса: все таймеры в памяти потеряны.
	for _, ev := range []testEvent{pending, awaitingServer, deployed, ready, running, finished} {
		svc.cancelEventTimer(ev.id)
	}
	if len(svc.eventTimers) != 0 {
		t.Fatalf("перед восстановлением таймеров быть не должно, есть %d", len(svc.eventTimers))
	}

	requireNoErr(t, svc.RecoverPendingEvents(context.Background()))

	for _, c := range []struct {
		name string
		ev   testEvent
		want bool
	}{
		{"pending — проверка минимума игроков", pending, true},
		{"confirmed без сервера — аренда и ожидание деплоя", awaitingServer, true},
		{"confirmed с сервером — ожидание открытия StartEvent", deployed, true},
		{"ready — ожидание нажатий сайд-лидеров", ready, true},
		{"in_progress — таймеры не нужны", running, false},
		{"finished — таймеры не нужны", finished, false},
	} {
		_, ok := svc.eventTimers[c.ev.id]
		if ok != c.want {
			t.Errorf("%s: цепочка таймеров %v, ожидалось %v", c.name, ok, c.want)
		}
	}
}
