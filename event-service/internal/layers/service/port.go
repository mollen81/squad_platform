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
	// CancelEvent — ручная отмена ивента создателем, пока тот ещё pending или
	// confirmed (переводит в статус "canceled"). Для "declined" (не набралось
	// минимума игроков к контрольной точке) — см. checkMinPlayers, это внутренний,
	// не ручкой вызываемый путь.
	CancelEvent(ctx context.Context, eventID, userCreateID string) error
	JoinToEvent(ctx context.Context, eventID, userID, clanID string, enemy bool) error
	GetEventMembersList(ctx context.Context, eventID string) ([]domain.User, error)
	// LeaveEvent — выход из ивента, пока тот ещё pending (после подтверждения
	// состав заморожен). Сайд-лидерам выход запрещён: они обязаны играть
	// каждую игру. Игрок заодно убирается из всех команд ивента.
	LeaveEvent(ctx context.Context, userID, eventID string) error
	// SetRole назначает роль внутри ОДНОЙ команды (одной игры), поэтому
	// принимает team_id: роли уровня всего ивента больше нет. Вызывать может
	// только сайд-лидер этой же команды и только пока её игра не началась.
	SetRole(ctx context.Context, teamID, yourID, userID string, role domain.Role) error

	GetTeamsByEventID(ctx context.Context, eventID string) ([]domain.Team, error)
	GetTeamByID(ctx context.Context, teamID string) (domain.Team, error)
	JoinUserToTeam(ctx context.Context, teamID, userEventID string, role domain.Role) error
	RemoveUserFromTeam(ctx context.Context, teamID, userEventID string) error
	StartTeamGame(ctx context.Context, team1ID, team2ID string) error
	FinishTeamGame(ctx context.Context, team1ID, team2ID, teamWinnerID string) error
	AddTeamMemberStats(ctx context.Context, teamID, userID string, kills, deaths, points, revival, destroyedVehicles int64) error
	GetTeamStats(ctx context.Context, teamID string) (domain.Team, []domain.TeamMember, error)

	// StartEvent — запуск ивента сайд-лидером. Доступен, когда ивент в статусе
	// ready (сервер задеплоился и выждал domain.ServerReadyDelay). Ивент
	// стартует, когда нажали ОБА сайд-лидера; если через
	// domain.StartVoteTimeout нажал только один — стартуем автоматически.
	// Каждое нажатие публикуется в Kafka, чтобы вторая сторона видела
	// готовность соперника.
	StartEvent(ctx context.Context, eventID, userID string) error
	// GetServerData отдаёт реквизиты сервера любому участнику ивента: user_id
	// приходит из API, которое уже проверило токен, поэтому здесь проверяется
	// только участие в ивенте.
	GetServerData(ctx context.Context, eventID, userID string) (serverID, serverPassword string, err error)

	// ServerPurchased и ServerDeployed вызываются входящим консьюмером на
	// сообщениях vps.purchased и vps.deployed соответственно (см.
	// transport/kafka.go): сервер под ивент покупает и разворачивает другой
	// микросервис. В vps.deployed времени нет — момент готовности сервис
	// берёт по своим часам.
	ServerPurchased(ctx context.Context, eventID, serverID, serverPassword string) error
	ServerDeployed(ctx context.Context, eventID, serverID string) error

	// RecoverPendingEvents переживает рестарт процесса: eventTimers живёт только
	// в памяти, поэтому при старте нужно заново расставить таймеры контроля/старта
	// по всем ивентам, которые ещё не завершены. Вызывается один раз из main при
	// запуске сервиса.
	RecoverPendingEvents(ctx context.Context) error
}

// EventProducer — то, что сервис публикует в Kafka. Интерфейс (а не
// *kafka.Producer) нужен, чтобы бизнес-логику можно было проверять без
// брокера: *kafka.Producer реализует его как есть.
type EventProducer interface {
	PublishEventCreated(ctx context.Context, event domain.Event) error
	PublishEventCanceled(ctx context.Context, eventID, userCreateID string) error
	PublishEventTimeUpdated(ctx context.Context, eventID, userCreateID string, newTimeStart time.Time) error
	PublishEventConfirmed(ctx context.Context, eventID string) error
	PublishEventDeclined(ctx context.Context, eventID string) error
	PublishEventStarted(ctx context.Context, eventID string, timeStart time.Time) error
	// PublishEventReady — сервер готов, StartEvent открыт для сайд-лидеров.
	PublishEventReady(ctx context.Context, eventID string, startAvailableAt time.Time) error
	// PublishSideReady — сайд-лидер нажал StartEvent. По side вторая сторона
	// понимает, что соперник готов; bothReady говорит, что ждать больше некого.
	PublishSideReady(ctx context.Context, eventID, userID, side string, readyAt time.Time, bothReady bool) error
	PublishEventFinished(ctx context.Context, eventID string, winnerSide string, timeFinish time.Time) error
	PublishRentServer(ctx context.Context, eventID string, playersList []string, timeStart time.Time) error
	PublishUserJoinedEvent(ctx context.Context, eventID, userID, clanID string, enemy bool, joinTime time.Time) error
	PublishUserLeftEvent(ctx context.Context, eventID, userID string) error
	PublishUserRoleChanged(ctx context.Context, eventID, teamID, userID string, role domain.Role) error
	PublishUserJoinedTeam(ctx context.Context, teamID, userID, userEventID string, role domain.Role) error
	PublishUserLeftTeam(ctx context.Context, teamID, userID, userEventID string) error
	PublishTeamGameStarted(ctx context.Context, teamID string, gameNumber int64, timeStart time.Time) error
	PublishTeamGameFinished(ctx context.Context, teamID string, winner bool, timeFinish time.Time) error
	PublishTeamMemberStatsAdded(ctx context.Context, teamID, userEventID string, kills, deaths, points, revival, destroyedVehicles int64) error
}

