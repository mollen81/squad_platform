package service

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	domain "event-service/internal/core/domain"
)

func TestCreateEventRejectsBadInput(t *testing.T) {
	creator, creatorClan := newID(), newID()
	enemy, enemyClan := newID(), newID()

	cases := []struct {
		name            string
		creator         string
		creatorClan     string
		enemy           string
		enemyClan       string
		eventName       string
		timeStart       time.Time
		targetGameCount int64
	}{
		{"пустой создатель", "", creatorClan, enemy, enemyClan, "name", futureStart(), 1},
		{"создатель не uuid", "не-uuid", creatorClan, enemy, enemyClan, "name", futureStart(), 1},
		{"пустой клан создателя", creator, "", enemy, enemyClan, "name", futureStart(), 1},
		{"пустой лидер второй стороны", creator, creatorClan, "", enemyClan, "name", futureStart(), 1},
		{"клан второй стороны не uuid", creator, creatorClan, enemy, "clan", "name", futureStart(), 1},
		{"создатель сам себе противник", creator, creatorClan, creator, enemyClan, "name", futureStart(), 1},
		{"пустое название", creator, creatorClan, enemy, enemyClan, "   ", futureStart(), 1},
		{"слишком длинное название", creator, creatorClan, enemy, enemyClan, strings.Repeat("я", maxEventNameLength+1), futureStart(), 1},
		{"старт слишком скоро", creator, creatorClan, enemy, enemyClan, "name", time.Now().Add(10 * time.Minute), 1},
		{"старт слишком далеко", creator, creatorClan, enemy, enemyClan, "name", time.Now().Add(50 * 24 * time.Hour), 1},
		{"игр меньше одной", creator, creatorClan, enemy, enemyClan, "name", futureStart(), 0},
		{"игр больше трёх", creator, creatorClan, enemy, enemyClan, "name", futureStart(), 4},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			svc, repo, producer := newTestService(t)

			err := svc.CreateEvent(context.Background(), c.creator, c.creatorClan, c.enemy, c.enemyClan, c.eventName, c.timeStart, c.targetGameCount)
			requireKind(t, err, domain.KindInvalidArgument)

			if len(repo.events) != 0 {
				t.Errorf("при отказе не должно создаваться ивентов, создано %d", len(repo.events))
			}
			if len(producer.kinds()) != 0 {
				t.Errorf("при отказе в Kafka ничего не должно уходить, ушло %v", producer.kinds())
			}
		})
	}
}

func TestCreateEventBuildsFullEvent(t *testing.T) {
	svc, repo, producer := newTestService(t)

	ev := createTestEvent(t, svc, 3)

	event := repo.events[ev.id]
	if event.Status != domain.EventStatusPending {
		t.Errorf("новый ивент должен быть pending, получен %s", event.Status)
	}
	if event.UserCount != 2 {
		t.Errorf("в ивенте должны быть оба сайд-лидера, user_count=%d", event.UserCount)
	}
	if event.TargetGameCount != 3 {
		t.Errorf("target_game_count=%d, ожидалось 3", event.TargetGameCount)
	}

	teams, err := svc.GetTeamsByEventID(context.Background(), ev.id)
	requireNoErr(t, err)
	if len(teams) != 6 {
		t.Fatalf("на 3 игры должно быть 6 команд, создано %d", len(teams))
	}

	// Сайд-лидер обязан играть каждую игру: он уже в составе всех своих
	// команд и с ролью squad_leader.
	for _, team := range teams {
		if team.Status != domain.TeamStatusPending {
			t.Errorf("команда %s: статус %s, ожидался pending", team.TeamID, team.Status)
		}
		if team.MembersCount != 1 {
			t.Errorf("команда %s: members_count=%d, ожидался сайд-лидер в составе", team.TeamID, team.MembersCount)
		}

		member, ok := repo.members[team.TeamID][team.SideLeaderID]
		if !ok {
			t.Errorf("команда %s: сайд-лидера нет в составе", team.TeamID)
			continue
		}
		if member.Role != domain.RoleSquadLeader {
			t.Errorf("команда %s: роль сайд-лидера %s, ожидалась squad_leader", team.TeamID, member.Role)
		}
	}

	want := []string{"event.created", "user.joined_event", "user.joined_event"}
	if got := producer.kinds(); !reflect.DeepEqual(got, want) {
		t.Errorf("в Kafka ушло %v, ожидалось %v", got, want)
	}

	// Цепочка таймеров ставится только после успешного создания.
	if _, ok := svc.eventTimers[ev.id]; !ok {
		t.Error("для созданного ивента не поставлена цепочка таймеров")
	}
}

