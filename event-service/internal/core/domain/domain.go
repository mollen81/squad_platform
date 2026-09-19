package domain

import "time"

type Role string

const (
	RolePlayer      Role = "player"
	RoleSquadLeader Role = "squad_leader"
	RoleSideLeader  Role = "side_leader"
)

type EventStatus string

const (
	// EventStatusPending — ивент создан, набор участников идёт, проверка на минимум игроков ещё не проводилась
	EventStatusPending EventStatus = "pending"
	// EventStatusConfirmed — проверка на минимум игроков прошла успешно, ивент подтверждён, ожидает старта
	EventStatusConfirmed EventStatus = "confirmed"
	// EventStatusInProgress — ивент начался, игры идут
	EventStatusInProgress EventStatus = "in_progress"
	// EventStatusFinished — все игры сыграны, ивент завершён
	EventStatusFinished EventStatus = "finished"
	// EventStatusDeclined — автоматически отменён: к контрольной точке не набралось
	// минимума игроков (см. controlEventTimerDenial)
	EventStatusDeclined EventStatus = "declined"
	// EventStatusCanceled — вручную отменён создателем через CancelEvent
	EventStatusCanceled EventStatus = "canceled"
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
	Status          EventStatus // pending | confirmed | in_progress | finished | declined | canceled
	WinnerSide      string      // "" (ещё не решено) | "ally" | "enemy" | "draw"
}

type User struct {
	UserEventID string
	UserID      string
	EventID     string
	ClanID      string
	Enemy       bool
	Role        Role
	JoinTime    time.Time
}

type Team struct {
	TeamID                 string
	EventID                string
	Winner                 bool
	SideLeaderID           string
	GameNumber             int64
	MembersCount           int64
	TimeStart              time.Time
	TimeFinish             time.Time
	TotalKills             int64
	TotalDeaths            int64
	TotalPoints            int64
	TotalRevival           int64
	TotalDestroyedVehicles int64
}

type TeamMember struct {
	TeamID            string
	UserEventID       string
	UserID            string
	Role              Role
	Kills             int64
	Deaths            int64
	Points            int64
	Revival           int64
	DestroyedVehicles int64
	// SixClanMembers — сигнал для другого микросервиса: перед стартом игры
	// он смотрит на этот флаг у каждого участника команды и решает, звать
	// его в игру (true) или проигнорировать (false). Считается по клану
	// внутри команды (team_members), не по всему ивенту.
	SixClanMembers bool
}
