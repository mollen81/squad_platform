package service

import (
	"context"
	"testing"

	domain "event-service/internal/core/domain"
)

func TestStartTeamGameValidatesInput(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 2)
	startEvent(t, svc, repo, ev)

	ally1, enemy1 := ev.teams(t, svc, 1)
	ally2, _ := ev.teams(t, svc, 2)

	requireKind(t, svc.StartTeamGame(context.Background(), "не-uuid", enemy1), domain.KindInvalidArgument)
	requireKind(t, svc.StartTeamGame(context.Background(), ally1, ""), domain.KindInvalidArgument)
	requireKind(t, svc.StartTeamGame(context.Background(), ally1, ally1), domain.KindInvalidArgument)

	// команды из разных игр
	requireKind(t, svc.StartTeamGame(context.Background(), ally2, enemy1), domain.KindInvalidArgument)

	// несуществующая команда
	requireKind(t, svc.StartTeamGame(context.Background(), newID(), enemy1), domain.KindNotFound)
}

func TestStartTeamGameRejectsTeamsFromDifferentEvents(t *testing.T) {
	svc, repo, _ := newTestService(t)

	first := createTestEvent(t, svc, 1)
	second := createTestEvent(t, svc, 1)
	startEvent(t, svc, repo, first)

	firstAlly, _ := first.teams(t, svc, 1)
	secondEnemy := ""
	_, secondEnemy = second.teams(t, svc, 1)

	// Игра идёт, поэтому команда первого ивента уже in_progress — проверяем,
	// что чужая команда отсекается независимо от этого.
	err := svc.StartTeamGame(context.Background(), firstAlly, secondEnemy)
	if err == nil {
		t.Fatal("команды из разных ивентов не должны стартовать вместе")
	}
}

func TestStartTeamGameRequiresRunningEvent(t *testing.T) {
	svc, _, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	ally, enemy := ev.teams(t, svc, 1)

	// Ивент ещё pending — игры стартуют только после его старта.
	requireKind(t, svc.StartTeamGame(context.Background(), ally, enemy), domain.KindFailedPrecondition)
}

func TestStartTeamGameRequiresPreviousGameFinished(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 2)
	startEvent(t, svc, repo, ev)

	ally2, enemy2 := ev.teams(t, svc, 2)

	// Первая игра только началась — вторую начинать рано.
	requireKind(t, svc.StartTeamGame(context.Background(), ally2, enemy2), domain.KindFailedPrecondition)

	ally1, enemy1 := ev.teams(t, svc, 1)
	requireNoErr(t, svc.FinishTeamGame(context.Background(), ally1, enemy1, ally1))
	requireNoErr(t, svc.StartTeamGame(context.Background(), ally2, enemy2))

	if got := repo.teams[ally2].Status; got != domain.TeamStatusInProgress {
		t.Errorf("статус команды второй игры %s, ожидался in_progress", got)
	}
}

func TestStartTeamGameRejectsSecondStart(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 2)
	startEvent(t, svc, repo, ev)

	ally1, enemy1 := ev.teams(t, svc, 1)
	requireKind(t, svc.StartTeamGame(context.Background(), ally1, enemy1), domain.KindFailedPrecondition)

	// Первая игра стартует вместе с ивентом: два сообщения о старте команд.
	if got := producer.count("team_game.started"); got != 2 {
		t.Errorf("сообщений о старте команд %d, ожидалось 2", got)
	}
}

func TestFinishTeamGameValidatesInput(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	startEvent(t, svc, repo, ev)
	ally, enemy := ev.teams(t, svc, 1)

	requireKind(t, svc.FinishTeamGame(context.Background(), "не-uuid", enemy, enemy), domain.KindInvalidArgument)
	requireKind(t, svc.FinishTeamGame(context.Background(), ally, ally, ally), domain.KindInvalidArgument)

	// победителем может быть только одна из этих двух команд
	requireKind(t, svc.FinishTeamGame(context.Background(), ally, enemy, newID()), domain.KindInvalidArgument)
}

func TestFinishTeamGameRequiresStartedGame(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 2)
	startEvent(t, svc, repo, ev)

	ally2, enemy2 := ev.teams(t, svc, 2)
	requireKind(t, svc.FinishTeamGame(context.Background(), ally2, enemy2, ally2), domain.KindFailedPrecondition)

	ally1, enemy1 := ev.teams(t, svc, 1)
	requireNoErr(t, svc.FinishTeamGame(context.Background(), ally1, enemy1, ally1))
	requireKind(t, svc.FinishTeamGame(context.Background(), ally1, enemy1, ally1), domain.KindFailedPrecondition)
}

