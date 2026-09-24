package service

import (
	"context"
	"testing"

	domain "event-service/internal/core/domain"
)

func TestJoinUserToTeamValidatesInput(t *testing.T) {
	svc, _, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	ally, _ := ev.teams(t, svc, 1)
	userID := joinPlayer(t, svc, ev.id, newID(), false)

	requireKind(t, svc.JoinUserToTeam(context.Background(), "не-uuid", userID, domain.RolePlayer), domain.KindInvalidArgument)
	requireKind(t, svc.JoinUserToTeam(context.Background(), ally, "", domain.RolePlayer), domain.KindInvalidArgument)

	// squad_leader — роль сайд-лидера, её не выдают через вступление
	requireKind(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RoleSquadLeader), domain.KindInvalidArgument)
	requireKind(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.Role("captain")), domain.KindInvalidArgument)

	requireKind(t, svc.JoinUserToTeam(context.Background(), newID(), userID, domain.RolePlayer), domain.KindNotFound)
	requireKind(t, svc.JoinUserToTeam(context.Background(), ally, newID(), domain.RolePlayer), domain.KindNotFound)
}

func TestJoinUserToTeamOnlyOwnSide(t *testing.T) {
	svc, _, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	ally, enemy := ev.teams(t, svc, 1)

	allyPlayer := joinPlayer(t, svc, ev.id, newID(), false)
	enemyPlayer := joinPlayer(t, svc, ev.id, newID(), true)

	// Играть можно только за ту сторону, за которую вошёл в ивент.
	requireKind(t, svc.JoinUserToTeam(context.Background(), ally, enemyPlayer, domain.RolePlayer), domain.KindFailedPrecondition)
	requireKind(t, svc.JoinUserToTeam(context.Background(), enemy, allyPlayer, domain.RolePlayer), domain.KindFailedPrecondition)

	requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, allyPlayer, domain.RolePlayer))
	requireNoErr(t, svc.JoinUserToTeam(context.Background(), enemy, enemyPlayer, domain.RolePlayer))
}

func TestJoinUserToTeamRejectsDuplicate(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	ally, _ := ev.teams(t, svc, 1)
	userID := joinPlayer(t, svc, ev.id, newID(), false)

	requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))
	requireKind(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer), domain.KindAlreadyExists)

	if got := repo.teams[ally].MembersCount; got != 2 {
		t.Errorf("members_count=%d, ожидалось 2 (сайд-лидер и игрок)", got)
	}
}

func TestJoinUserToTeamRespectsTeamLimit(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	ally, _ := ev.teams(t, svc, 1)

	// Одно место уже занял сайд-лидер, свободных — MaxTeamMembers-1.
	for i := 0; i < MaxTeamMembers-1; i++ {
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))
	}

	if got := repo.teams[ally].MembersCount; got != MaxTeamMembers {
		t.Fatalf("members_count=%d, ожидалось %d", got, MaxTeamMembers)
	}

	extra := joinPlayer(t, svc, ev.id, newID(), false)
	requireKind(t, svc.JoinUserToTeam(context.Background(), ally, extra, domain.RolePlayer), domain.KindFailedPrecondition)

	if got := repo.teams[ally].MembersCount; got != MaxTeamMembers {
		t.Errorf("отказ не должен менять состав: members_count=%d", got)
	}
}

func TestJoinUserToTeamRejectsAfterGameStarted(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	ally, _ := ev.teams(t, svc, 1)
	userID := joinPlayer(t, svc, ev.id, newID(), false)

	startEvent(t, svc, repo, ev)

	requireKind(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer), domain.KindFailedPrecondition)
}

func TestJoinUserToTeamRejectsOnDeadEvent(t *testing.T) {
	for _, status := range []domain.EventStatus{domain.EventStatusCanceled, domain.EventStatusDeclined, domain.EventStatusFinished} {
		t.Run(string(status), func(t *testing.T) {
			svc, repo, _ := newTestService(t)
			ev := createTestEvent(t, svc, 1)
			ally, _ := ev.teams(t, svc, 1)
			userID := joinPlayer(t, svc, ev.id, newID(), false)

			event := repo.events[ev.id]
			event.Status = status
			repo.events[ev.id] = event

			requireKind(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer), domain.KindFailedPrecondition)
		})
	}
}

