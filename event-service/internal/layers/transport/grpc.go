package transport

import (
	"context"

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
	err := t.eventService.CreateEvent(ctx, req.GetUserCreateId(), req.GetEventName(), req.GetTimeStart().AsTime())

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

func (t *GRPCTransport) CreateGame(ctx context.Context, req *pb.CreateGameRequest) (*pb.CreateGameResponse, error) {
	err := t.eventService.CreateGame(ctx, req.GetEventId(), req.GetMapId(), req.GetTimeStart().AsTime())

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
