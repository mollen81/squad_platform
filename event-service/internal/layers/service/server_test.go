package service

import (
	"context"
	"strings"
	"testing"
	"time"

	domain "event-service/internal/core/domain"
)

func TestServerPurchased(t *testing.T) {
	t.Run("невалидные данные", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		requireKind(t, svc.ServerPurchased(context.Background(), "не-uuid", newServerIP(), "pass"), domain.KindInvalidArgument)
		// Адрес сервера — строка, но пустой или неприлично длинный не принимается.
		requireKind(t, svc.ServerPurchased(context.Background(), ev.id, "", "pass"), domain.KindInvalidArgument)
		requireKind(t, svc.ServerPurchased(context.Background(), ev.id, "   ", "pass"), domain.KindInvalidArgument)
		requireKind(t, svc.ServerPurchased(context.Background(), ev.id, strings.Repeat("9", maxServerIPLength+1), "pass"), domain.KindInvalidArgument)
		requireKind(t, svc.ServerPurchased(context.Background(), ev.id, newServerIP(), ""), domain.KindInvalidArgument)
	})

	t.Run("несуществующий ивент", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		requireKind(t, svc.ServerPurchased(context.Background(), newID(), newID(), "pass"), domain.KindNotFound)
	})

	t.Run("только для подтверждённого ивента", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		// Ивент ещё pending: сервер под него никто не заказывал.
		requireKind(t, svc.ServerPurchased(context.Background(), ev.id, newID(), "pass"), domain.KindFailedPrecondition)

		confirmEvent(t, svc, repo, ev)
		requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, newID(), "pass"))
		expectStatus(t, repo, ev.id, domain.EventStatusPurchased)
	})

	t.Run("реквизиты сохраняются", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		confirmEvent(t, svc, repo, ev)
		serverIP := newServerIP()

		requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, serverIP, "secret"))
		expectStatus(t, repo, ev.id, domain.EventStatusPurchased)

		repo.mu.Lock()
		event := repo.events[ev.id]
		repo.mu.Unlock()

		if event.ServerIP != serverIP || event.ServerPassword != "secret" {
			t.Errorf("реквизиты сервера не сохранились: %q / %q", event.ServerIP, event.ServerPassword)
		}
	})

	t.Run("повторная доставка сообщения безвредна", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		confirmEvent(t, svc, repo, ev)
		serverIP := newServerIP()

		requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, serverIP, "secret"))
		requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, serverIP, "secret"))

		repo.mu.Lock()
		event := repo.events[ev.id]
		repo.mu.Unlock()

		if event.ServerIP != serverIP {
			t.Errorf("id сервера %q", event.ServerIP)
		}
		expectStatus(t, repo, ev.id, domain.EventStatusPurchased)
	})

	t.Run("второй, другой сервер не принимается", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		confirmEvent(t, svc, repo, ev)
		serverIP := newServerIP()
		requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, serverIP, "secret"))

		requireKind(t, svc.ServerPurchased(context.Background(), ev.id, newID(), "secret"), domain.KindFailedPrecondition)

		repo.mu.Lock()
		event := repo.events[ev.id]
		repo.mu.Unlock()

		if event.ServerIP != serverIP {
			t.Errorf("id сервера подменился на %q", event.ServerIP)
		}
	})
}