type EventRepository interface {
	// WithTx выполняет fn в одной SERIALIZABLE-транзакции: все методы
	// репозитория, вызванные с переданным в fn ctx, идут через неё. При
	// конфликте сериализации/дедлоке fn перезапускается целиком, поэтому
	// внутри fn — только работа с БД, публикации в Kafka — после WithTx.
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error

	CreateEvent(ctx context.Context, event domain.Event) error
	GetEventsByCreatorId(ctx context.Context, userCreateID string) ([]*domain.Event, error)
	GetLastEventByCreatorId(ctx context.Context, userCreateID string) (*domain.Event, error)
	GetEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error)
	GetUnfinishedEventsByUserID(ctx context.Context, userCreateID string) ([]*domain.Event, error)
	GetUnfinishedEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error)
	// GetAllUnfinishedEvents отдаёт все ещё не завершённые ивенты (status !=
	// 'finished'), без фильтра по создателю — используется при рестарте сервиса,
	// чтобы заново расставить таймеры (см. RecoverPendingEvents).
	GetAllUnfinishedEvents(ctx context.Context) ([]domain.Event, error)
	GetEventByID(ctx context.Context, eventID string) (domain.Event, error)
	// GetEventByIDForUpdate блокирует строку ивента (SELECT ... FOR UPDATE)
	// до конца транзакции — имеет смысл только внутри WithTx.
	GetEventByIDForUpdate(ctx context.Context, eventID string) (domain.Event, error)
	// GetEventByIDAndServerIDForUpdate ищет ивент по паре (event_id,
	// server_id) — проверка сообщения vps.deployed. Пустой результат значит,
	// что такой пары нет. Блокирует строку, как и GetEventByIDForUpdate.
	GetEventByIDAndServerIDForUpdate(ctx context.Context, eventID, serverID string) (domain.Event, error)
	UpdateTimeEvent(ctx context.Context, eventID string, newTimeStart time.Time) error
	UpdateTimeFinishEvent(ctx context.Context, eventID string, newTimeFinish time.Time) error
	// UpdateEventStatus переводит ивент в один из следующих статусов жизненного
	// цикла:
	//   pending    -> confirmed  (прошла проверка на минимум игроков)
	//   confirmed  -> in_progress (стартовали игры)
	//   pending    -> declined   (не набралось минимума игроков к контрольной точке)
	//   pending/confirmed -> canceled (создатель вручную отменил через CancelEvent)
	// Реализация обязана менять статус только из ожидаемого предыдущего — так
	// вызов не может откатить ивент назад, если сработает повторно или с
	// опозданием. Переход в "finished" отдельный (см. FinishEventDB), потому
	// что вместе со статусом там же пишутся time_finish и winner_side.
	UpdateEventStatus(ctx context.Context, eventID string, status domain.EventStatus) error
	RenameEvent(ctx context.Context, eventID, oldName, newName string) error
	JoinToEvent(ctx context.Context, userEventID, eventID, userID, clanID string, enemy bool, joinTime time.Time) error
	IsUserInEvent(ctx context.Context, eventID, userID string) (bool, error)
	GetEventMembersList(ctx context.Context, eventID string) ([]domain.User, error)
	LeaveEvent(ctx context.Context, userID, eventID string) error
	GetUserByID(ctx context.Context, eventID, userID string) (domain.User, error)
	// GetUserByUserEventID нужен, чтобы по team.side_leader_id узнать сторону
	// команды (users.enemy).
	GetUserByUserEventID(ctx context.Context, userEventID string) (domain.User, error)
	GetUserIDsByEventID(ctx context.Context, eventID string) ([]string, error)
	// IncrementEventGameCount возвращает game_count после инкремента.
	IncrementEventGameCount(ctx context.Context, eventID string) (int64, error)
	// SetServerData сохраняет id и пароль купленного сервера (vps.purchased).
	SetServerData(ctx context.Context, eventID, serverID, serverPassword string) error
	// SetServerDeployedAt запоминает момент получения vps.deployed.
	// Отказывает, если момент уже записан: повторная доставка сообщения не
	// должна сдвигать открытие StartEvent.
	SetServerDeployedAt(ctx context.Context, eventID string, deployedAt time.Time) error
	// MarkSideReady отмечает нажатие StartEvent одной стороной. Отказывает,
	// если эта сторона уже нажимала.
	MarkSideReady(ctx context.Context, eventID string, enemySide bool, readyAt time.Time) error
	// MarkRentServerSent помечает, что сигнал об аренде сервера уже ушёл:
	// после рестарта сервиса таймер аренды отрабатывает заново, и без флага
	// server.rent публиковался бы повторно.
	MarkRentServerSent(ctx context.Context, eventID string) error
	// FinishEventDB переводит ивент в статус "finished" и одновременно пишет
	// time_finish и итоговый winner_side — атомарно, одним UPDATE'ом, а не
	// отдельным вызовом UpdateEventStatus + отдельным UPDATE на эти поля.
	FinishEventDB(ctx context.Context, eventID, winnerSide string, timeFinish time.Time) error

	CreateTeam(ctx context.Context, team domain.Team) error
	GetTeamsByEventID(ctx context.Context, eventID string) ([]domain.Team, error)
	GetTeamByID(ctx context.Context, teamID string) (domain.Team, error)
	JoinUserToTeam(ctx context.Context, teamID, userEventID string, role domain.Role) error
	RemoveUserFromTeam(ctx context.Context, teamID, userID string) error
	// RemoveUserFromAllTeams убирает игрока из всех команд ивента (нужен при
	// выходе из ивента: иначе строку users не удалить — на неё ссылается
	// team_members) и возвращает id команд, из которых он реально удалён.
	RemoveUserFromAllTeams(ctx context.Context, eventID, userEventID string) ([]string, error)
	// UpdateTeamMemberRole меняет роль игрока внутри команды — роль есть
	// только там (см. SetRole).
	UpdateTeamMemberRole(ctx context.Context, teamID, userEventID string, role domain.Role) error
	// CheckSixClanMembers — six_clan_members считается по клану ВНУТРИ
	// команды (team_members), не по всему ивенту: другой микросервис перед
	// стартом игры смотрит на этот флаг у каждого участника team_members и
	// решает, звать его или нет. Считает ДРУГИХ участников teamID с тем же
	// clanID (исключая userEventID): порог — 5 других однокланников, то есть
	// 6 вместе с самим игроком.
	CheckSixClanMembers(ctx context.Context, teamID, userEventID, clanID string) (bool, error)
	UpdateTeamMemberSixClanMembers(ctx context.Context, teamID, userEventID string, hasSixClanMembers bool) error
	UpdateSixClanMembersForClanInTeam(ctx context.Context, teamID, clanID string, hasSixClanMembers bool) error
	CountClanMembersInTeam(ctx context.Context, teamID, clanID string) (int, error)
	// StartTeamGame стартует сразу обе команды одной игры (game_number) одним
	// UPDATE'ом, поэтому принимает оба team_id.
	StartTeamGame(ctx context.Context, team1ID, team2ID string, timeStart time.Time) error
	// SumTeamMemberStats суммирует индивидуальную статистику всех team_members
	// этой команды (её заносят по одному через AddTeamMemberStats, ПОСЛЕ того как
	// FinishTeamGame уже закрыл игру) — используется для обновления
	// Team.total_* после каждого вызова AddTeamMemberStats.
	SumTeamMemberStats(ctx context.Context, teamID string) (kills, deaths, points, revival, destroyedVehicles int64, err error)
	// UpdateTeamTotals перезаписывает Team.total_* посчитанной через
	// SumTeamMemberStats суммой — вызывается после каждого AddTeamMemberStats.
	UpdateTeamTotals(ctx context.Context, teamID string, kills, deaths, points, revival, destroyedVehicles int64) error
	// FinishTeamGame закрывает ровно одну игру: помечает time_finish и winner
	// сразу у обеих команд (team1ID/team2ID), winnerTeamID — один из этих двух.
	// Статистика team_members сюда не входит — она появляется позже.
	FinishTeamGame(ctx context.Context, team1ID, team2ID, winnerTeamID string, timeFinish time.Time) error
	AddTeamMemberStats(ctx context.Context, teamID, userEventID string, kills, deaths, points, revival, destroyedVehicles int64) error
	// GetTeamStats возвращает статистику по игрокам вместе с их user_id — он не
	// хранится в team_members (только user_event_id), достаётся JOIN'ом с users.
	GetTeamStats(ctx context.Context, teamID string) ([]domain.TeamMember, error)
	IsUserInTeam(ctx context.Context, teamID, userEventID string) (bool, error)
	GetUserEventIDByUserIDAndTeamID(ctx context.Context, userID, teamID string) (string, error)
}
