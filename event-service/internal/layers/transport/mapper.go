package transport

import (
	domain "event-service/internal/core/domain"
	pb "event-service/internal/core/proto"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoEvent(e domain.Event) *pb.Event {
	return &pb.Event{
		EventId:           e.EventID,
		EventName:         e.Name,
		UserCreateId:      e.UserCreateID,
		EnemySideLeaderId: e.EnemySideLeader,
		UserCount:         e.UserCount,
		TimeStart:         timestamppb.New(e.TimeStart),
		TimeFinish:        timestamppb.New(e.TimeFinish),
		CreateTime:        timestamppb.New(e.CreateTime),
		WinnerSide:        e.WinnerSide,
		IsStarted:         e.IsStarted,
		IsFinished:        e.IsFinished,
		TargetGameCount:   e.TargetGameCount,
		GameCount:         e.GameCount,
	}
}

func toProtoEvents(events []*domain.Event) []*pb.Event {
	result := make([]*pb.Event, 0, len(events))
	for _, e := range events {
		if e == nil {
			continue
		}
		result = append(result, toProtoEvent(*e))
	}
	return result
}

func toProtoEventsSlice(events []domain.Event) []*pb.Event {
	result := make([]*pb.Event, 0, len(events))
	for _, e := range events {
		result = append(result, toProtoEvent(e))
	}
	return result
}

func toProtoUser(u domain.User) *pb.User {
	return &pb.User{
		UserEventId:    u.UserEventID,
		UserId:         u.UserID,
		EventId:        u.EventID,
		ClanId:         u.ClanID,
		Enemy:          u.Enemy,
		Role:           string(u.Role),
		SixClanMembers: u.SixClanMembers,
		JoinTime:       timestamppb.New(u.JoinTime),
	}
}

func toProtoUsersSlice(users []domain.User) []*pb.User {
	result := make([]*pb.User, 0, len(users))
	for _, e := range users {
		result = append(result, toProtoUser(e))
	}
	return result
}

func toProtoTeam(t domain.Team) *pb.Team {
	team := &pb.Team{
		TeamId:             t.TeamID,
		EventId:            t.EventID,
		SideLeaderId:       t.SideLeaderID,
		GameNumber:         t.GameNumber,
		MembersCount:       t.MembersCount,
		Winner:             t.Winner,
		Kills:              t.Kills,
		Deaths:             t.Deaths,
		Revival:            t.Revival,
		EquipmentDestroyed: t.EquipmentDestroyed,
	}

	// пустой (zero-value) time.Time -> оставляем поле в proto как nil,
	// а не как timestamp "0001-01-01" — чтобы клиент мог проверить "игра не началась/не окончена" по наличию поля
	if !t.TimeStart.IsZero() {
		team.TimeStart = timestamppb.New(t.TimeStart)
	}
	if !t.TimeFinish.IsZero() {
		team.TimeFinish = timestamppb.New(t.TimeFinish)
	}

	return team
}

func toProtoTeams(teams []domain.Team) []*pb.Team {
	result := make([]*pb.Team, 0, len(teams))
	for _, t := range teams {
		result = append(result, toProtoTeam(t))
	}
	return result
}

func toProtoTeamMember(m domain.TeamMember) *pb.TeamMember {
	return &pb.TeamMember{
		TeamId:      m.TeamID,
		UserEventId: m.UserEventID,
		UserId:      m.UserID,
		Role:        string(m.Role),
		Kills:       m.Kills,
		Deaths:      m.Deaths,
		Points:      m.Points,
	}
}

func toProtoTeamMembers(members []domain.TeamMember) []*pb.TeamMember {
	result := make([]*pb.TeamMember, 0, len(members))
	for _, m := range members {
		result = append(result, toProtoTeamMember(m))
	}
	return result
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