func TestFinishTeamGameFinishesEventAfterLastGame(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	startEvent(t, svc, repo, ev)

	ally, enemy := ev.teams(t, svc, 1)
	requireNoErr(t, svc.FinishTeamGame(context.Background(), ally, enemy, enemy))

	event := repo.events[ev.id]
	if event.Status != domain.EventStatusFinished {
		t.Errorf("статус ивента %s, ожидался finished", event.Status)
	}
	if event.WinnerSide != "enemy" {
		t.Errorf("winner_side=%q, ожидалось enemy", event.WinnerSide)
	}
	if event.GameCount != 1 {
		t.Errorf("game_count=%d, ожидалось 1", event.GameCount)
	}
	if producer.count("event.finished") != 1 {
		t.Error("не опубликовано завершение ивента")
	}
	if producer.count("team_game.finished") != 2 {
		t.Errorf("сообщений о конце игры %d, ожидалось 2", producer.count("team_game.finished"))
	}
}

func TestFinishTeamGameEndsSeriesEarlyAtTwoZero(t *testing.T) {
	svc, repo, producer := newTestService(t)
	ev := createTestEvent(t, svc, 3)
	startEvent(t, svc, repo, ev)

	ally1, enemy1 := ev.teams(t, svc, 1)
	playGame(t, svc, ally1, enemy1, ally1)

	ally2, enemy2 := ev.teams(t, svc, 2)
	playGame(t, svc, ally2, enemy2, ally2)

	// 2:0 — третью игру играть незачем.
	event := repo.events[ev.id]
	if event.Status != domain.EventStatusFinished {
		t.Errorf("статус ивента %s, ожидался finished при счёте 2:0", event.Status)
	}
	if event.WinnerSide != "ally" {
		t.Errorf("winner_side=%q, ожидалось ally", event.WinnerSide)
	}
	if event.GameCount != 2 {
		t.Errorf("game_count=%d, ожидалось 2", event.GameCount)
	}

	ally3, enemy3 := ev.teams(t, svc, 3)
	if got := repo.teams[ally3].Status; got != domain.TeamStatusPending {
		t.Errorf("несыгранная третья игра осталась в статусе %s, ожидался pending", got)
	}
	requireKind(t, svc.StartTeamGame(context.Background(), ally3, enemy3), domain.KindFailedPrecondition)

	if producer.count("event.finished") != 1 {
		t.Error("завершение ивента должно быть опубликовано ровно один раз")
	}
}

func TestFinishTeamGamePlaysThirdGameAtOneOne(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 3)
	startEvent(t, svc, repo, ev)

	ally1, enemy1 := ev.teams(t, svc, 1)
	playGame(t, svc, ally1, enemy1, ally1)

	ally2, enemy2 := ev.teams(t, svc, 2)
	playGame(t, svc, ally2, enemy2, enemy2)

	if got := repo.events[ev.id].Status; got != domain.EventStatusInProgress {
		t.Fatalf("при счёте 1:1 ивент должен продолжаться, статус %s", got)
	}

	ally3, enemy3 := ev.teams(t, svc, 3)
	playGame(t, svc, ally3, enemy3, enemy3)

	event := repo.events[ev.id]
	if event.Status != domain.EventStatusFinished {
		t.Errorf("статус ивента %s, ожидался finished", event.Status)
	}
	if event.WinnerSide != "enemy" {
		t.Errorf("winner_side=%q, ожидалось enemy", event.WinnerSide)
	}
	if event.GameCount != 3 {
		t.Errorf("game_count=%d, ожидалось 3", event.GameCount)
	}
}

func TestFinishTeamGameDrawOnEvenSeries(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 2)
	startEvent(t, svc, repo, ev)

	ally1, enemy1 := ev.teams(t, svc, 1)
	playGame(t, svc, ally1, enemy1, ally1)

	ally2, enemy2 := ev.teams(t, svc, 2)
	playGame(t, svc, ally2, enemy2, enemy2)

	event := repo.events[ev.id]
	if event.Status != domain.EventStatusFinished {
		t.Errorf("статус ивента %s, ожидался finished", event.Status)
	}
	if event.WinnerSide != "draw" {
		t.Errorf("winner_side=%q, при счёте 1:1 из двух игр ожидалась ничья", event.WinnerSide)
	}
}

