package transport

import (
	"context"

	"event-service/internal/core/domain"
	pb "event-service/internal/core/proto"
	service "event-service/internal/layers/service"
)

type GRPCTransport struct {
	pb.UnimplementedEventServiceServer
	eventService service.EventService
}

func NewGRPCHandler(eventService service.EventService) *GRPCTransport {
	return &GRPCTransport{
		eventService: eventService,
	}
}

func (t *GRPCTransport) CreateEvent(ctx context.Context, req *pb.CreateEventRequest) (*pb.CreateEventResponse, error) {
	err := t.eventService.CreateEvent(ctx, req.GetUserCreateId(), req.GetEnemySideLeader(), req.GetEventName(), req.GetTimeStart().AsTime())

	return &pb.CreateEventResponse{
		Error: errString(err),
	}, nil
}

func (t *GRPCTransport) GetEventsByUserID(ctx context.Context, req *pb.GetEventsByUserIDRequest) (*pb.GetEventsByUserIDResponse, error) {
	events, err := t.eventService.GetEventsByUserID(ctx, req.GetUserCreateId())

	return &pb.GetEventsByUserIDResponse{
		Events: toProtoEvents(events),
		Error:  errString(err),
	}, nil
}

func (t *GRPCTransport) GetEventsByEventName(ctx context.Context, req *pb.GetEventsByEventNameRequest) (*pb.GetEventsByEventNameResponse, error) {
	events, err := t.eventService.GetEventsByEventName(ctx, req.GetEventName())

	return &pb.GetEventsByEventNameResponse{
		Events: toProtoEventsSlice(events),
		Error:  errString(err),
	}, nil
}

func (t *GRPCTransport) UpdateTimeEvent(ctx context.Context, req *pb.UpdateTimeEventRequest) (*pb.UpdateTimeEventResponse, error) {
	err := t.eventService.UpdateTimeEvent(ctx, req.GetEventId(), req.GetUserCreateId(), req.GetNewTimeStart().AsTime())

	return &pb.UpdateTimeEventResponse{
		Error: errString(err),
	}, nil
}

func (t *GRPCTransport) DeleteEvent(ctx context.Context, req *pb.DeleteEventRequest) (*pb.DeleteEventResponse, error) {
	err := t.eventService.DeleteEvent(ctx, req.GetEventId(), req.GetUserCreateId())

	return &pb.DeleteEventResponse{
		Error: errString(err),
	}, nil
}

func (t *GRPCTransport) JoinToEvent(ctx context.Context, req *pb.JoinToEventRequest) (*pb.JoinToEventResponse, error) {
	err := t.eventService.JoinToEvent(ctx, req.GetEventId(), req.GetUserId(), req.GetJoinTime().AsTime())

	return &pb.JoinToEventResponse{
		Error: errString(err),
	}, nil
}

func (t *GRPCTransport) LeaveEvent(ctx context.Context, req *pb.LeaveEventRequest) (*pb.LeaveEventResponse, error) {
	err := t.eventService.LeaveEvent(ctx, req.GetUserId(), req.GetEventId())

	return &pb.LeaveEventResponse{
		Error: errString(err),
	}, nil
}

func (t *GRPCTransport) SetRole(ctx context.Context, req *pb.SetRoleRequest) (*pb.SetRoleResponse, error) {
	err := t.eventService.SetRole(ctx, req.GetEventId(), req.GetSideLeaderId(), req.GetUserId(), domain.Role(req.GetRole()))

	return &pb.SetRoleResponse{
		Error: errString(err),
	}, nil
}

func (t *GRPCTransport) CreateTeamsForEvent(ctx context.Context, req *pb.CreateTeamsForEventRequest) (*pb.CreateTeamsForEventResponse, error) {
	err := t.eventService.CreateTeamsForEvent(ctx, req.GetEventId())

	return &pb.CreateTeamsForEventResponse{
		Error: errString(err),
	}, nil
}

func (t *GRPCTransport) GetTeamsByEventID(ctx context.Context, req *pb.GetTeamsByEventIDRequest) (*pb.GetTeamsByEventIDResponse, error) {
	teams, err := t.eventService.GetTeamsByEventID(ctx, req.GetEventId())

	return &pb.GetTeamsByEventIDResponse{
		Teams: toProtoTeams(teams),
		Error: errString(err),
	}, nil
}

func (t *GRPCTransport) StartEvent(ctx context.Context, req *pb.StartEventRequest) (*pb.StartEventResponse, error) {
	err := t.eventService.StartEvent(ctx, req.GetEventId(), req.GetSideLeaderId())

	return &pb.StartEventResponse{
		Error: errString(err),
	}, nil
}

func (t *GRPCTransport) CreateGame(ctx context.Context, req *pb.CreateGameRequest) (*pb.CreateGameResponse, error) {
	err := t.eventService.CreateGame(ctx, req.GetEventId(), req.GetMapName(), req.GetTimeStart().AsTime())

	return &pb.CreateGameResponse{
		Error: errString(err),
	}, nil
}