func TestServerDeployed(t *testing.T) {
	t.Run("невалидные данные", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		requireKind(t, svc.ServerDeployed(context.Background(), "не-uuid", newServerIP()), domain.KindInvalidArgument)
		requireKind(t, svc.ServerDeployed(context.Background(), ev.id, ""), domain.KindInvalidArgument)
		requireKind(t, svc.ServerDeployed(context.Background(), ev.id, strings.Repeat("9", maxServerIPLength+1)), domain.KindInvalidArgument)
	})

	t.Run("только для купленного сервера", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		confirmEvent(t, svc, repo, ev)

		// Сервер ещё не покупался — применять сообщение не к чему.
		requireKind(t, svc.ServerDeployed(context.Background(), ev.id, newID()), domain.KindFailedPrecondition)

		serverIP := newServerIP()
		requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, serverIP, "secret"))
		requireNoErr(t, svc.ServerDeployed(context.Background(), ev.id, serverIP))
		expectStatus(t, repo, ev.id, domain.EventStatusDeployed)
	})

	t.Run("чужой server_ip отклоняется", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		confirmEvent(t, svc, repo, ev)
		requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, newID(), "secret"))

		// Пары (event_id, server_ip) в базе нет — сообщение не про этот ивент.
		requireKind(t, svc.ServerDeployed(context.Background(), ev.id, newID()), domain.KindNotFound)
		expectStatus(t, repo, ev.id, domain.EventStatusPurchased)
	})

	t.Run("несуществующий ивент", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		requireKind(t, svc.ServerDeployed(context.Background(), newID(), newID()), domain.KindNotFound)
	})

	t.Run("момент готовности берётся по своим часам и ставится цепочка", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		confirmEvent(t, svc, repo, ev)
		serverIP := newServerIP()
		requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, serverIP, "secret"))

		before := time.Now()
		requireNoErr(t, svc.ServerDeployed(context.Background(), ev.id, serverIP))
		after := time.Now()

		repo.mu.Lock()
		event := repo.events[ev.id]
		repo.mu.Unlock()

		// Времени в сообщении нет: готовность считается с момента получения.
		if event.ServerDeployedAt.Before(before) || event.ServerDeployedAt.After(after) {
			t.Errorf("момент готовности %v вне интервала обработки [%v, %v]", event.ServerDeployedAt, before, after)
		}
		if got := event.StartAvailableAt(); !got.Equal(event.ServerDeployedAt.Add(domain.ServerReadyDelay)) {
			t.Errorf("StartEvent открывается в %v, ожидалось +%v от деплоя", got, domain.ServerReadyDelay)
		}
		if _, ok := svc.eventTimers[ev.id]; !ok {
			t.Error("после деплоя должна стоять цепочка ожидания StartEvent")
		}
	})

	t.Run("повторная доставка не сдвигает открытие StartEvent", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		confirmEvent(t, svc, repo, ev)

		serverIP := newServerIP()
		requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, serverIP, "secret"))
		requireNoErr(t, svc.ServerDeployed(context.Background(), ev.id, serverIP))

		repo.mu.Lock()
		first := repo.events[ev.id].ServerDeployedAt
		repo.mu.Unlock()

		time.Sleep(5 * time.Millisecond)
		requireNoErr(t, svc.ServerDeployed(context.Background(), ev.id, serverIP))

		repo.mu.Lock()
		second := repo.events[ev.id].ServerDeployedAt
		repo.mu.Unlock()

		if !second.Equal(first) {
			t.Errorf("момент готовности стал %v, ожидался первый — %v", second, first)
		}
	})

}