func TestAddTeamMemberStats(t *testing.T) {
	setup := func(t *testing.T) (*eventService, *fakeRepo, *fakeProducer, testEvent, string, string) {
		t.Helper()

		svc, repo, producer := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, enemy := ev.teams(t, svc, 1)

		player := joinPlayer(t, svc, ev.id, newID(), false)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, player, domain.RolePlayer))

		startEvent(t, svc, repo, ev)
		requireNoErr(t, svc.FinishTeamGame(context.Background(), ally, enemy, ally))

		return svc, repo, producer, ev, ally, player
	}

	t.Run("невалидные данные", func(t *testing.T) {
		svc, _, _, _, ally, player := setup(t)

		requireKind(t, svc.AddTeamMemberStats(context.Background(), "не-uuid", player, 1, 1, 1, 1, 1), domain.KindInvalidArgument)
		requireKind(t, svc.AddTeamMemberStats(context.Background(), ally, "", 1, 1, 1, 1, 1), domain.KindInvalidArgument)
		requireKind(t, svc.AddTeamMemberStats(context.Background(), ally, player, -1, 0, 0, 0, 0), domain.KindInvalidArgument)
		requireKind(t, svc.AddTeamMemberStats(context.Background(), ally, player, 0, -5, 0, 0, 0), domain.KindInvalidArgument)
		requireKind(t, svc.AddTeamMemberStats(context.Background(), ally, player, 0, 0, maxStatValue+1, 0, 0), domain.KindInvalidArgument)
	})

	t.Run("статистика только после конца игры", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		player := joinPlayer(t, svc, ev.id, newID(), false)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, player, domain.RolePlayer))

		requireKind(t, svc.AddTeamMemberStats(context.Background(), ally, player, 1, 1, 1, 1, 1), domain.KindFailedPrecondition)

		startEvent(t, svc, repo, ev)
		requireKind(t, svc.AddTeamMemberStats(context.Background(), ally, player, 1, 1, 1, 1, 1), domain.KindFailedPrecondition)
	})

	t.Run("игрок не из ивента", func(t *testing.T) {
		svc, _, _, _, ally, _ := setup(t)

		requireKind(t, svc.AddTeamMemberStats(context.Background(), ally, newID(), 1, 1, 1, 1, 1), domain.KindNotFound)
	})

	t.Run("итоги команды пересчитываются, а не накапливаются", func(t *testing.T) {
		svc, repo, producer, ev, ally, player := setup(t)

		requireNoErr(t, svc.AddTeamMemberStats(context.Background(), ally, player, 10, 2, 100, 1, 0))
		requireNoErr(t, svc.AddTeamMemberStats(context.Background(), ally, ev.creator, 5, 1, 50, 0, 2))

		team, stats, err := svc.GetTeamStats(context.Background(), ally)
		requireNoErr(t, err)

		if team.TotalKills != 15 || team.TotalDeaths != 3 || team.TotalPoints != 150 || team.TotalRevival != 1 || team.TotalDestroyedVehicles != 2 {
			t.Errorf("итоги команды посчитаны неверно: %+v", team)
		}
		if len(stats) != 2 {
			t.Errorf("в статистике %d игроков, ожидалось 2", len(stats))
		}

		// Повторная отправка исправленных чисел заменяет прежние, а не удваивает итог.
		requireNoErr(t, svc.AddTeamMemberStats(context.Background(), ally, player, 1, 0, 10, 0, 0))

		if got := repo.teams[ally].TotalKills; got != 6 {
			t.Errorf("total_kills=%d, ожидалось 6 (1 + 5)", got)
		}
		if producer.count("team_member.stats_added") != 3 {
			t.Errorf("сообщений о статистике %d, ожидалось 3", producer.count("team_member.stats_added"))
		}
	})
}

func TestGetTeamStatsReturnsTeamAndMembers(t *testing.T) {
	svc, _, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	ally, _ := ev.teams(t, svc, 1)

	team, stats, err := svc.GetTeamStats(context.Background(), ally)
	requireNoErr(t, err)

	if team.TeamID != ally {
		t.Errorf("вернулась команда %q, ожидалась %q", team.TeamID, ally)
	}
	if len(stats) != 1 {
		t.Fatalf("в составе %d игроков, ожидался сайд-лидер", len(stats))
	}
	if stats[0].UserID != ev.creator {
		t.Errorf("user_id участника %q, ожидался %q", stats[0].UserID, ev.creator)
	}

	_, _, err = svc.GetTeamStats(context.Background(), newID())
	requireKind(t, err, domain.KindNotFound)
}

