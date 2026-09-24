package domain

import "time"

// Role — роль игрока ВНУТРИ команды (team_members), то есть в рамках одной
// конкретной игры. В ивенте целиком (users) роли нет: один и тот же человек
// может быть в первой игре сквадным, а во второй — обычным игроком.
type Role string

const (
	// RolePlayer — рядовой игрок
	RolePlayer Role = "player"
	// RoleSquadLeader — главный в команде; его получает сайд-лидер стороны
	// при создании ивента, через SetRole эта роль не выдаётся
	RoleSquadLeader Role = "squad_leader"
	// RoleSideLeader — старше рядовых игроков, но ниже squad_leader
	RoleSideLeader Role = "side_leader"
)

type EventStatus string

const (
	// EventStatusPending — ивент создан, набор участников идёт, проверка на минимум игроков ещё не проводилась
	EventStatusPending EventStatus = "pending"
	// EventStatusConfirmed — проверка на минимум игроков прошла успешно, ивент подтверждён, ожидает старта.
	// Состав ивента с этого момента заморожен: ни войти, ни выйти уже нельзя.
	EventStatusConfirmed EventStatus = "confirmed"
	// EventStatusInProgress — ивент начался, игры идут
	EventStatusInProgress EventStatus = "in_progress"
	// EventStatusFinished — сыграны все игры или победитель определился досрочно, ивент завершён
	EventStatusFinished EventStatus = "finished"
	// EventStatusDeclined — автоматически отменён: к контрольной точке не набралось
	// минимума игроков (см. controlEventTimerDenial)
	EventStatusDeclined EventStatus = "declined"
	// EventStatusCanceled — вручную отменён создателем через CancelEvent
	EventStatusCanceled EventStatus = "canceled"
)

// TeamStatus — состояние участия команды в её игре. Команда (team) — это и
// есть участие одной стороны в одной игре, поэтому её статус и есть статус
// самой игры для этой стороны.
type TeamStatus string

const (
	// TeamStatusPending — игра ещё не началась: только в этом статусе можно
	// менять состав команды (JoinUserToTeam/RemoveUserFromTeam) и роли (SetRole)
	TeamStatusPending TeamStatus = "pending"
	// TeamStatusInProgress — игра идёт: состав и роли заморожены
	TeamStatusInProgress TeamStatus = "in_progress"
	// TeamStatusFinished — игра закончена: можно заносить статистику игроков
	TeamStatusFinished TeamStatus = "finished"
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
	// RentServerSent — сигнал об аренде сервера уже отправлен. Хранится в БД,
	// а не в памяти, чтобы после рестарта сервиса он не ушёл повторно.
	RentServerSent bool
}

type User struct {
	UserEventID string
	UserID      string
	EventID     string
	ClanID      string
	// Enemy — сторона, за которую игрок вошёл в ивент. Войти в команду
	// противоположной стороны нельзя (см. JoinUserToTeam).
	Enemy    bool
	JoinTime time.Time
}

type Team struct {
	TeamID                 string
	EventID                string
	Winner                 bool
	SideLeaderID           string
	GameNumber             int64
	MembersCount           int64
	Status                 TeamStatus // pending | in_progress | finished
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