func TestCreateEventRollsBackOnFailure(t *testing.T) {
	svc, repo, producer := newTestService(t)
	repo.fail("CreateTeam", errors.New("база отказала на создании команды"))

	err := svc.CreateEvent(context.Background(), newID(), newID(), newID(), newID(), "name", futureStart(), 2)
	if err == nil {
		t.Fatal("ожидалась ошибка создания")
	}

	// Ивент, участники и команды создаются одной транзакцией: не должно
	// остаться ни "полуивента", ни таймера, ни сообщений в Kafka.
	if len(repo.events) != 0 || len(repo.users) != 0 || len(repo.teams) != 0 {
		t.Errorf("после отката осталось: ивентов %d, участников %d, команд %d", len(repo.events), len(repo.users), len(repo.teams))
	}
	if len(producer.kinds()) != 0 {
		t.Errorf("после отката в Kafka ушло %v", producer.kinds())
	}
	if len(svc.eventTimers) != 0 {
		t.Errorf("после отката осталось %d цепочек таймеров", len(svc.eventTimers))
	}
}

func TestJoinToEvent(t *testing.T) {
	t.Run("невалидные данные", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		requireKind(t, svc.JoinToEvent(context.Background(), "не-uuid", newID(), newID(), false), domain.KindInvalidArgument)
		requireKind(t, svc.JoinToEvent(context.Background(), ev.id, "", newID(), false), domain.KindInvalidArgument)
		requireKind(t, svc.JoinToEvent(context.Background(), ev.id, newID(), "", false), domain.KindInvalidArgument)
	})

	t.Run("несуществующий ивент", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		requireKind(t, svc.JoinToEvent(context.Background(), newID(), newID(), newID(), false), domain.KindNotFound)
	})

	t.Run("успешный вход считается в user_count", func(t *testing.T) {
		svc, repo, producer := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		userID := joinPlayer(t, svc, ev.id, newID(), false)

		if got := repo.events[ev.id].UserCount; got != 3 {
			t.Errorf("user_count=%d, ожидалось 3", got)
		}
		if producer.count("user.joined_event") != 3 {
			t.Errorf("должно быть 3 сообщения о входе, получено %d", producer.count("user.joined_event"))
		}

		user, err := svc.eventRepo.GetUserByID(context.Background(), ev.id, userID)
		requireNoErr(t, err)
		if user.Enemy {
			t.Error("игрок вошёл за свою сторону, а записан как enemy")
		}
	})

	t.Run("повторный вход того же игрока", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		userID := joinPlayer(t, svc, ev.id, newID(), false)
		err := svc.JoinToEvent(context.Background(), ev.id, userID, newID(), false)
		requireKind(t, err, domain.KindAlreadyExists)

		if got := repo.events[ev.id].UserCount; got != 3 {
			t.Errorf("отказ не должен менять user_count, получено %d", got)
		}
	})

	t.Run("после подтверждения состав заморожен", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		event := repo.events[ev.id]
		event.Status = domain.EventStatusConfirmed
		repo.events[ev.id] = event

		requireKind(t, svc.JoinToEvent(context.Background(), ev.id, newID(), newID(), false), domain.KindFailedPrecondition)
	})
}