func TestFinishTeamGameRejectsMismatchedTeams(t *testing.T) {
	t.Run("команды разных ивентов", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		first := createTestEvent(t, svc, 1)
		second := createTestEvent(t, svc, 1)
		startEvent(t, svc, repo, first)
		startEvent(t, svc, repo, second)

		firstAlly, _ := first.teams(t, svc, 1)
		_, secondEnemy := second.teams(t, svc, 1)

		err := svc.FinishTeamGame(context.Background(), firstAlly, secondEnemy, firstAlly)
		requireKind(t, err, domain.KindInvalidArgument)
	})

	t.Run("команды разных игр", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 2)
		startEvent(t, svc, repo, ev)

		ally1, _ := ev.teams(t, svc, 1)
		_, enemy2 := ev.teams(t, svc, 2)

		err := svc.FinishTeamGame(context.Background(), ally1, enemy2, ally1)
		requireKind(t, err, domain.KindInvalidArgument)
	})

	t.Run("несуществующая вторая команда", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		startEvent(t, svc, repo, ev)
		ally, _ := ev.teams(t, svc, 1)

		err := svc.FinishTeamGame(context.Background(), ally, newID(), ally)
		requireKind(t, err, domain.KindNotFound)
	})
}

func TestStartTeamGameRejectsMissingSecondTeam(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	startEvent(t, svc, repo, ev)
	ally, _ := ev.teams(t, svc, 1)

	requireKind(t, svc.StartTeamGame(context.Background(), ally, newID()), domain.KindNotFound)
}

func TestFinishTeamGameFailsWhenSideLeadersUnreadable(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	startEvent(t, svc, repo, ev)
	ally, enemy := ev.teams(t, svc, 1)

	// Итог серии считается по сайд-лидерам: без них ивент не завершить.
	repo.fail("GetUserByID", errDB)

	requireInternal(t, svc.FinishTeamGame(context.Background(), ally, enemy, ally))

	repo.unfail("GetUserByID")
	if got := repo.events[ev.id].GameCount; got != 0 {
		t.Errorf("после отката game_count=%d, ожидался 0", got)
	}
}

func TestFinishTeamGameFailsWhenTeamsUnreadable(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	startEvent(t, svc, repo, ev)
	ally, enemy := ev.teams(t, svc, 1)

	repo.fail("GetTeamsByEventID", errDB)

	requireInternal(t, svc.FinishTeamGame(context.Background(), ally, enemy, ally))
}

func TestFinishTeamGameRequiresRunningEvent(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 2)
	startEvent(t, svc, repo, ev)
	ally, enemy := ev.teams(t, svc, 1)

	// Игра идёт, но сам ивент уже не в работе.
	setEventFields(repo, ev.id, func(event *domain.Event) {
		event.Status = domain.EventStatusFinished
	})

	requireKind(t, svc.FinishTeamGame(context.Background(), ally, enemy, ally), domain.KindFailedPrecondition)
}

func TestFinishTeamGameFailsWhenGameCannotBeClosed(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	startEvent(t, svc, repo, ev)
	ally, enemy := ev.teams(t, svc, 1)

	repo.fail("FinishTeamGame", errDB)
	requireInternal(t, svc.FinishTeamGame(context.Background(), ally, enemy, ally))
}

func TestAddTeamMemberStatsFailsWhenTeamUnreadable(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	ally, _ := ev.teams(t, svc, 1)

	repo.fail("GetTeamByID", errDB)
	requireInternal(t, svc.AddTeamMemberStats(context.Background(), ally, ev.creator, 1, 1, 1, 1, 1))
}

