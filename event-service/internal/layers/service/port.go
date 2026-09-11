package service

import (
	"context"
	domain "event-service/internal/core/domain"
	"time"
)

type EventService interface {
	CreateEvent(ctx context.Context, userCreatorID, creatorClanID, enemySideLeaderID, enemySideLeaderClanID, eventName string, timeStart time.Time, targetGameCount int64) error
	GetEventsByCreatorId(ctx context.Context, userCreateID string) ([]*domain.Event, error)
	GetLastEventByCreatorId(ctx context.Context, userCreateID string) (*domain.Event, error)
	GetEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error)
	GetUnfinishedEventsByUserID(ctx context.Context, userCreateID string) ([]*domain.Event, error)
	GetUnfinishedEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error)
	UpdateTimeEvent(ctx context.Context, eventID, userCreateID string, newTimeStart time.Time) error
	DeleteEvent(ctx context.Context, eventID, userCreateID string) error
	JoinToEvent(ctx context.Context, eventID, userID, clanID string, enemy bool) error
	GetEventMembersList(ctx context.Context, eventID string) ([]domain.User, error)
	LeaveEvent(ctx context.Context, userID, eventID string) error
	SetRole(ctx context.Context, eventID, yourID, userID string, role domain.Role) error

	GetTeamsByEventID(ctx context.Context, eventID string) ([]domain.Team, error)
	GetTeamByID(ctx context.Context, teamID string) (domain.Team, error)
	AddUserToTeam(ctx context.Context, teamID, userEventID string, role domain.Role) error
	RemoveUserFromTeam(ctx context.Context, teamID, userEventID string) error
	StartTeamGame(ctx context.Context, teamID string) error
	FinishTeamGame(ctx context.Context, teamID string, winner bool, kills, deaths, revival, equipmentDestroyed int64) error
	AddTeamMemberStats(ctx context.Context, teamID, userEventID string, kills, deaths, points int64) error
	GetTeamStats(ctx context.Context, teamID string) ([]domain.TeamMember, error)
}

type EventRepository interface {
	CreateEvent(ctx context.Context, event domain.Event) error
	GetEventsByCreatorId(ctx context.Context, userCreateID string) ([]*domain.Event, error)
	GetLastEventByCreatorId(ctx context.Context, userCreateID string) (*domain.Event, error)
	GetEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error)
	GetUnfinishedEventsByUserID(ctx context.Context, userCreateID string) ([]*domain.Event, error)
	GetUnfinishedEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error)
	GetEventByID(ctx context.Context, eventID string) (domain.Event, error)
	UpdateTimeEvent(ctx context.Context, eventID string, newTimeStart time.Time) error
	UpdateTimeFinishEvent(ctx context.Context, eventID string, newTimeFinish time.Time) error
	RenameEvent(ctx context.Context, eventID, oldName, newName string) error
	DeleteEvent(ctx context.Context, eventID string) error
	JoinToEvent(ctx context.Context, userEventID, eventID, userID, clanID string, enemy bool, joinTime time.Time) error
	IsUserInEvent(ctx context.Context, eventID, userID string) (bool, error)
	GetEventMembersList(ctx context.Context, eventID string) ([]domain.User, error)
	LeaveEvent(ctx context.Context, userID, eventID string) error
	UpdateUserRole(ctx context.Context, userID, eventID string, role domain.Role) error
	GetUserByID(ctx context.Context, eventID, userID string) (domain.User, error)
	CheckSixClanMembers(ctx context.Context, eventID, userID, clanID string) (bool, error)
	UpdateUserSixClanMembers(ctx context.Context, userID, eventID string, hasSixClanMembers bool) error
	UpdateSixClanMembersForClan(ctx context.Context, eventID, clanID string, hasSixClanMembers bool) error
	CountClanMembersInEvent(ctx context.Context, eventID, clanID string) (int, error)
	GetUserIDsByEventID(ctx context.Context, eventID string) ([]string, error)
	StartEventDB(ctx context.Context, eventID string) error
	IncrementEventGameCount(ctx context.Context, eventID string) error
	FinishEventDB(ctx context.Context, eventID, winnerSide string) error

	CreateTeam(ctx context.Context, team domain.Team) error
	GetTeamsByEventID(ctx context.Context, eventID string) ([]domain.Team, error)
	GetTeamByID(ctx context.Context, teamID string) (domain.Team, error)
	AddUserToTeam(ctx context.Context, teamID, userEventID string, role domain.Role) error
	RemoveUserFromTeam(ctx context.Context, teamID, userEventID string) error
	StartTeamGame(ctx context.Context, teamID string, timeStart time.Time) error
	FinishTeamGame(ctx context.Context, teamID string, timeFinish time.Time, winner bool, kills, deaths, revival, equipmentDestroyed int64) error
	AddTeamMemberStats(ctx context.Context, teamID, userEventID string, kills, deaths, points int64) error
	GetTeamStats(ctx context.Context, teamID string) ([]domain.TeamMember, error)
}