func (t *GRPCTransport) GetGameByID(ctx context.Context, req *pb.GetGameByIDRequest) (*pb.GetGameByIDResponse, error) {
	game, err := t.eventService.GetGameByID(ctx, req.GetGameId())

	resp := &pb.GetGameByIDResponse{
		Error: errString(err),
	}
	if err == nil {
		resp.Game = toProtoGame(game)
	}

	return resp, nil
}

func (t *GRPCTransport) GetGamesByEventID(ctx context.Context, req *pb.GetGamesByEventIDRequest) (*pb.GetGamesByEventIDResponse, error) {
	games, err := t.eventService.GetGamesByEventID(ctx, req.GetEventId())

	return &pb.GetGamesByEventIDResponse{
		Games: toProtoGames(games),
		Error: errString(err),
	}, nil
}

func (t *GRPCTransport) UpdateGameWinner(ctx context.Context, req *pb.UpdateGameWinnerRequest) (*pb.UpdateGameWinnerResponse, error) {
	err := t.eventService.UpdateGameWinner(ctx, req.GetGameId(), req.GetWinnerTeamId())

	return &pb.UpdateGameWinnerResponse{
		Error: errString(err),
	}, nil
}

func (t *GRPCTransport) UpdateGameLoser(ctx context.Context, req *pb.UpdateGameLoserRequest) (*pb.UpdateGameLoserResponse, error) {
	err := t.eventService.UpdateGameLoser(ctx, req.GetGameId(), req.GetLoserTeamId())

	return &pb.UpdateGameLoserResponse{
		Error: errString(err),
	}, nil
}

func (t *GRPCTransport) FinishGame(ctx context.Context, req *pb.FinishGameRequest) (*pb.FinishGameResponse, error) {
	err := t.eventService.FinishGame(ctx, req.GetGameId(), req.GetTimeFinish().AsTime())

	return &pb.FinishGameResponse{
		Error: errString(err),
	}, nil
}

func (t *GRPCTransport) DeleteGame(ctx context.Context, req *pb.DeleteGameRequest) (*pb.DeleteGameResponse, error) {
	err := t.eventService.DeleteGame(ctx, req.GetGameId())

	return &pb.DeleteGameResponse{
		Error: errString(err),
	}, nil
}

func (t *GRPCTransport) AddUserStatsToGame(ctx context.Context, req *pb.AddUserStatsToGameRequest) (*pb.AddUserStatsToGameResponse, error) {
	err := t.eventService.AddUserStatsToGame(ctx, req.GetGameId(), req.GetUserId(), req.GetKills(), req.GetDeaths(), req.GetPoints())

	return &pb.AddUserStatsToGameResponse{
		Error: errString(err),
	}, nil
}

func (t *GRPCTransport) GetGameStats(ctx context.Context, req *pb.GetGameStatsRequest) (*pb.GetGameStatsResponse, error) {
	stats, err := t.eventService.GetGameStats(ctx, req.GetGameId())

	return &pb.GetGameStatsResponse{
		Stats: toProtoGameUserStatsSlice(stats),
		Error: errString(err),
	}, nil
}

func (t *GRPCTransport) GetTeamByID(ctx context.Context, req *pb.GetTeamByIDRequest) (*pb.GetTeamByIDResponse, error) {
	team, err := t.eventService.GetTeamByID(ctx, req.GetTeamId())

	resp := &pb.GetTeamByIDResponse{
		Error: errString(err),
	}
	if err == nil {
		resp.Team = toProtoTeam(team)
	}

	return resp, nil
}

func (t *GRPCTransport) AddUserToTeam(ctx context.Context, req *pb.AddUserToTeamRequest) (*pb.AddUserToTeamResponse, error) {
	err := t.eventService.AddUserToTeam(ctx, req.GetTeamId(), req.GetUserId(), req.GetClanId(), domain.Role(req.GetRole()))

	return &pb.AddUserToTeamResponse{
		Error: errString(err),
	}, nil
}

func (t *GRPCTransport) RemoveUserFromTeam(ctx context.Context, req *pb.RemoveUserFromTeamRequest) (*pb.RemoveUserFromTeamResponse, error) {
	err := t.eventService.RemoveUserFromTeam(ctx, req.GetTeamId(), req.GetUserId())

	return &pb.RemoveUserFromTeamResponse{
		Error: errString(err),
	}, nil
}

func toProtoTeam(team domain.Team) *pb.Team {
	members := make([]*pb.User, 0)
	for _, member := range team.Members {
		if member.UserID != "" {
			members = append(members, &pb.User{
				UserEventId:     member.UserEventID,
				UserId:          member.UserID,
				ClanId:          member.ClanID,
				TeamId:          member.TeamID,
				Role:            string(member.Role),
				SixClanMembers:  member.SixClanMembers,
			})
		}
	}

	return &pb.Team{
		TeamId:       team.TeamID,
		EventId:      team.EventID,
		SideLeaderId: team.SideLeaderID,
		IsConfirmed:  team.IsConfirmed,
		Members:      members,
	}
}

func toProtoTeams(teams []domain.Team) []*pb.Team {
	result := make([]*pb.Team, 0, len(teams))
	for _, team := range teams {
		result = append(result, toProtoTeam(team))
	}
	return result
}
