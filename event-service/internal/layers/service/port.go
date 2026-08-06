package service

import (
	"context"
	"time"
	domain "event-service/internal/core/domain"
)

type EventService interface {
	CreateEvent(ctx context.Context, userCreateID, eventName string, timeStart time.Time) error
	GetEventsByUserID(ctx context.Context, userCreateID string) ([]*domain.Event, error)
	GetEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error)
	UpdateTimeEvent(ctx context.Context, eventID, userCreateID string, newTimeStart time.Time) error
	DeleteEvent(ctx context.Context, eventID, userCreateID string) error
	JoinToEvent(ctx context.Context, eventID, userID string, joinTime time.Time) error
	LeaveEvent(ctx context.Context, userID, eventID string) error
	CreateGame(ctx context.Context, eventID, mapName string, timeStart time.Time) error
	GetGameByID(ctx context.Context, gameID string) (domain.Game, error)
	GetGamesByEventID(ctx context.Context, eventID string) ([]domain.Game, error)
	UpdateGameWinner(ctx context.Context, gameID, winnerTeamID string) error
	UpdateGameLoser(ctx context.Context, gameID, loserTeamID string) error
	FinishGame(ctx context.Context, gameID string, timeFinish time.Time) error
	DeleteGame(ctx context.Context, gameID string) error
	AddUserStatsToGame(ctx context.Context, gameID, userID string, kills, deaths, points int64) error
	GetGameStats(ctx context.Context, gameID string) ([]domain.GameUserStats, error)
	GetTeamByID(ctx context.Context, teamID string) (domain.Team, error)
	AddUserToTeam(ctx context.Context, teamID, userID, clanID string, role domain.Role) error
	RemoveUserFromTeam(ctx context.Context, teamID, userID string) error
}

type EventRepository interface {
	CreateEvent(ctx context.Context, event domain.Event) error
	GetEventsByUserID(ctx context.Context, userCreateID string) ([]*domain.Event, error)
	GetEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error)
	GetEventByID(ctx context.Context, eventID string) (domain.Event, error)
	UpdateTimeEvent(ctx context.Context, eventID string, newTimeStart time.Time) error
	UpdateTimeFinishEvent(ctx context.Context, eventID string, newTimeFinish time.Time) error
	UpdateEventWinner(ctx context.Context, eventID, winnerTeamID string) error
	UpdateEventLoser(ctx context.Context, eventID, loserTeamID string) error
	RenameEvent(ctx context.Context, eventID, oldName, newName string) error
	DeleteEvent(ctx context.Context, eventID string) error
	JoinToEvent(ctx context.Context, userID, eventID string, joinTime time.Time) error
	LeaveEvent(ctx context.Context, userID, eventID string) error
	CreateGame(ctx context.Context, game domain.Game) error
	GetGameByID(ctx context.Context, gameID string) (domain.Game, error)
	GetGamesByEventID(ctx context.Context, eventID string) ([]domain.Game, error)
	UpdateGameWinner(ctx context.Context, gameID, winnerTeamID string) error
	UpdateGameLoser(ctx context.Context, gameID, loserTeamID string) error
	FinishGame(ctx context.Context, gameID string, timeFinish time.Time) error
	DeleteGame(ctx context.Context, gameID string) error
	GetUserByID(ctx context.Context, userID string) (domain.User, error)
	AddUserStatsToGame(ctx context.Context, stats domain.GameUserStats) error
	GetGameStats(ctx context.Context, gameID string) ([]domain.GameUserStats, error)
	CreateTeam(ctx context.Context, team domain.Team) error
	GetTeamByID(ctx context.Context, teamID string) (domain.Team, error)
	AddUserToTeam(ctx context.Context, teamID, userID, clanID string, role domain.Role) error
	RemoveUserFromTeam(ctx context.Context, teamID, userID string) error
}