func TestSixClanMembersFlagFollowsClanSize(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	ally, _ := ev.teams(t, svc, 1)

	clan := newID()
	players := make([]string, 0, 6)
	for i := 0; i < 6; i++ {
		userID := joinPlayer(t, svc, ev.id, clan, false)
		players = append(players, userID)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))
	}

	// Флаг выставляется всей группе клана, а не только тому, кто перевалил порог.
	if got := countSixClanFlags(repo, ally); got != 6 {
		t.Errorf("шестерых однокланников должно быть помечено 6, помечено %d", got)
	}

	// Сайд-лидер из другого клана флага не получает.
	sideLeaderID := repo.teams[ally].SideLeaderID
	if repo.members[ally][sideLeaderID].SixClanMembers {
		t.Error("сайд-лидер из другого клана не должен иметь six_clan_members")
	}

	// Ушёл один — условие перестало выполняться, флаг снимается у всех.
	requireNoErr(t, svc.RemoveUserFromTeam(context.Background(), ally, players[0]))
	if got := countSixClanFlags(repo, ally); got != 0 {
		t.Errorf("после ухода однокланника флаг должен сняться со всех, осталось %d", got)
	}
}

func TestSixClanMembersFlagClearedOnLeaveEvent(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	ally, _ := ev.teams(t, svc, 1)

	clan := newID()
	players := make([]string, 0, 6)
	for i := 0; i < 6; i++ {
		userID := joinPlayer(t, svc, ev.id, clan, false)
		players = append(players, userID)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))
	}

	requireNoErr(t, svc.LeaveEvent(context.Background(), players[0], ev.id))

	if got := countSixClanFlags(repo, ally); got != 0 {
		t.Errorf("после выхода из ивента флаг должен сняться со всех, осталось %d", got)
	}
}

func countSixClanFlags(repo *fakeRepo, teamID string) int {
	count := 0
	for _, member := range repo.members[teamID] {
		if member.SixClanMembers {
			count++
		}
	}
	return count
}

func TestRemoveUserFromTeam(t *testing.T) {
	t.Run("невалидные данные", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)

		requireKind(t, svc.RemoveUserFromTeam(context.Background(), "не-uuid", newID()), domain.KindInvalidArgument)
		requireKind(t, svc.RemoveUserFromTeam(context.Background(), ally, "не-uuid"), domain.KindInvalidArgument)
	})

	t.Run("сайд-лидера убрать нельзя", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)

		requireKind(t, svc.RemoveUserFromTeam(context.Background(), ally, ev.creator), domain.KindFailedPrecondition)
	})

	t.Run("игрока не из состава", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)

		requireKind(t, svc.RemoveUserFromTeam(context.Background(), ally, userID), domain.KindNotFound)
	})

	t.Run("после старта игры состав заморожен", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))

		startEvent(t, svc, repo, ev)

		requireKind(t, svc.RemoveUserFromTeam(context.Background(), ally, userID), domain.KindFailedPrecondition)
	})

	t.Run("успешное удаление", func(t *testing.T) {
		svc, repo, producer := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))

		requireNoErr(t, svc.RemoveUserFromTeam(context.Background(), ally, userID))

		if got := repo.teams[ally].MembersCount; got != 1 {
			t.Errorf("members_count=%d, ожидался только сайд-лидер", got)
		}
		if producer.count("user.left_team") != 1 {
			t.Error("не опубликован выход из команды")
		}
	})
}