func TestStartEvent(t *testing.T) {
	t.Run("невалидные данные", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		requireKind(t, svc.StartEvent(context.Background(), "не-uuid", ev.creator), domain.KindInvalidArgument)
		requireKind(t, svc.StartEvent(context.Background(), ev.id, ""), domain.KindInvalidArgument)
	})

	t.Run("звать может только сайд-лидер", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		player := joinPlayer(t, svc, ev.id, newID(), false)
		confirmEvent(t, svc, repo, ev)
		deployServer(t, svc, repo, ev)

		requireKind(t, svc.StartEvent(context.Background(), ev.id, player), domain.KindPermissionDenied)
		requireKind(t, svc.StartEvent(context.Background(), ev.id, newID()), domain.KindPermissionDenied)
	})

	t.Run("на каждом шаге до ready — своя ошибка", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		// Ивент ещё pending: игроков не набралось.
		requireKind(t, svc.StartEvent(context.Background(), ev.id, ev.creator), domain.KindFailedPrecondition)

		// Подтверждён, но сервер не куплен.
		confirmEvent(t, svc, repo, ev)
		expectStartError(t, svc, ev, "server is not purchased yet")

		// Куплен, но vps.deployed ещё не приходило.
		serverIP := newServerIP()
		requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, serverIP, "secret"))
		expectStartError(t, svc, ev, "server is not deployed yet")

		// Задеплоился, но ServerReadyDelay ещё не прошёл: гейт закрыт.
		requireNoErr(t, svc.ServerDeployed(context.Background(), ev.id, serverIP))
		expectStartError(t, svc, ev, "server is not ready yet")
	})

	t.Run("одного нажатия недостаточно", func(t *testing.T) {
		svc, repo, producer := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		confirmEvent(t, svc, repo, ev)
		deployServer(t, svc, repo, ev)

		requireNoErr(t, svc.StartEvent(context.Background(), ev.id, ev.creator))

		if got := eventStatus(repo, ev.id); got != domain.EventStatusReady {
			t.Errorf("статус %s, ожидался ready: ждём второго лидера", got)
		}
		if producer.count("event.side_ready") != 1 {
			t.Errorf("сообщений о готовности %d, ожидалось 1", producer.count("event.side_ready"))
		}
		if producer.count("event.started") != 0 {
			t.Error("ивент стартовал по одному нажатию")
		}

		repo.mu.Lock()
		event := repo.events[ev.id]
		repo.mu.Unlock()

		if event.AllyReadyAt.IsZero() || !event.EnemyReadyAt.IsZero() {
			t.Errorf("отмечена не та сторона: ally=%v enemy=%v", event.AllyReadyAt, event.EnemyReadyAt)
		}
	})

	t.Run("повторное нажатие той же стороной", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		confirmEvent(t, svc, repo, ev)
		deployServer(t, svc, repo, ev)

		requireNoErr(t, svc.StartEvent(context.Background(), ev.id, ev.creator))
		requireKind(t, svc.StartEvent(context.Background(), ev.id, ev.creator), domain.KindFailedPrecondition)

		if got := eventStatus(repo, ev.id); got != domain.EventStatusReady {
			t.Errorf("статус %s, ожидался ready", got)
		}
	})

	t.Run("после второго нажатия ивент стартует", func(t *testing.T) {
		svc, repo, producer := newTestService(t)
		ev := createTestEvent(t, svc, 2)
		confirmEvent(t, svc, repo, ev)
		deployServer(t, svc, repo, ev)

		requireNoErr(t, svc.StartEvent(context.Background(), ev.id, ev.enemy))
		requireNoErr(t, svc.StartEvent(context.Background(), ev.id, ev.creator))

		if got := eventStatus(repo, ev.id); got != domain.EventStatusInProgress {
			t.Errorf("статус %s, ожидался in_progress", got)
		}
		if producer.count("event.side_ready") != 2 {
			t.Errorf("сообщений о готовности %d, ожидалось 2", producer.count("event.side_ready"))
		}
		if producer.count("event.started") != 1 {
			t.Error("не опубликован старт ивента")
		}
		if producer.count("team_game.started") != 2 {
			t.Errorf("сообщений о старте команд %d, ожидалось 2", producer.count("team_game.started"))
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

		ally2, _ := ev.teams(t, svc, 2)
		repo.mu.Lock()
		secondGame := repo.teams[ally2].Status
		repo.mu.Unlock()

		if secondGame != domain.TeamStatusPending {
			t.Errorf("вторая игра в статусе %s, стартовать должна только первая", secondGame)
		}
	})

	t.Run("после старта звать уже нельзя", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		startEvent(t, svc, repo, ev)

		requireKind(t, svc.StartEvent(context.Background(), ev.id, ev.creator), domain.KindFailedPrecondition)
	})

	t.Run("сбой репозитория", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		confirmEvent(t, svc, repo, ev)
		deployServer(t, svc, repo, ev)

		repo.fail("MarkSideReady", errDB)
		requireInternal(t, svc.StartEvent(context.Background(), ev.id, ev.creator))
	})

	t.Run("сбой публикации", func(t *testing.T) {
		svc, repo, producer := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		confirmEvent(t, svc, repo, ev)
		deployServer(t, svc, repo, ev)

		producer.fail("event.side_ready", errKafka)
		if err := svc.StartEvent(context.Background(), ev.id, ev.creator); err == nil {
			t.Error("ошибка публикации должна возвращаться")
		}

		// Нажатие уже закоммичено.
		repo.mu.Lock()
		event := repo.events[ev.id]
		repo.mu.Unlock()

		if event.AllyReadyAt.IsZero() {
			t.Error("нажатие должно остаться в базе: публикация идёт после коммита")
		}
	})
}

