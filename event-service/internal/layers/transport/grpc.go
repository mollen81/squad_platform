package transport

import (
	"context"

	"event-service/internal/core/domain"
	pb "event-service/internal/core/proto"
	service "event-service/internal/layers/service"
)

// Ошибки уходят клиенту gRPC-статусами (см. fail), а не полем error в теле
// ответа: поля error из ответов убраны, как и в остальных сервисах платформы.
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
	err := t.eventService.CreateEvent(ctx, req.GetUserCreatorId(), req.GetCreatorClanId(), req.GetEnemySideLeaderId(), req.GetEnemySideLeaderClanId(), req.GetEventName(), req.GetTimeStart().AsTime(), req.GetTargetGameCount())
	if err != nil {
		return nil, fail("CreateEvent", err)
	}

	return &pb.CreateEventResponse{}, nil
}

func (t *GRPCTransport) GetEventsByCreatorId(ctx context.Context, req *pb.GetEventsByCreatorIdRequest) (*pb.GetEventsByCreatorIdResponse, error) {
	events, err := t.eventService.GetEventsByCreatorId(ctx, req.GetUserCreateId())
	if err != nil {
		return nil, fail("GetEventsByCreatorId", err)
	}

	return &pb.GetEventsByCreatorIdResponse{
		Events: toProtoEvents(events),
	}, nil
}

func (t *GRPCTransport) GetLastEventByCreatorId(ctx context.Context, req *pb.GetLastEventByCreatorIdRequest) (*pb.GetLastEventByCreatorIdResponse, error) {
	event, err := t.eventService.GetLastEventByCreatorId(ctx, req.GetUserCreateId())
	if err != nil {
		return nil, fail("GetLastEventByCreatorId", err)
	}

	resp := &pb.GetLastEventByCreatorIdResponse{}
	if event != nil {
		resp.Event = toProtoEvent(*event)
	}

	return resp, nil
}

func (t *GRPCTransport) GetEventsByEventName(ctx context.Context, req *pb.GetEventsByEventNameRequest) (*pb.GetEventsByEventNameResponse, error) {
	events, err := t.eventService.GetEventsByEventName(ctx, req.GetEventName())
	if err != nil {
		return nil, fail("GetEventsByEventName", err)
	}

	return &pb.GetEventsByEventNameResponse{
		Events: toProtoEventsSlice(events),
	}, nil
}

func (t *GRPCTransport) UpdateTimeEvent(ctx context.Context, req *pb.UpdateTimeEventRequest) (*pb.UpdateTimeEventResponse, error) {
	err := t.eventService.UpdateTimeEvent(ctx, req.GetEventId(), req.GetUserCreateId(), req.GetNewTimeStart().AsTime())
	if err != nil {
		return nil, fail("UpdateTimeEvent", err)
	}

	return &pb.UpdateTimeEventResponse{}, nil
}

func (t *GRPCTransport) CancelEvent(ctx context.Context, req *pb.CancelEventRequest) (*pb.CancelEventResponse, error) {
	err := t.eventService.CancelEvent(ctx, req.GetEventId(), req.GetUserCreateId())
	if err != nil {
		return nil, fail("CancelEvent", err)
	}

	return &pb.CancelEventResponse{}, nil
}

func (t *GRPCTransport) JoinToEvent(ctx context.Context, req *pb.JoinToEventRequest) (*pb.JoinToEventResponse, error) {
	err := t.eventService.JoinToEvent(ctx, req.GetEventId(), req.GetUserId(), req.GetClanId(), req.GetEnemy())
	if err != nil {
		return nil, fail("JoinToEvent", err)
	}

	return &pb.JoinToEventResponse{}, nil
}

func (t *GRPCTransport) LeaveEvent(ctx context.Context, req *pb.LeaveEventRequest) (*pb.LeaveEventResponse, error) {
	err := t.eventService.LeaveEvent(ctx, req.GetUserId(), req.GetEventId())
	if err != nil {
		return nil, fail("LeaveEvent", err)
	}

	return &pb.LeaveEventResponse{}, nil
}

func (t *GRPCTransport) SetRole(ctx context.Context, req *pb.SetRoleRequest) (*pb.SetRoleResponse, error) {
	err := t.eventService.SetRole(ctx, req.GetTeamId(), req.GetSideLeaderId(), req.GetUserId(), domain.Role(req.GetRole()))
	if err != nil {
		return nil, fail("SetRole", err)
	}

	return &pb.SetRoleResponse{}, nil
}

func (t *GRPCTransport) GetTeamsByEventID(ctx context.Context, req *pb.GetTeamsByEventIDRequest) (*pb.GetTeamsByEventIDResponse, error) {
	teams, err := t.eventService.GetTeamsByEventID(ctx, req.GetEventId())
	if err != nil {
		return nil, fail("GetTeamsByEventID", err)
	}

	return &pb.GetTeamsByEventIDResponse{
		Teams: toProtoTeams(teams),
	}, nil
}

