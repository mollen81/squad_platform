package domain

import "time"

type Role string

const (
	RolePlayer      Role = "player"
	RoleSquadLeader Role = "squad_leader"
	RoleSideLeader  Role = "side_leader"
)

type Event struct {
	EventID         string
	Name            string
	UserCreateID    string
	EnemySideLeader string
	TimeStart       time.Time
	TimeFinish      time.Time
	CreateTime      time.Time
	UserCount       int64
	TargetGameCount int64
	GameCount       int64
	IsStarted       bool
	IsFinished      bool
	WinnerSide      string // "" (ещё не решено) | "ally" | "enemy" | "draw"
}

type User struct {
	UserEventID    string
	UserID         string
	EventID        string
	ClanID         string
	Enemy          bool
	SixClanMembers bool
	Role           Role
	JoinTime       time.Time
}

type Team struct {
	TeamID             string
	EventID            string
	Winner             bool
	SideLeaderID       string
	GameNumber         int64
	MembersCount       int64
	TimeStart          time.Time
	TimeFinish         time.Time
	Kills              int64
	Deaths             int64
	Revival            int64
	EquipmentDestroyed int64
}

type TeamMember struct {
	TeamID      string
	UserEventID string
	UserID      string
	Role        Role
	Kills       int64
	Deaths      int64
	Points      int64
}