func TestGetServerData(t *testing.T) {
	t.Run("невалидные данные", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		_, _, err := svc.GetServerData(context.Background(), "не-uuid", ev.creator)
		requireKind(t, err, domain.KindInvalidArgument)

		_, _, err = svc.GetServerData(context.Background(), ev.id, "")
		requireKind(t, err, domain.KindInvalidArgument)
	})

	t.Run("несуществующий ивент", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		_, _, err := svc.GetServerData(context.Background(), newID(), newID())
		requireKind(t, err, domain.KindNotFound)
	})

	t.Run("только участникам ивента", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		confirmEvent(t, svc, repo, ev)
		requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, newID(), "secret"))

		_, _, err := svc.GetServerData(context.Background(), ev.id, newID())
		requireKind(t, err, domain.KindNotFound)
	})

	t.Run("пока сервер не куплен", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		_, _, err := svc.GetServerData(context.Background(), ev.id, ev.creator)
		requireKind(t, err, domain.KindFailedPrecondition)
	})

	t.Run("участник получает реквизиты", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		player := joinPlayer(t, svc, ev.id, newID(), true)
		confirmEvent(t, svc, repo, ev)
		serverIP := newServerIP()
		requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, serverIP, "secret"))

		for _, userID := range []string{ev.creator, ev.enemy, player} {
			gotID, password, err := svc.GetServerData(context.Background(), ev.id, userID)
			requireNoErr(t, err)

			if gotID != serverIP || password != "secret" {
				t.Errorf("участник %s получил %q / %q", userID, gotID, password)
			}
		}
	})

	t.Run("после отмены реквизиты не отдаются", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		confirmEvent(t, svc, repo, ev)
		requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, newID(), "secret"))
		requireNoErr(t, svc.CancelEvent(context.Background(), ev.id, ev.creator))

		_, _, err := svc.GetServerData(context.Background(), ev.id, ev.creator)
		requireKind(t, err, domain.KindFailedPrecondition)
	})

	t.Run("сбой репозитория", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		confirmEvent(t, svc, repo, ev)
		requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, newID(), "secret"))

		repo.fail("GetUserByID", errDB)
		_, _, err := svc.GetServerData(context.Background(), ev.id, ev.creator)
		requireInternal(t, err)
	})
}

func TestStartEventFailsOnBrokenStart(t *testing.T) {
	cases := []struct {
		name    string
		prepare func(t *testing.T, svc *eventService, repo *fakeRepo, producer *fakeProducer, ev testEvent)
		kind    *domain.ErrorKind
	}{
		{
			name: "не удалось сменить статус",
			prepare: func(t *testing.T, svc *eventService, repo *fakeRepo, producer *fakeProducer, ev testEvent) {
				repo.fail("UpdateEventStatus", errDB)
			},
		},
		{
			name: "не удалось прочитать команды",
			prepare: func(t *testing.T, svc *eventService, repo *fakeRepo, producer *fakeProducer, ev testEvent) {
				repo.fail("GetTeamsByEventID", errDB)
			},
		},
		{
			name: "не удалось стартовать первую игру",
			prepare: func(t *testing.T, svc *eventService, repo *fakeRepo, producer *fakeProducer, ev testEvent) {
				repo.fail("StartTeamGame", errDB)
			},
		},
		{
			name: "команд первой игры нет",
			prepare: func(t *testing.T, svc *eventService, repo *fakeRepo, producer *fakeProducer, ev testEvent) {
				repo.mu.Lock()
				defer repo.mu.Unlock()

				for teamID, team := range repo.teams {
					if team.EventID == ev.id {
						delete(repo.teams, teamID)
					}
				}
			},
		},
		{
			name: "не ушло сообщение о старте игры",
			prepare: func(t *testing.T, svc *eventService, repo *fakeRepo, producer *fakeProducer, ev testEvent) {
				producer.fail("team_game.started", errKafka)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			svc, repo, producer := newTestService(t)
			ev := createTestEvent(t, svc, 1)
			confirmEvent(t, svc, repo, ev)
			deployServer(t, svc, repo, ev)

			requireNoErr(t, svc.StartEvent(context.Background(), ev.id, ev.creator))
			c.prepare(t, svc, repo, producer, ev)

			if err := svc.StartEvent(context.Background(), ev.id, ev.enemy); err == nil {
				t.Fatal("ожидалась ошибка старта")
			}
		})
	}
}