func TestLeaveEvent(t *testing.T) {
	t.Run("сайд-лидеры выйти не могут", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		requireKind(t, svc.LeaveEvent(context.Background(), ev.creator, ev.id), domain.KindPermissionDenied)
		requireKind(t, svc.LeaveEvent(context.Background(), ev.enemy, ev.id), domain.KindPermissionDenied)
	})

	t.Run("не участник ивента", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		requireKind(t, svc.LeaveEvent(context.Background(), newID(), ev.id), domain.KindNotFound)
	})

	t.Run("после подтверждения выйти нельзя", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)

		event := repo.events[ev.id]
		event.Status = domain.EventStatusConfirmed
		repo.events[ev.id] = event

		requireKind(t, svc.LeaveEvent(context.Background(), userID, ev.id), domain.KindFailedPrecondition)
	})

	t.Run("выход убирает игрока из всех команд", func(t *testing.T) {
		svc, repo, producer := newTestService(t)
		ev := createTestEvent(t, svc, 2)
		userID := joinPlayer(t, svc, ev.id, newID(), false)

		ally1, _ := ev.teams(t, svc, 1)
		ally2, _ := ev.teams(t, svc, 2)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally1, userID, domain.RolePlayer))
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally2, userID, domain.RolePlayer))

		requireNoErr(t, svc.LeaveEvent(context.Background(), userID, ev.id))

		if got := repo.events[ev.id].UserCount; got != 2 {
			t.Errorf("user_count=%d, ожидалось 2", got)
		}
		for _, teamID := range []string{ally1, ally2} {
			if got := repo.teams[teamID].MembersCount; got != 1 {
				t.Errorf("команда %s: members_count=%d, ожидался только сайд-лидер", teamID, got)
			}
			if len(repo.members[teamID]) != 1 {
				t.Errorf("команда %s: в составе осталось %d строк", teamID, len(repo.members[teamID]))
			}
		}
		if producer.count("user.left_event") != 1 {
			t.Error("не опубликован выход из ивента")
		}
	})
}

func TestUpdateTimeEvent(t *testing.T) {
	t.Run("менять время может только создатель", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		err := svc.UpdateTimeEvent(context.Background(), ev.id, newID(), futureStart())
		requireKind(t, err, domain.KindPermissionDenied)
	})

	t.Run("новое время проверяется", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		requireKind(t, svc.UpdateTimeEvent(context.Background(), ev.id, ev.creator, time.Now().Add(time.Minute)), domain.KindInvalidArgument)
		requireKind(t, svc.UpdateTimeEvent(context.Background(), ev.id, ev.creator, time.Now().Add(100*24*time.Hour)), domain.KindInvalidArgument)
	})

	t.Run("после подтверждения время не меняется", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		event := repo.events[ev.id]
		event.Status = domain.EventStatusConfirmed
		repo.events[ev.id] = event

		requireKind(t, svc.UpdateTimeEvent(context.Background(), ev.id, ev.creator, futureStart()), domain.KindFailedPrecondition)
	})

	t.Run("успешный перенос переставляет таймер", func(t *testing.T) {
		svc, repo, producer := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		oldTimer := svc.eventTimers[ev.id]
		newTime := time.Now().Add(3 * time.Hour)
		requireNoErr(t, svc.UpdateTimeEvent(context.Background(), ev.id, ev.creator, newTime))

		if got := repo.events[ev.id].TimeStart; !got.Equal(newTime) {
			t.Errorf("время старта не обновилось: %v", got)
		}
		if svc.eventTimers[ev.id] == oldTimer {
			t.Error("цепочка таймеров не переставлена на новое время")
		}
		if producer.count("event.time_updated") != 1 {
			t.Error("не опубликован перенос времени")
		}
	})
}

func TestCancelEvent(t *testing.T) {
	t.Run("отменить может только создатель", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		requireKind(t, svc.CancelEvent(context.Background(), ev.id, newID()), domain.KindPermissionDenied)
	})

	t.Run("несуществующий ивент", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		requireKind(t, svc.CancelEvent(context.Background(), newID(), newID()), domain.KindNotFound)
	})

	t.Run("отмена гасит таймеры", func(t *testing.T) {
		svc, repo, producer := newTestService(t)
		ev := createTestEvent(t, svc, 1)

		requireNoErr(t, svc.CancelEvent(context.Background(), ev.id, ev.creator))

		if got := repo.events[ev.id].Status; got != domain.EventStatusCanceled {
			t.Errorf("статус %s, ожидался canceled", got)
		}
		if _, ok := svc.eventTimers[ev.id]; ok {
			t.Error("после отмены цепочка таймеров должна быть снята")
		}
		if producer.count("event.canceled") != 1 {
			t.Error("не опубликована отмена")
		}
	})

	t.Run("уже начатый ивент не отменяется", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		startEvent(t, svc, repo, ev)

		requireKind(t, svc.CancelEvent(context.Background(), ev.id, ev.creator), domain.KindFailedPrecondition)
	})
}