func TestSetRole(t *testing.T) {
	t.Run("невалидные данные", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)

		requireKind(t, svc.SetRole(context.Background(), "не-uuid", ev.creator, userID, domain.RolePlayer), domain.KindInvalidArgument)
		requireKind(t, svc.SetRole(context.Background(), ally, "", userID, domain.RolePlayer), domain.KindInvalidArgument)
		requireKind(t, svc.SetRole(context.Background(), ally, ev.creator, userID, domain.RoleSquadLeader), domain.KindInvalidArgument)
		requireKind(t, svc.SetRole(context.Background(), ally, ev.creator, userID, domain.Role("captain")), domain.KindInvalidArgument)
	})

	t.Run("чужой сайд-лидер до команды не дотягивается", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))

		err := svc.SetRole(context.Background(), ally, ev.enemy, userID, domain.RoleSideLeader)
		requireKind(t, err, domain.KindPermissionDenied)

		// И обычный игрок роли не раздаёт.
		other := joinPlayer(t, svc, ev.id, newID(), false)
		requireKind(t, svc.SetRole(context.Background(), ally, other, userID, domain.RoleSideLeader), domain.KindPermissionDenied)
	})

	t.Run("роль самого сайд-лидера не меняется", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)

		err := svc.SetRole(context.Background(), ally, ev.creator, ev.creator, domain.RolePlayer)
		requireKind(t, err, domain.KindFailedPrecondition)
	})

	t.Run("игрок должен быть в составе", func(t *testing.T) {
		svc, _, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)

		requireKind(t, svc.SetRole(context.Background(), ally, ev.creator, userID, domain.RoleSideLeader), domain.KindNotFound)
	})

	t.Run("роль действует только в своей игре", func(t *testing.T) {
		svc, repo, producer := newTestService(t)
		ev := createTestEvent(t, svc, 2)
		ally1, _ := ev.teams(t, svc, 1)
		ally2, _ := ev.teams(t, svc, 2)

		userID := joinPlayer(t, svc, ev.id, newID(), false)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally1, userID, domain.RolePlayer))
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally2, userID, domain.RolePlayer))

		requireNoErr(t, svc.SetRole(context.Background(), ally1, ev.creator, userID, domain.RoleSideLeader))

		user, err := svc.eventRepo.GetUserByID(context.Background(), ev.id, userID)
		requireNoErr(t, err)

		if got := repo.members[ally1][user.UserEventID].Role; got != domain.RoleSideLeader {
			t.Errorf("роль в первой игре %s, ожидалась side_leader", got)
		}
		if got := repo.members[ally2][user.UserEventID].Role; got != domain.RolePlayer {
			t.Errorf("роль во второй игре %s — роль одной игры не должна утекать в другую", got)
		}
		if producer.count("user.role_changed") != 1 {
			t.Error("не опубликована смена роли")
		}
	})

	t.Run("после старта игры роли заморожены", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))

		startEvent(t, svc, repo, ev)

		requireKind(t, svc.SetRole(context.Background(), ally, ev.creator, userID, domain.RoleSideLeader), domain.KindFailedPrecondition)
	})
}

func TestSetRoleFailsOnUnreadableData(t *testing.T) {
	t.Run("команда", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))

		repo.failOn["GetTeamByID"] = errDB
		requireInternal(t, svc.SetRole(context.Background(), ally, ev.creator, userID, domain.RoleSideLeader))
	})

	t.Run("вызывающий", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))

		repo.failOn["GetUserByID"] = errDB
		requireInternal(t, svc.SetRole(context.Background(), ally, ev.creator, userID, domain.RoleSideLeader))
	})
}

func TestTeamCallsFailWhenEventRowUnreadable(t *testing.T) {
	cases := []struct {
		name string
		call func(svc *eventService, ev testEvent, teamID string) error
	}{
		{"JoinUserToTeam", func(svc *eventService, ev testEvent, teamID string) error {
			return svc.JoinUserToTeam(context.Background(), teamID, newID(), domain.RolePlayer)
		}},
		{"RemoveUserFromTeam", func(svc *eventService, ev testEvent, teamID string) error {
			return svc.RemoveUserFromTeam(context.Background(), teamID, ev.creator)
		}},
		{"AddTeamMemberStats", func(svc *eventService, ev testEvent, teamID string) error {
			return svc.AddTeamMemberStats(context.Background(), teamID, ev.creator, 1, 1, 1, 1, 1)
		}},
		{"SetRole", func(svc *eventService, ev testEvent, teamID string) error {
			return svc.SetRole(context.Background(), teamID, ev.creator, newID(), domain.RolePlayer)
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			svc, repo, _ := newTestService(t)
			ev := createTestEvent(t, svc, 1)
			ally, _ := ev.teams(t, svc, 1)

			repo.failOn["GetEventByIDForUpdate"] = errDB
			requireInternal(t, c.call(svc, ev, ally))
		})
	}
}