func TestStartEventRollsBackOnFailedStart(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	confirmEvent(t, svc, repo, ev)
	deployServer(t, svc, repo, ev)

	requireNoErr(t, svc.StartEvent(context.Background(), ev.id, ev.creator))
	repo.fail("StartTeamGame", errDB)

	requireInternal(t, svc.StartEvent(context.Background(), ev.id, ev.enemy))

	// Нажатие второй стороны и статус откатились вместе с игрой: иначе ивент
	// остался бы in_progress без запущенной игры.
	repo.mu.Lock()
	event := repo.events[ev.id]
	repo.mu.Unlock()

	if event.Status != domain.EventStatusReady {
		t.Errorf("статус %s, ожидался ready после отката", event.Status)
	}
	if !event.EnemyReadyAt.IsZero() {
		t.Error("нажатие второй стороны должно было откатиться вместе с транзакцией")
	}
}

func TestServerDeployedFailsOnRepositoryErrors(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	confirmEvent(t, svc, repo, ev)

	// С момента создания у ивента уже стоит своя цепочка таймеров: при
	// неудачном деплое она не должна подмениться цепочкой ожидания StartEvent.
	before := svc.eventTimers[ev.id]

	serverIP := newServerIP()
	requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, serverIP, "secret"))

	repo.fail("SetServerDeployedAt", errDB)
	requireInternal(t, svc.ServerDeployed(context.Background(), ev.id, serverIP))

	if svc.eventTimers[ev.id] != before {
		t.Error("при неудачном деплое цепочка таймеров не должна переставляться")
	}
}

func TestServerPurchasedFailsOnRepositoryError(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	confirmEvent(t, svc, repo, ev)

	repo.fail("SetServerData", errDB)
	requireInternal(t, svc.ServerPurchased(context.Background(), ev.id, newID(), "pass"))
}

// expectStatus — статус ивента в базе.
func expectStatus(t *testing.T, repo *fakeRepo, eventID string, want domain.EventStatus) {
	t.Helper()

	if got := eventStatus(repo, eventID); got != want {
		t.Errorf("статус %s, ожидался %s", got, want)
	}
}

// expectStartError — StartEvent отказывает, и в тексте видно, чего не хватает.
func expectStartError(t *testing.T, svc *eventService, ev testEvent, want string) {
	t.Helper()

	err := svc.StartEvent(context.Background(), ev.id, ev.creator)
	requireKind(t, err, domain.KindFailedPrecondition)

	if !strings.Contains(err.Error(), want) {
		t.Errorf("ошибка %q не содержит %q", err.Error(), want)
	}
}

func TestServerDeployedFailsWhenPairLookupBroken(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	confirmEvent(t, svc, repo, ev)
	serverIP := newServerIP()
	requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, serverIP, "secret"))

	repo.fail("GetEventByIDAndServerIPForUpdate", errDB)
	requireInternal(t, svc.ServerDeployed(context.Background(), ev.id, serverIP))
}

