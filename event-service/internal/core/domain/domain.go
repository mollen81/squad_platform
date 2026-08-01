package domain

import "time"

type Role string

const (
	RolePlayer      Role = "player"
	RoleSquadLeader Role = "squad_leader"
	RoleSideLeader  Role = "side_leader"
)

type Event struct {
	ID string
	Name string
	UserCreateID string
	UserCount int64
	TimeStart time.Time
	TimeFinish time.Time
	CreateTime time.Time
	Event_team_winner string
	Event_team_loser string
}

type User struct {
	ID string
	ClanID string
	TeamID string
	Role Role
}

type Game struct {
	ID string
	EventID string
	MapID string
	Game_team_winner_id string
	Game_team_loser_id string
	TimeStart time.Time
	TimeFinish time.Time
}

type GameUserStats struct {
	ID string
	User
	Kills int64
	Deaths int64
	Points int64
}

type Team struct {
	ID string
	Members [50]User
}