package transport

import (
	domain "event-service/internal/core/domain"
	pb "event-service/internal/core/proto"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoEvent(e domain.Event) *pb.Event {
	return &pb.Event{
		EventId:         e.EventID,
		Name:            e.Name,
		UserCreateId:    e.UserCreateID,
		EnemySideLeader: e.EnemySideLeader,
		UserCount:       e.UserCount,
		TimeStart:       timestamppb.New(e.TimeStart),
		TimeFinish:      timestamppb.New(e.TimeFinish),
		CreateTime:      timestamppb.New(e.CreateTime),
		EventTeamWinner: e.Event_team_winner,
		EventTeamLoser:  e.Event_team_loser,
		IsConfirmed:     e.IsConfirmed,
		IsStarted:       e.IsStarted,
		IsFinished:      e.IsFinished,
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
		UserEventId: u.UserEventID,
		UserId:      u.UserID,
		ClanId:      u.ClanID,
		TeamId:      u.TeamID,
		Role:        string(u.Role),
	}
}

func toProtoGame(g domain.Game) *pb.Game {
	return &pb.Game{
		GameId:           g.GameID,
		EventId:          g.EventID,
		MapName:            g.MapName,
		GameTeamWinnerId: g.Game_team_winner_id,
		GameTeamLoserId:  g.Game_team_loser_id,
		TimeStart:        timestamppb.New(g.TimeStart),
		TimeFinish:       timestamppb.New(g.TimeFinish),
	}
}

func toProtoGames(games []domain.Game) []*pb.Game {
	result := make([]*pb.Game, 0, len(games))
	for _, g := range games {
		result = append(result, toProtoGame(g))
	}
	return result
}

func toProtoGameUserStats(s domain.GameUserStats) *pb.GameUserStats {
	return &pb.GameUserStats{
		GameUserStatsId: s.GameUserStatsID,
		Game:            toProtoGame(s.Game),
		User:            toProtoUser(s.User),
		Kills:           s.Kills,
		Deaths:          s.Deaths,
		Points:          s.Points,
	}
}

func toProtoGameUserStatsSlice(stats []domain.GameUserStats) []*pb.GameUserStats {
	result := make([]*pb.GameUserStats, 0, len(stats))
	for _, s := range stats {
		result = append(result, toProtoGameUserStats(s))
	}
	return result
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