func TestServerDeployedFailsWhenStatusUpdateBroken(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	confirmEvent(t, svc, repo, ev)
	serverIP := newServerIP()
	requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, serverIP, "secret"))

	repo.fail("UpdateEventStatus", errDB)
	requireInternal(t, svc.ServerDeployed(context.Background(), ev.id, serverIP))

	// Откат: момент готовности не записан, статус не сдвинулся.
	repo.unfail("UpdateEventStatus")
	repo.mu.Lock()
	event := repo.events[ev.id]
	repo.mu.Unlock()

	if !event.ServerDeployedAt.IsZero() {
		t.Error("момент готовности должен был откатиться вместе со статусом")
	}
	expectStatus(t, repo, ev.id, domain.EventStatusPurchased)
}

func TestServerPurchasedFailsWhenStatusUpdateBroken(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	confirmEvent(t, svc, repo, ev)

	repo.fail("UpdateEventStatus", errDB)
	requireInternal(t, svc.ServerPurchased(context.Background(), ev.id, newID(), "secret"))

	// Реквизиты откатились вместе со статусом.
	repo.unfail("UpdateEventStatus")
	repo.mu.Lock()
	event := repo.events[ev.id]
	repo.mu.Unlock()

	if event.ServerIP != "" {
		t.Errorf("реквизиты сервера остались после отката: %q", event.ServerIP)
	}
	expectStatus(t, repo, ev.id, domain.EventStatusConfirmed)
}

func TestServerDeployedRejectsAfterCancel(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	confirmEvent(t, svc, repo, ev)
	serverIP := newServerIP()
	requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, serverIP, "secret"))
	requireNoErr(t, svc.CancelEvent(context.Background(), ev.id, ev.creator))

	// Ивент отменён, пока сервер разворачивался.
	requireKind(t, svc.ServerDeployed(context.Background(), ev.id, serverIP), domain.KindFailedPrecondition)
	expectStatus(t, repo, ev.id, domain.EventStatusCanceled)
}

func TestCancelEventAllowedUntilStart(t *testing.T) {
	for _, status := range []domain.EventStatus{
		domain.EventStatusPending,
		domain.EventStatusConfirmed,
		domain.EventStatusPurchased,
		domain.EventStatusDeployed,
		domain.EventStatusReady,
	} {
		t.Run(string(status), func(t *testing.T) {
			svc, repo, _ := newTestService(t)
			ev := createTestEvent(t, svc, 1)

			setEventFields(repo, ev.id, func(event *domain.Event) {
				event.Status = status
			})

			requireNoErr(t, svc.CancelEvent(context.Background(), ev.id, ev.creator))
			expectStatus(t, repo, ev.id, domain.EventStatusCanceled)
		})
	}
}

func TestServerPurchasedAcceptsAnyAddressForm(t *testing.T) {
	for _, address := range []string{"1.2.3.4", "10.0.0.1:27015", "2001:db8::1", "srv-01.example.com"} {
		t.Run(address, func(t *testing.T) {
			svc, repo, _ := newTestService(t)
			ev := createTestEvent(t, svc, 1)
			confirmEvent(t, svc, repo, ev)

			requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, address, "secret"))

			gotIP, _, err := svc.GetServerData(context.Background(), ev.id, ev.creator)
			requireNoErr(t, err)

			if gotIP != address {
				t.Errorf("адрес сервера %q, ожидался %q", gotIP, address)
			}
		})
	}
}

func TestServerIPTrimmedOnBothMessages(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	confirmEvent(t, svc, repo, ev)

	// Пробелы по краям не должны мешать сверке пары в vps.deployed.
	requireNoErr(t, svc.ServerPurchased(context.Background(), ev.id, "  10.1.2.3  ", "secret"))
	requireNoErr(t, svc.ServerDeployed(context.Background(), ev.id, " 10.1.2.3 "))

	expectStatus(t, repo, ev.id, domain.EventStatusDeployed)

	gotIP, _, err := svc.GetServerData(context.Background(), ev.id, ev.creator)
	requireNoErr(t, err)
	if gotIP != "10.1.2.3" {
		t.Errorf("адрес сервера %q, ожидался обрезанный по краям", gotIP)
	}
}
