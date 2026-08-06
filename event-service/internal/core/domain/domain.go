package domain

import "time"

type Role string

const (
	RolePlayer      Role = "player"
	RoleSquadLeader Role = "squad_leader"
	RoleSideLeader  Role = "side_leader"
)

type Event struct {
	EventID string
	Name string
	UserCreateID string
	EnemySideLeader string
	UserCount int64
	TimeStart time.Time
	TimeFinish time.Time
	CreateTime time.Time
	Event_team_winner string
	Event_team_loser string
}

type User struct {
	UserEventID string
	UserID string
	ClanID string
	TeamID string
	Role Role
}

type Game struct {
	GameID string
	EventID string
	UserCreateID string
	EnemySideLeader string
	Team1ID string
	Team2ID string
	MapName string
	Game_team_winner_id string
	Game_team_loser_id string
	TimeStart time.Time
	TimeFinish time.Time
}

type GameUserStats struct {
	GameUserStatsID string
	Game
	User
	Kills int64
	Deaths int64
	Points int64
}

type Team struct {
	TeamID string
	GameID string
	Members [50]User
}