func TestGettersValidateInput(t *testing.T) {
	svc, _, _ := newTestService(t)

	_, err := svc.GetEventsByCreatorId(context.Background(), "не-uuid")
	requireKind(t, err, domain.KindInvalidArgument)

	_, err = svc.GetLastEventByCreatorId(context.Background(), "")
	requireKind(t, err, domain.KindInvalidArgument)

	_, err = svc.GetUnfinishedEventsByUserID(context.Background(), "не-uuid")
	requireKind(t, err, domain.KindInvalidArgument)

	_, err = svc.GetEventsByEventName(context.Background(), "  ")
	requireKind(t, err, domain.KindInvalidArgument)

	_, err = svc.GetUnfinishedEventsByEventName(context.Background(), "")
	requireKind(t, err, domain.KindInvalidArgument)

	_, err = svc.GetEventMembersList(context.Background(), "не-uuid")
	requireKind(t, err, domain.KindInvalidArgument)

	_, err = svc.GetTeamsByEventID(context.Background(), "не-uuid")
	requireKind(t, err, domain.KindInvalidArgument)

	_, err = svc.GetTeamByID(context.Background(), "не-uuid")
	requireKind(t, err, domain.KindInvalidArgument)

	_, _, err = svc.GetTeamStats(context.Background(), "не-uuid")
	requireKind(t, err, domain.KindInvalidArgument)
}

func TestGettersReturnData(t *testing.T) {
	svc, _, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)

	events, err := svc.GetEventsByCreatorId(context.Background(), ev.creator)
	requireNoErr(t, err)
	if len(events) != 1 {
		t.Errorf("GetEventsByCreatorId вернул %d ивентов, ожидался 1", len(events))
	}

	byName, err := svc.GetEventsByEventName(context.Background(), "test event")
	requireNoErr(t, err)
	if len(byName) != 1 {
		t.Errorf("GetEventsByEventName вернул %d ивентов, ожидался 1", len(byName))
	}

	unfinished, err := svc.GetUnfinishedEventsByUserID(context.Background(), ev.creator)
	requireNoErr(t, err)
	if len(unfinished) != 1 {
		t.Errorf("GetUnfinishedEventsByUserID вернул %d ивентов, ожидался 1", len(unfinished))
	}

	members, err := svc.GetEventMembersList(context.Background(), ev.id)
	requireNoErr(t, err)
	if len(members) != 2 {
		t.Errorf("в ивенте %d участников, ожидалось 2", len(members))
	}

	_, err = svc.GetTeamByID(context.Background(), newID())
	requireKind(t, err, domain.KindNotFound)
}

func TestGettersPropagateRepositoryFailures(t *testing.T) {
	cases := []struct {
		name   string
		method string
		call   func(svc *eventService) error
	}{
		{"GetEventsByCreatorId", "GetEventsByCreatorId", func(svc *eventService) error {
			_, err := svc.GetEventsByCreatorId(context.Background(), newID())
			return err
		}},
		{"GetLastEventByCreatorId", "GetLastEventByCreatorId", func(svc *eventService) error {
			_, err := svc.GetLastEventByCreatorId(context.Background(), newID())
			return err
		}},
		{"GetEventsByEventName", "GetEventsByEventName", func(svc *eventService) error {
			_, err := svc.GetEventsByEventName(context.Background(), "турнир")
			return err
		}},
		{"GetUnfinishedEventsByUserID", "GetUnfinishedEventsByUserID", func(svc *eventService) error {
			_, err := svc.GetUnfinishedEventsByUserID(context.Background(), newID())
			return err
		}},
		{"GetUnfinishedEventsByEventName", "GetUnfinishedEventsByEventName", func(svc *eventService) error {
			_, err := svc.GetUnfinishedEventsByEventName(context.Background(), "турнир")
			return err
		}},
		{"GetEventMembersList", "GetEventMembersList", func(svc *eventService) error {
			_, err := svc.GetEventMembersList(context.Background(), newID())
			return err
		}},
		{"GetTeamsByEventID", "GetTeamsByEventID", func(svc *eventService) error {
			_, err := svc.GetTeamsByEventID(context.Background(), newID())
			return err
		}},
		{"GetTeamByID", "GetTeamByID", func(svc *eventService) error {
			_, err := svc.GetTeamByID(context.Background(), newID())
			return err
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			svc, repo, _ := newTestService(t)
			repo.failOn[c.method] = errDB

			requireInternal(t, c.call(svc))
		})
	}
}