func TestGameCallsFailOnRepositoryErrors(t *testing.T) {
	t.Run("StartTeamGame: чтение команд ивента", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 2)
		startEvent(t, svc, repo, ev)

		ally1, enemy1 := ev.teams(t, svc, 1)
		requireNoErr(t, svc.FinishTeamGame(context.Background(), ally1, enemy1, ally1))

		ally2, enemy2 := ev.teams(t, svc, 2)
		repo.fail("GetTeamsByEventID", errDB)

		requireInternal(t, svc.StartTeamGame(context.Background(), ally2, enemy2))
	})

	t.Run("FinishTeamGame: счётчик игр", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		startEvent(t, svc, repo, ev)
		ally, enemy := ev.teams(t, svc, 1)
		repo.fail("IncrementEventGameCount", errDB)

		requireInternal(t, svc.FinishTeamGame(context.Background(), ally, enemy, ally))

		// Игра не закрыта: откат вернул её в исходное состояние.
		if got := repo.teams[ally].Status; got != domain.TeamStatusInProgress {
			t.Errorf("статус команды %s, ожидался in_progress после отката", got)
		}
	})

	t.Run("FinishTeamGame: завершение ивента", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		startEvent(t, svc, repo, ev)
		ally, enemy := ev.teams(t, svc, 1)
		repo.fail("FinishEventDB", errDB)

		requireInternal(t, svc.FinishTeamGame(context.Background(), ally, enemy, ally))

		if got := repo.events[ev.id].GameCount; got != 0 {
			t.Errorf("после отката game_count=%d, ожидался 0", got)
		}
	})

	t.Run("AddTeamMemberStats: запись игрока", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		startEvent(t, svc, repo, ev)
		ally, enemy := ev.teams(t, svc, 1)
		requireNoErr(t, svc.FinishTeamGame(context.Background(), ally, enemy, ally))
		repo.fail("AddTeamMemberStats", errDB)

		requireInternal(t, svc.AddTeamMemberStats(context.Background(), ally, ev.creator, 1, 1, 1, 1, 1))
	})

	t.Run("AddTeamMemberStats: пересчёт итогов", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		startEvent(t, svc, repo, ev)
		ally, enemy := ev.teams(t, svc, 1)
		requireNoErr(t, svc.FinishTeamGame(context.Background(), ally, enemy, ally))
		repo.fail("SumTeamMemberStats", errDB)

		requireInternal(t, svc.AddTeamMemberStats(context.Background(), ally, ev.creator, 1, 1, 1, 1, 1))
	})

	t.Run("AddTeamMemberStats: запись итогов", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		startEvent(t, svc, repo, ev)
		ally, enemy := ev.teams(t, svc, 1)
		requireNoErr(t, svc.FinishTeamGame(context.Background(), ally, enemy, ally))
		repo.fail("UpdateTeamTotals", errDB)

		requireInternal(t, svc.AddTeamMemberStats(context.Background(), ally, ev.creator, 1, 1, 1, 1, 1))
	})

	t.Run("GetTeamStats", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		repo.fail("GetTeamStats", errDB)

		_, _, err := svc.GetTeamStats(context.Background(), ally)
		requireInternal(t, err)
	})
}

func TestGamePublishFailuresAreReported(t *testing.T) {
	t.Run("StartTeamGame", func(t *testing.T) {
		svc, repo, producer := newTestService(t)
		ev := createTestEvent(t, svc, 2)
		startEvent(t, svc, repo, ev)

		ally1, enemy1 := ev.teams(t, svc, 1)
		requireNoErr(t, svc.FinishTeamGame(context.Background(), ally1, enemy1, ally1))

		ally2, enemy2 := ev.teams(t, svc, 2)
		producer.fail("team_game.started", errKafka)

		if err := svc.StartTeamGame(context.Background(), ally2, enemy2); err == nil {
			t.Error("ошибка публикации должна возвращаться")
		}
	})

	t.Run("FinishTeamGame: конец игры", func(t *testing.T) {
		svc, repo, producer := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		startEvent(t, svc, repo, ev)
		ally, enemy := ev.teams(t, svc, 1)
		producer.fail("team_game.finished", errKafka)

		if err := svc.FinishTeamGame(context.Background(), ally, enemy, ally); err == nil {
			t.Error("ошибка публикации должна возвращаться")
		}
	})

	t.Run("FinishTeamGame: конец ивента", func(t *testing.T) {
		svc, repo, producer := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		startEvent(t, svc, repo, ev)
		ally, enemy := ev.teams(t, svc, 1)
		producer.fail("event.finished", errKafka)

		if err := svc.FinishTeamGame(context.Background(), ally, enemy, ally); err == nil {
			t.Error("ошибка публикации должна возвращаться")
		}

		// Сам ивент при этом уже завершён в базе.
		if got := eventStatus(repo, ev.id); got != domain.EventStatusFinished {
			t.Errorf("статус %s, ожидался finished", got)
		}
	})

	t.Run("AddTeamMemberStats", func(t *testing.T) {
		svc, repo, producer := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		startEvent(t, svc, repo, ev)
		ally, enemy := ev.teams(t, svc, 1)
		requireNoErr(t, svc.FinishTeamGame(context.Background(), ally, enemy, ally))
		producer.fail("team_member.stats_added", errKafka)

		if err := svc.AddTeamMemberStats(context.Background(), ally, ev.creator, 1, 1, 1, 1, 1); err == nil {
			t.Error("ошибка публикации должна возвращаться")
		}
	})
}