func (t *GRPCTransport) GetTeamByID(ctx context.Context, req *pb.GetTeamByIDRequest) (*pb.GetTeamByIDResponse, error) {
	team, err := t.eventService.GetTeamByID(ctx, req.GetTeamId())
	if err != nil {
		return nil, fail("GetTeamByID", err)
	}

	return &pb.GetTeamByIDResponse{
		Team: toProtoTeam(team),
	}, nil
}

func (t *GRPCTransport) JoinUserToTeam(ctx context.Context, req *pb.JoinUserToTeamRequest) (*pb.JoinUserToTeamResponse, error) {
	err := t.eventService.JoinUserToTeam(ctx, req.GetTeamId(), req.GetUserId(), domain.Role(req.GetRole()))
	if err != nil {
		return nil, fail("JoinUserToTeam", err)
	}

	return &pb.JoinUserToTeamResponse{}, nil
}

func (t *GRPCTransport) RemoveUserFromTeam(ctx context.Context, req *pb.RemoveUserFromTeamRequest) (*pb.RemoveUserFromTeamResponse, error) {
	err := t.eventService.RemoveUserFromTeam(ctx, req.GetTeamId(), req.GetUserId())
	if err != nil {
		return nil, fail("RemoveUserFromTeam", err)
	}

	return &pb.RemoveUserFromTeamResponse{}, nil
}

func (t *GRPCTransport) StartTeamGame(ctx context.Context, req *pb.StartTeamGameRequest) (*pb.StartTeamGameResponse, error) {
	err := t.eventService.StartTeamGame(ctx, req.GetTeam1Id(), req.GetTeam2Id())
	if err != nil {
		return nil, fail("StartTeamGame", err)
	}

	return &pb.StartTeamGameResponse{}, nil
}

func (t *GRPCTransport) FinishTeamGame(ctx context.Context, req *pb.FinishTeamGameRequest) (*pb.FinishTeamGameResponse, error) {
	err := t.eventService.FinishTeamGame(ctx, req.GetTeam1Id(), req.GetTeam2Id(), req.GetTeamWinnerId())
	if err != nil {
		return nil, fail("FinishTeamGame", err)
	}

	return &pb.FinishTeamGameResponse{}, nil
}

func (t *GRPCTransport) AddTeamMemberStats(ctx context.Context, req *pb.AddTeamMemberStatsRequest) (*pb.AddTeamMemberStatsResponse, error) {
	err := t.eventService.AddTeamMemberStats(ctx, req.GetTeamId(), req.GetUserId(), req.GetKills(), req.GetDeaths(), req.GetPoints(), req.GetRevival(), req.GetDestroyedVehicles())
	if err != nil {
		return nil, fail("AddTeamMemberStats", err)
	}

	return &pb.AddTeamMemberStatsResponse{}, nil
}

func (t *GRPCTransport) GetTeamStats(ctx context.Context, req *pb.GetTeamStatsRequest) (*pb.GetTeamStatsResponse, error) {
	team, stats, err := t.eventService.GetTeamStats(ctx, req.GetTeamId())
	if err != nil {
		return nil, fail("GetTeamStats", err)
	}

	return &pb.GetTeamStatsResponse{
		Stats:                  toProtoTeamMembers(stats),
		TotalKills:             team.TotalKills,
		TotalDeaths:            team.TotalDeaths,
		TotalPoints:            team.TotalPoints,
		TotalRevival:           team.TotalRevival,
		TotalDestroyedVehicles: team.TotalDestroyedVehicles,
	}, nil
}

func (t *GRPCTransport) GetEventMembersList(ctx context.Context, req *pb.GetEventMembersListRequest) (*pb.GetEventMembersListResponse, error) {
	users, err := t.eventService.GetEventMembersList(ctx, req.GetEventId())
	if err != nil {
		return nil, fail("GetEventMembersList", err)
	}

	return &pb.GetEventMembersListResponse{
		Users: toProtoUsersSlice(users),
	}, nil
}

// Публичного StartEvent RPC больше нет: старт ивента и первой игры происходит
// автоматически по таймеру внутри eventService (controlEventTimerDenial →
// checkMinPlayers → startEventAndGames), а не по вызову от клиента.

func (t *GRPCTransport) GetUnfinishedEventsByUserID(ctx context.Context, req *pb.GetUnfinishedEventsByUserIDRequest) (*pb.GetUnfinishedEventsByUserIDResponse, error) {
	events, err := t.eventService.GetUnfinishedEventsByUserID(ctx, req.GetUserCreateId())
	if err != nil {
		return nil, fail("GetUnfinishedEventsByUserID", err)
	}

	return &pb.GetUnfinishedEventsByUserIDResponse{
		Events: toProtoEvents(events),
	}, nil
}

func (t *GRPCTransport) GetUnfinishedEventsByEventName(ctx context.Context, req *pb.GetUnfinishedEventsByEventNameRequest) (*pb.GetUnfinishedEventsByEventNameResponse, error) {
	events, err := t.eventService.GetUnfinishedEventsByEventName(ctx, req.GetEventName())
	if err != nil {
		return nil, fail("GetUnfinishedEventsByEventName", err)
	}

	return &pb.GetUnfinishedEventsByEventNameResponse{
		Events: toProtoEventsSlice(events),
	}, nil
}