func TestGetUnfinishedEventsByEventNameReturnsOnlyActive(t *testing.T) {
	svc, repo, _ := newTestService(t)

	active := createTestEvent(t, svc, 1)
	canceled := createTestEvent(t, svc, 1)
	requireNoErr(t, svc.CancelEvent(context.Background(), canceled.id, canceled.creator))

	events, err := svc.GetUnfinishedEventsByEventName(context.Background(), "test event")
	requireNoErr(t, err)

	if len(events) != 1 {
		t.Fatalf("вернулось %d ивентов, ожидался только активный", len(events))
	}
	if events[0].EventID != active.id {
		t.Errorf("вернулся ивент %q, ожидался %q", events[0].EventID, active.id)
	}
	if repo.events[canceled.id].Status != domain.EventStatusCanceled {
		t.Error("отменённый ивент должен остаться в базе, просто не попадать в выдачу")
	}
}

func TestCreateEventPublishFailureOnSecondMessage(t *testing.T) {
	svc, repo, producer := newTestService(t)
	producer.fail("user.joined_event", errors.New("kafka недоступна"))

	err := svc.CreateEvent(context.Background(), newID(), newID(), newID(), newID(), "name", futureStart(), 1)
	if err == nil {
		t.Fatal("ошибка публикации должна возвращаться вызывающему")
	}

	// Ивент уже закоммичен и таймер поставлен — потеряться он не должен.
	if len(repo.events) != 1 {
		t.Errorf("ивентов в базе %d, ожидался 1", len(repo.events))
	}
	if len(svc.eventTimers) != 1 {
		t.Errorf("цепочек таймеров %d, ожидалась 1", len(svc.eventTimers))
	}
}

func TestMutatingCallsValidateIdentifiers(t *testing.T) {
	svc, _, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	ally, enemy := ev.teams(t, svc, 1)

	cases := []struct {
		name string
		err  error
	}{
		{"UpdateTimeEvent: event_id", svc.UpdateTimeEvent(context.Background(), "не-uuid", ev.creator, futureStart())},
		{"UpdateTimeEvent: user_create_id", svc.UpdateTimeEvent(context.Background(), ev.id, "не-uuid", futureStart())},
		{"CancelEvent: event_id", svc.CancelEvent(context.Background(), "", ev.creator)},
		{"CancelEvent: user_create_id", svc.CancelEvent(context.Background(), ev.id, "не-uuid")},
		{"LeaveEvent: event_id", svc.LeaveEvent(context.Background(), ev.creator, "не-uuid")},
		{"LeaveEvent: user_id", svc.LeaveEvent(context.Background(), "не-uuid", ev.id)},
		{"SetRole: user_id", svc.SetRole(context.Background(), ally, ev.creator, "не-uuid", domain.RolePlayer)},
		{"FinishTeamGame: team2_id", svc.FinishTeamGame(context.Background(), ally, "не-uuid", ally)},
		{"FinishTeamGame: team_winner_id", svc.FinishTeamGame(context.Background(), ally, enemy, "не-uuid")},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			requireKind(t, c.err, domain.KindInvalidArgument)
		})
	}
}

func TestCreateEventRollsBackWhenSideLeaderCannotJoinTeam(t *testing.T) {
	svc, repo, producer := newTestService(t)
	repo.fail("JoinUserToTeam", errDB)

	requireInternal(t, svc.CreateEvent(context.Background(), newID(), newID(), newID(), newID(), "name", futureStart(), 1))

	if len(repo.events) != 0 || len(repo.teams) != 0 || len(repo.users) != 0 {
		t.Errorf("после отката осталось: ивентов %d, команд %d, участников %d", len(repo.events), len(repo.teams), len(repo.users))
	}
	if len(producer.kinds()) != 0 {
		t.Errorf("после отката в Kafka ушло %v", producer.kinds())
	}
}

// Публикация идёт после коммита: если Kafka отказала, изменение в БД уже
// сделано — эти тесты фиксируют такое поведение, чтобы оно менялось осознанно.
func TestPublishFailureAfterCommit(t *testing.T) {
	t.Run("CreateEvent", func(t *testing.T) {
		svc, repo, producer := newTestService(t)
		producer.fail("event.created", errors.New("kafka недоступна"))

		err := svc.CreateEvent(context.Background(), newID(), newID(), newID(), newID(), "name", futureStart(), 1)
		if err == nil {
			t.Fatal("ошибка публикации должна возвращаться вызывающему")
		}
		if len(repo.events) != 1 {
			t.Errorf("ивент уже закоммичен, в базе должно быть 1, есть %d", len(repo.events))
		}
	})

	t.Run("JoinToEvent", func(t *testing.T) {
		svc, repo, producer := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		producer.fail("user.joined_event", errors.New("kafka недоступна"))

		if err := svc.JoinToEvent(context.Background(), ev.id, newID(), newID(), false); err == nil {
			t.Fatal("ошибка публикации должна возвращаться вызывающему")
		}
		if got := repo.events[ev.id].UserCount; got != 3 {
			t.Errorf("вход уже закоммичен, user_count=%d, ожидалось 3", got)
		}
	})
}