func TestRefreshSixClanMembersKeepsFlagWhileClanBigEnough(t *testing.T) {
	svc, repo, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	ally, _ := ev.teams(t, svc, 1)

	clan := newID()
	players := make([]string, 0, 7)
	for i := 0; i < 7; i++ {
		userID := joinPlayer(t, svc, ev.id, clan, false)
		players = append(players, userID)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))
	}

	// Ушёл один из семи — шестеро остались, флаг у них сохраняется.
	requireNoErr(t, svc.RemoveUserFromTeam(context.Background(), ally, players[0]))

	if got := countSixClanFlags(repo, ally); got != 6 {
		t.Errorf("помечено %d однокланников, ожидалось 6", got)
	}
}

func TestRefreshSixClanMembersIgnoresEmptyClan(t *testing.T) {
	svc, _, _ := newTestService(t)
	ev := createTestEvent(t, svc, 1)
	ally, _ := ev.teams(t, svc, 1)

	// Без клана считать нечего — и обращения к базе быть не должно.
	requireNoErr(t, svc.refreshSixClanMembers(context.Background(), ally, ""))
}

func TestTeamCallsFailOnRepositoryErrors(t *testing.T) {
	t.Run("JoinUserToTeam: сторона команды", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		repo.failOn["GetUserByUserEventID"] = errDB

		requireInternal(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))
	})

	t.Run("JoinUserToTeam: вставка в состав", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		repo.failOn["JoinUserToTeam"] = errDB

		requireInternal(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))
	})

	t.Run("JoinUserToTeam: подсчёт однокланников", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		repo.failOn["CheckSixClanMembers"] = errDB

		requireInternal(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))

		if got := repo.teams[ally].MembersCount; got != 1 {
			t.Errorf("после отката members_count=%d, ожидался только сайд-лидер", got)
		}
	})

	t.Run("JoinUserToTeam: пометка одиночки", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		repo.failOn["UpdateTeamMemberSixClanMembers"] = errDB

		requireInternal(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))
	})

	t.Run("RemoveUserFromTeam: проверка состава", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))
		repo.failOn["IsUserInTeam"] = errDB

		requireInternal(t, svc.RemoveUserFromTeam(context.Background(), ally, userID))
	})

	t.Run("RemoveUserFromTeam: удаление", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))
		repo.failOn["RemoveUserFromTeam"] = errDB

		requireInternal(t, svc.RemoveUserFromTeam(context.Background(), ally, userID))
	})

	t.Run("RemoveUserFromTeam: снятие флага клана", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		clan := newID()

		var players []string
		for i := 0; i < 6; i++ {
			userID := joinPlayer(t, svc, ev.id, clan, false)
			players = append(players, userID)
			requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))
		}

		repo.failOn["UpdateSixClanMembersForClanInTeam"] = errDB
		requireInternal(t, svc.RemoveUserFromTeam(context.Background(), ally, players[0]))

		delete(repo.failOn, "UpdateSixClanMembersForClanInTeam")
		if got := repo.teams[ally].MembersCount; got != 7 {
			t.Errorf("после отката members_count=%d, ожидалось 7", got)
		}
	})

	t.Run("SetRole: запись роли", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))
		repo.failOn["UpdateTeamMemberRole"] = errDB

		requireInternal(t, svc.SetRole(context.Background(), ally, ev.creator, userID, domain.RoleSideLeader))
	})
}

func TestTeamPublishFailuresAreReported(t *testing.T) {
	t.Run("JoinUserToTeam", func(t *testing.T) {
		svc, _, producer := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		producer.failOn["user.joined_team"] = errKafka

		if err := svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer); err == nil {
			t.Error("ошибка публикации должна возвращаться")
		}
	})

	t.Run("RemoveUserFromTeam", func(t *testing.T) {
		svc, _, producer := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))
		producer.failOn["user.left_team"] = errKafka

		if err := svc.RemoveUserFromTeam(context.Background(), ally, userID); err == nil {
			t.Error("ошибка публикации должна возвращаться")
		}
	})

	t.Run("SetRole", func(t *testing.T) {
		svc, _, producer := newTestService(t)
		ev := createTestEvent(t, svc, 1)
		ally, _ := ev.teams(t, svc, 1)
		userID := joinPlayer(t, svc, ev.id, newID(), false)
		requireNoErr(t, svc.JoinUserToTeam(context.Background(), ally, userID, domain.RolePlayer))
		producer.failOn["user.role_changed"] = errKafka

		if err := svc.SetRole(context.Background(), ally, ev.creator, userID, domain.RoleSideLeader); err == nil {
			t.Error("ошибка публикации должна возвращаться")
		}
	})
}