func TestNewEventServiceBuildsUsableService(t *testing.T) {
	svc := NewEventService(newFakeRepo(), newFakeProducer())
	if svc == nil {
		t.Fatal("NewEventService вернул nil")
	}

	// Сервис сразу готов к работе: проверки входа не требуют ни БД, ни Kafka.
	requireKind(t, svc.JoinToEvent(context.Background(), "не-uuid", newID(), newID(), false), domain.KindInvalidArgument)
}

// Сбой репозитория не должен превращаться в доменную ошибку: категории нет,
// значит транспорт отдаст Internal, а не "не найдено" или "нельзя".
func TestEventCallsFailOnRepositoryErrors(t *testing.T) {
	t.Run("CreateEvent", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		repo.fail("CreateEvent", errDB)

		requireInternal(t, svc.CreateEvent(context.Background(), newID(), newID(), newID(), newID(), "name", futureStart(), 1))
	})

	t.Run("JoinToEvent: проверка участия", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		repo.fail("IsUserInEvent", errDB)

		requireInternal(t, svc.JoinToEvent(context.Background(), ev.id, newID(), newID(), false))
	})

	t.Run("JoinToEvent: вставка участника", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		repo.fail("JoinToEvent", errDB)

		requireInternal(t, svc.JoinToEvent(context.Background(), ev.id, newID(), newID(), false))
	})

	t.Run("чтение ивента под блокировкой", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		repo.fail("GetEventByIDForUpdate", errDB)

		requireInternal(t, svc.CancelEvent(context.Background(), ev.id, ev.creator))
	})

	t.Run("UpdateTimeEvent", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		repo.fail("UpdateTimeEvent", errDB)

		requireInternal(t, svc.UpdateTimeEvent(context.Background(), ev.id, ev.creator, futureStart()))
	})

	t.Run("CancelEvent", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		repo.fail("UpdateEventStatus", errDB)

		requireInternal(t, svc.CancelEvent(context.Background(), ev.id, ev.creator))
		if _, ok := svc.eventTimers[ev.id]; !ok {
			t.Error("неудачная отмена не должна гасить таймеры ивента")
		}
	})

	t.Run("LeaveEvent: удаление участника", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		repo.fail("LeaveEvent", errDB)

		requireInternal(t, svc.LeaveEvent(context.Background(), userID, ev.id))
	})

	t.Run("LeaveEvent: снятие из составов", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		repo.fail("RemoveUserFromAllTeams", errDB)

		requireInternal(t, svc.LeaveEvent(context.Background(), userID, ev.id))

		// Транзакция откатилась целиком: игрок остался в ивенте.
		if got := repo.events[ev.id].UserCount; got != 3 {
			t.Errorf("user_count=%d, ожидалось 3 после отката", got)
		}
	})

	t.Run("LeaveEvent: пересчёт однокланников", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))
		repo.fail("CountClanMembersInTeam", errDB)

		requireInternal(t, svc.LeaveEvent(context.Background(), userID, ev.id))
	})
}

func TestEventPublishFailuresAreReported(t *testing.T) {
	t.Run("LeaveEvent", func(t *testing.T) {
		svc, _, producer := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		producer.fail("user.left_event", errKafka)

		if err := svc.LeaveEvent(context.Background(), userID, ev.id); err == nil {
			t.Error("ошибка публикации должна возвращаться")
		}
	})

	t.Run("CancelEvent", func(t *testing.T) {
		svc, _, producer := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		producer.fail("event.canceled", errKafka)

		if err := svc.CancelEvent(context.Background(), ev.id, ev.creator); err == nil {
			t.Error("ошибка публикации должна возвращаться")
		}
	})

	t.Run("UpdateTimeEvent", func(t *testing.T) {
		svc, _, producer := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		producer.fail("event.time_updated", errKafka)

		if err := svc.UpdateTimeEvent(context.Background(), ev.id, ev.creator, futureStart()); err == nil {
			t.Error("ошибка публикации должна возвращаться")
		}
	})
}
