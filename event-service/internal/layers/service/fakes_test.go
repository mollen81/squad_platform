package service

import (
	"context"
	"errors"
	"sort"
	"sync"
	"testing"
	"time"

	domain "event-service/internal/core/domain"
	uuid "github.com/google/uuid"
)

// fakeRepo — хранилище в памяти, повторяющее поведение postgres-репозитория:
// те же условия перед изменением (статусы, уникальность, "затронута ли
// строка") и тот же откат всего сделанного, если функция внутри WithTx
// вернула ошибку. За счёт этого тесты проверяют бизнес-логику сервиса, а не
// удобные заглушки.
type fakeRepo struct {
	mu      sync.Mutex
	events  map[string]domain.Event
	users   map[string]domain.User                  // по user_event_id
	teams   map[string]domain.Team                  // по team_id
	members map[string]map[string]domain.TeamMember // team_id -> user_event_id

	// failOn заставляет конкретный метод вернуть ошибку — так проверяется,
	// что при сбое посреди транзакции не остаётся половины изменений
	failOn map[string]error
	inTx   bool
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		events:  map[string]domain.Event{},
		users:   map[string]domain.User{},
		teams:   map[string]domain.Team{},
		members: map[string]map[string]domain.TeamMember{},
		failOn:  map[string]error{},
	}
}

func (f *fakeRepo) check(method string) error {
	if err, ok := f.failOn[method]; ok {
		return err
	}
	return nil
}

func (f *fakeRepo) snapshot() *fakeRepo {
	clone := newFakeRepo()
	for k, v := range f.events {
		clone.events[k] = v
	}
	for k, v := range f.users {
		clone.users[k] = v
	}
	for k, v := range f.teams {
		clone.teams[k] = v
	}
	for teamID, ms := range f.members {
		clone.members[teamID] = map[string]domain.TeamMember{}
		for k, v := range ms {
			clone.members[teamID][k] = v
		}
	}
	return clone
}

func (f *fakeRepo) restore(from *fakeRepo) {
	f.events, f.users, f.teams, f.members = from.events, from.users, from.teams, from.members
}

func (f *fakeRepo) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	f.mu.Lock()
	if f.inTx {
		f.mu.Unlock()
		return fn(ctx)
	}

	before := f.snapshot()
	f.inTx = true
	f.mu.Unlock()

	err := fn(ctx)

	f.mu.Lock()
	f.inTx = false
	if err != nil {
		f.restore(before)
	}
	f.mu.Unlock()

	return err
}

// ── events ──────────────────────────────────────────────────────────────────

func (f *fakeRepo) CreateEvent(ctx context.Context, event domain.Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("CreateEvent"); err != nil {
		return err
	}

	if _, exists := f.events[event.EventID]; exists {
		return domain.AlreadyExists("event %q already exists", event.EventID)
	}

	f.events[event.EventID] = event
	return nil
}

func (f *fakeRepo) GetEventByID(ctx context.Context, eventID string) (domain.Event, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("GetEventByID"); err != nil {
		return domain.Event{}, err
	}

	return f.events[eventID], nil
}

func (f *fakeRepo) GetEventByIDForUpdate(ctx context.Context, eventID string) (domain.Event, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("GetEventByIDForUpdate"); err != nil {
		return domain.Event{}, err
	}

	return f.events[eventID], nil
}

func (f *fakeRepo) GetEventsByCreatorId(ctx context.Context, userCreateID string) ([]*domain.Event, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("GetEventsByCreatorId"); err != nil {
		return nil, err
	}

	result := make([]*domain.Event, 0)
	for _, e := range f.sortedEvents() {
		if e.UserCreateID == userCreateID {
			event := e
			result = append(result, &event)
		}
	}
	return result, nil
}

func (f *fakeRepo) GetLastEventByCreatorId(ctx context.Context, userCreateID string) (*domain.Event, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("GetLastEventByCreatorId"); err != nil {
		return nil, err
	}

	var last *domain.Event
	for _, e := range f.sortedEvents() {
		if e.UserCreateID != userCreateID {
			continue
		}
		if last == nil || e.CreateTime.After(last.CreateTime) {
			event := e
			last = &event
		}
	}
	return last, nil
}

func (f *fakeRepo) GetEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("GetEventsByEventName"); err != nil {
		return nil, err
	}

	result := make([]domain.Event, 0)
	for _, e := range f.sortedEvents() {
		if e.Name == eventName {
			result = append(result, e)
		}
	}
	return result, nil
}

func (f *fakeRepo) GetUnfinishedEventsByUserID(ctx context.Context, userCreateID string) ([]*domain.Event, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("GetUnfinishedEventsByUserID"); err != nil {
		return nil, err
	}

	result := make([]*domain.Event, 0)
	for _, e := range f.sortedEvents() {
		if e.UserCreateID == userCreateID && isUnfinished(e.Status) {
			event := e
			result = append(result, &event)
		}
	}
	return result, nil
}

func (f *fakeRepo) GetUnfinishedEventsByEventName(ctx context.Context, eventName string) ([]domain.Event, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("GetUnfinishedEventsByEventName"); err != nil {
		return nil, err
	}

	result := make([]domain.Event, 0)
	for _, e := range f.sortedEvents() {
		if e.Name == eventName && isUnfinished(e.Status) {
			result = append(result, e)
		}
	}
	return result, nil
}

func (f *fakeRepo) GetAllUnfinishedEvents(ctx context.Context) ([]domain.Event, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("GetAllUnfinishedEvents"); err != nil {
		return nil, err
	}

	result := make([]domain.Event, 0)
	for _, e := range f.sortedEvents() {
		if isUnfinished(e.Status) {
			result = append(result, e)
		}
	}
	return result, nil
}

func isUnfinished(status domain.EventStatus) bool {
	return status == domain.EventStatusPending ||
		status == domain.EventStatusConfirmed ||
		status == domain.EventStatusInProgress
}

func (f *fakeRepo) sortedEvents() []domain.Event {
	ids := make([]string, 0, len(f.events))
	for id := range f.events {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	result := make([]domain.Event, 0, len(ids))
	for _, id := range ids {
		result = append(result, f.events[id])
	}
	return result
}

func (f *fakeRepo) UpdateTimeEvent(ctx context.Context, eventID string, newTimeStart time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("UpdateTimeEvent"); err != nil {
		return err
	}

	event, ok := f.events[eventID]
	if !ok || event.Status != domain.EventStatusPending {
		return domain.FailedPrecondition("cannot change time of event %q: it is not pending", eventID)
	}

	event.TimeStart = newTimeStart
	f.events[eventID] = event
	return nil
}

func (f *fakeRepo) UpdateTimeFinishEvent(ctx context.Context, eventID string, newTimeFinish time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	event, ok := f.events[eventID]
	if !ok {
		return domain.NotFound("event %q not found", eventID)
	}

	event.TimeFinish = newTimeFinish
	f.events[eventID] = event
	return nil
}

func (f *fakeRepo) UpdateEventStatus(ctx context.Context, eventID string, status domain.EventStatus) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("UpdateEventStatus"); err != nil {
		return err
	}

	allowedFrom := map[domain.EventStatus][]domain.EventStatus{
		domain.EventStatusConfirmed:  {domain.EventStatusPending},
		domain.EventStatusInProgress: {domain.EventStatusConfirmed},
		domain.EventStatusDeclined:   {domain.EventStatusPending},
		domain.EventStatusCanceled:   {domain.EventStatusPending, domain.EventStatusConfirmed},
	}

	from, ok := allowedFrom[status]
	if !ok {
		return domain.FailedPrecondition("unsupported target status %q", status)
	}

	event := f.events[eventID]
	for _, allowed := range from {
		if event.Status == allowed {
			event.Status = status
			f.events[eventID] = event
			return nil
		}
	}

	return domain.FailedPrecondition("cannot move event %q to status %q", eventID, status)
}

func (f *fakeRepo) RenameEvent(ctx context.Context, eventID, oldName, newName string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	event, ok := f.events[eventID]
	if !ok || event.Name != oldName {
		return domain.NotFound("event %q not found", eventID)
	}

	event.Name = newName
	f.events[eventID] = event
	return nil
}

func (f *fakeRepo) IncrementEventGameCount(ctx context.Context, eventID string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("IncrementEventGameCount"); err != nil {
		return 0, err
	}

	event, ok := f.events[eventID]
	if !ok {
		return 0, domain.NotFound("event %q not found", eventID)
	}

	event.GameCount++
	f.events[eventID] = event
	return event.GameCount, nil
}

func (f *fakeRepo) MarkRentServerSent(ctx context.Context, eventID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("MarkRentServerSent"); err != nil {
		return err
	}

	event, ok := f.events[eventID]
	if !ok {
		return domain.NotFound("event %q not found", eventID)
	}

	event.RentServerSent = true
	f.events[eventID] = event
	return nil
}

func (f *fakeRepo) FinishEventDB(ctx context.Context, eventID, winnerSide string, timeFinish time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("FinishEventDB"); err != nil {
		return err
	}

	event, ok := f.events[eventID]
	if !ok || event.Status != domain.EventStatusInProgress {
		return domain.FailedPrecondition("cannot finish event %q: it is not in_progress", eventID)
	}

	event.Status = domain.EventStatusFinished
	event.WinnerSide = winnerSide
	event.TimeFinish = timeFinish
	f.events[eventID] = event
	return nil
}

// ── users ───────────────────────────────────────────────────────────────────

func (f *fakeRepo) JoinToEvent(ctx context.Context, userEventID, eventID, userID, clanID string, enemy bool, joinTime time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("JoinToEvent"); err != nil {
		return err
	}

	for _, u := range f.users {
		if u.EventID == eventID && u.UserID == userID {
			return domain.AlreadyExists("user %q has already joined event %q", userID, eventID)
		}
	}

	f.users[userEventID] = domain.User{
		UserEventID: userEventID,
		UserID:      userID,
		EventID:     eventID,
		ClanID:      clanID,
		Enemy:       enemy,
		JoinTime:    joinTime,
	}

	event := f.events[eventID]
	event.UserCount++
	f.events[eventID] = event
	return nil
}

func (f *fakeRepo) IsUserInEvent(ctx context.Context, eventID, userID string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("IsUserInEvent"); err != nil {
		return false, err
	}

	for _, u := range f.users {
		if u.EventID == eventID && u.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeRepo) GetEventMembersList(ctx context.Context, eventID string) ([]domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("GetEventMembersList"); err != nil {
		return nil, err
	}

	result := make([]domain.User, 0)
	for _, u := range f.sortedUsers() {
		if u.EventID == eventID {
			result = append(result, u)
		}
	}
	return result, nil
}

func (f *fakeRepo) LeaveEvent(ctx context.Context, userID, eventID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("LeaveEvent"); err != nil {
		return err
	}

	for id, u := range f.users {
		if u.EventID == eventID && u.UserID == userID {
			delete(f.users, id)
			event := f.events[eventID]
			event.UserCount--
			f.events[eventID] = event
			return nil
		}
	}

	return domain.NotFound("user %q not found in event %q", userID, eventID)
}

func (f *fakeRepo) GetUserByID(ctx context.Context, eventID, userID string) (domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("GetUserByID"); err != nil {
		return domain.User{}, err
	}

	for _, u := range f.sortedUsers() {
		if u.EventID == eventID && u.UserID == userID {
			return u, nil
		}
	}

	return domain.User{}, domain.NotFound("user %q not found in event %q", userID, eventID)
}

func (f *fakeRepo) GetUserByUserEventID(ctx context.Context, userEventID string) (domain.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("GetUserByUserEventID"); err != nil {
		return domain.User{}, err
	}

	user, ok := f.users[userEventID]
	if !ok {
		return domain.User{}, domain.NotFound("user_event_id %q not found", userEventID)
	}
	return user, nil
}

func (f *fakeRepo) GetUserIDsByEventID(ctx context.Context, eventID string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("GetUserIDsByEventID"); err != nil {
		return nil, err
	}

	result := make([]string, 0)
	for _, u := range f.sortedUsers() {
		if u.EventID == eventID {
			result = append(result, u.UserID)
		}
	}
	return result, nil
}

func (f *fakeRepo) GetUserEventIDByUserIDAndTeamID(ctx context.Context, userID, teamID string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("GetUserEventIDByUserIDAndTeamID"); err != nil {
		return "", err
	}

	team, ok := f.teams[teamID]
	if !ok {
		return "", domain.NotFound("team %q not found", teamID)
	}

	for _, u := range f.sortedUsers() {
		if u.EventID == team.EventID && u.UserID == userID {
			return u.UserEventID, nil
		}
	}

	return "", domain.NotFound("user %q not found in event of team %q", userID, teamID)
}

func (f *fakeRepo) sortedUsers() []domain.User {
	ids := make([]string, 0, len(f.users))
	for id := range f.users {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	result := make([]domain.User, 0, len(ids))
	for _, id := range ids {
		result = append(result, f.users[id])
	}
	return result
}

// ── teams ───────────────────────────────────────────────────────────────────

func (f *fakeRepo) CreateTeam(ctx context.Context, team domain.Team) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("CreateTeam"); err != nil {
		return err
	}

	if team.Status == "" {
		team.Status = domain.TeamStatusPending
	}

	f.teams[team.TeamID] = team
	f.members[team.TeamID] = map[string]domain.TeamMember{}
	return nil
}

func (f *fakeRepo) GetTeamsByEventID(ctx context.Context, eventID string) ([]domain.Team, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("GetTeamsByEventID"); err != nil {
		return nil, err
	}

	result := make([]domain.Team, 0)
	for _, t := range f.teams {
		if t.EventID == eventID {
			result = append(result, t)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].GameNumber != result[j].GameNumber {
			return result[i].GameNumber < result[j].GameNumber
		}
		return result[i].TeamID < result[j].TeamID
	})
	return result, nil
}

func (f *fakeRepo) GetTeamByID(ctx context.Context, teamID string) (domain.Team, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("GetTeamByID"); err != nil {
		return domain.Team{}, err
	}

	return f.teams[teamID], nil
}

func (f *fakeRepo) JoinUserToTeam(ctx context.Context, teamID, userEventID string, role domain.Role) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("JoinUserToTeam"); err != nil {
		return err
	}

	if f.members[teamID] == nil {
		f.members[teamID] = map[string]domain.TeamMember{}
	}

	if _, exists := f.members[teamID][userEventID]; exists {
		return domain.AlreadyExists("user is already in team %q", teamID)
	}

	f.members[teamID][userEventID] = domain.TeamMember{
		TeamID:      teamID,
		UserEventID: userEventID,
		Role:        role,
	}

	team := f.teams[teamID]
	team.MembersCount++
	f.teams[teamID] = team
	return nil
}

func (f *fakeRepo) RemoveUserFromTeam(ctx context.Context, teamID, userEventID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("RemoveUserFromTeam"); err != nil {
		return err
	}

	if _, exists := f.members[teamID][userEventID]; !exists {
		return domain.NotFound("user is not a member of team %q", teamID)
	}

	delete(f.members[teamID], userEventID)
	team := f.teams[teamID]
	team.MembersCount--
	f.teams[teamID] = team
	return nil
}

func (f *fakeRepo) RemoveUserFromAllTeams(ctx context.Context, eventID, userEventID string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("RemoveUserFromAllTeams"); err != nil {
		return nil, err
	}

	teamIDs := make([]string, 0)
	for _, t := range f.sortedTeams() {
		if t.EventID != eventID {
			continue
		}
		if _, exists := f.members[t.TeamID][userEventID]; !exists {
			continue
		}

		delete(f.members[t.TeamID], userEventID)
		team := f.teams[t.TeamID]
		team.MembersCount--
		f.teams[t.TeamID] = team
		teamIDs = append(teamIDs, t.TeamID)
	}

	return teamIDs, nil
}

func (f *fakeRepo) UpdateTeamMemberRole(ctx context.Context, teamID, userEventID string, role domain.Role) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("UpdateTeamMemberRole"); err != nil {
		return err
	}

	member, exists := f.members[teamID][userEventID]
	if !exists {
		return domain.NotFound("user is not a member of team %q", teamID)
	}

	member.Role = role
	f.members[teamID][userEventID] = member
	return nil
}

func (f *fakeRepo) CheckSixClanMembers(ctx context.Context, teamID, userEventID, clanID string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("CheckSixClanMembers"); err != nil {
		return false, err
	}

	count := 0
	for id := range f.members[teamID] {
		if id == userEventID {
			continue
		}
		if f.users[id].ClanID == clanID {
			count++
		}
	}
	return count >= 5, nil
}

func (f *fakeRepo) UpdateTeamMemberSixClanMembers(ctx context.Context, teamID, userEventID string, hasSixClanMembers bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("UpdateTeamMemberSixClanMembers"); err != nil {
		return err
	}

	member, exists := f.members[teamID][userEventID]
	if !exists {
		return domain.NotFound("user is not a member of team %q", teamID)
	}

	member.SixClanMembers = hasSixClanMembers
	f.members[teamID][userEventID] = member
	return nil
}

func (f *fakeRepo) UpdateSixClanMembersForClanInTeam(ctx context.Context, teamID, clanID string, hasSixClanMembers bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("UpdateSixClanMembersForClanInTeam"); err != nil {
		return err
	}

	for id, member := range f.members[teamID] {
		if f.users[id].ClanID != clanID {
			continue
		}
		member.SixClanMembers = hasSixClanMembers
		f.members[teamID][id] = member
	}
	return nil
}

func (f *fakeRepo) CountClanMembersInTeam(ctx context.Context, teamID, clanID string) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("CountClanMembersInTeam"); err != nil {
		return 0, err
	}

	count := 0
	for id := range f.members[teamID] {
		if f.users[id].ClanID == clanID {
			count++
		}
	}
	return count, nil
}

func (f *fakeRepo) IsUserInTeam(ctx context.Context, teamID, userEventID string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("IsUserInTeam"); err != nil {
		return false, err
	}

	_, exists := f.members[teamID][userEventID]
	return exists, nil
}

func (f *fakeRepo) StartTeamGame(ctx context.Context, team1ID, team2ID string, timeStart time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("StartTeamGame"); err != nil {
		return err
	}

	for _, id := range []string{team1ID, team2ID} {
		if f.teams[id].Status != domain.TeamStatusPending {
			return domain.FailedPrecondition("cannot start game: it is already started or finished")
		}
	}

	for _, id := range []string{team1ID, team2ID} {
		team := f.teams[id]
		team.Status = domain.TeamStatusInProgress
		team.TimeStart = timeStart
		f.teams[id] = team
	}
	return nil
}

func (f *fakeRepo) FinishTeamGame(ctx context.Context, team1ID, team2ID, winnerTeamID string, timeFinish time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("FinishTeamGame"); err != nil {
		return err
	}

	for _, id := range []string{team1ID, team2ID} {
		if f.teams[id].Status != domain.TeamStatusInProgress {
			return domain.FailedPrecondition("cannot finish game: it is not in progress")
		}
	}

	for _, id := range []string{team1ID, team2ID} {
		team := f.teams[id]
		team.Status = domain.TeamStatusFinished
		team.TimeFinish = timeFinish
		team.Winner = id == winnerTeamID
		f.teams[id] = team
	}
	return nil
}

func (f *fakeRepo) AddTeamMemberStats(ctx context.Context, teamID, userEventID string, kills, deaths, points, revival, destroyedVehicles int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("AddTeamMemberStats"); err != nil {
		return err
	}

	member, exists := f.members[teamID][userEventID]
	if !exists {
		return domain.NotFound("player is not a member of team %q", teamID)
	}

	member.Kills, member.Deaths, member.Points = kills, deaths, points
	member.Revival, member.DestroyedVehicles = revival, destroyedVehicles
	f.members[teamID][userEventID] = member
	return nil
}

func (f *fakeRepo) SumTeamMemberStats(ctx context.Context, teamID string) (kills, deaths, points, revival, destroyedVehicles int64, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("SumTeamMemberStats"); err != nil {
		return 0, 0, 0, 0, 0, err
	}

	for _, m := range f.members[teamID] {
		kills += m.Kills
		deaths += m.Deaths
		points += m.Points
		revival += m.Revival
		destroyedVehicles += m.DestroyedVehicles
	}
	return kills, deaths, points, revival, destroyedVehicles, nil
}

func (f *fakeRepo) UpdateTeamTotals(ctx context.Context, teamID string, kills, deaths, points, revival, destroyedVehicles int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("UpdateTeamTotals"); err != nil {
		return err
	}

	team := f.teams[teamID]
	team.TotalKills, team.TotalDeaths, team.TotalPoints = kills, deaths, points
	team.TotalRevival, team.TotalDestroyedVehicles = revival, destroyedVehicles
	f.teams[teamID] = team
	return nil
}

func (f *fakeRepo) GetTeamStats(ctx context.Context, teamID string) ([]domain.TeamMember, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.check("GetTeamStats"); err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(f.members[teamID]))
	for id := range f.members[teamID] {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	result := make([]domain.TeamMember, 0, len(ids))
	for _, id := range ids {
		member := f.members[teamID][id]
		member.UserID = f.users[id].UserID
		result = append(result, member)
	}
	return result, nil
}

func (f *fakeRepo) sortedTeams() []domain.Team {
	ids := make([]string, 0, len(f.teams))
	for id := range f.teams {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	result := make([]domain.Team, 0, len(ids))
	for _, id := range ids {
		result = append(result, f.teams[id])
	}
	return result
}

// ── продюсер ────────────────────────────────────────────────────────────────

// fakeProducer запоминает опубликованное: тесты проверяют и сам факт
// публикации, и её порядок (сообщения по ивенту должны идти в том же порядке,
// в каком менялось состояние).
type fakeProducer struct {
	mu        sync.Mutex
	published []string
	failOn    map[string]error
}

func newFakeProducer() *fakeProducer {
	return &fakeProducer{failOn: map[string]error{}}
}

func (p *fakeProducer) record(kind string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err, ok := p.failOn[kind]; ok {
		return err
	}

	p.published = append(p.published, kind)
	return nil
}

func (p *fakeProducer) kinds() []string {
	p.mu.Lock()
	defer p.mu.Unlock()

	return append([]string(nil), p.published...)
}

func (p *fakeProducer) count(kind string) int {
	count := 0
	for _, k := range p.kinds() {
		if k == kind {
			count++
		}
	}
	return count
}

func (p *fakeProducer) PublishEventCreated(ctx context.Context, event domain.Event) error {
	return p.record("event.created")
}

func (p *fakeProducer) PublishEventCanceled(ctx context.Context, eventID, userCreateID string) error {
	return p.record("event.canceled")
}

func (p *fakeProducer) PublishEventTimeUpdated(ctx context.Context, eventID, userCreateID string, newTimeStart time.Time) error {
	return p.record("event.time_updated")
}

func (p *fakeProducer) PublishEventConfirmed(ctx context.Context, eventID string) error {
	return p.record("event.confirmed")
}

func (p *fakeProducer) PublishEventDeclined(ctx context.Context, eventID string) error {
	return p.record("event.declined")
}

func (p *fakeProducer) PublishEventStarted(ctx context.Context, eventID string, timeStart time.Time) error {
	return p.record("event.started")
}

func (p *fakeProducer) PublishEventFinished(ctx context.Context, eventID string, winnerSide string, timeFinish time.Time) error {
	return p.record("event.finished")
}

func (p *fakeProducer) PublishRentServer(ctx context.Context, eventID string, playersList []string, timeStart time.Time) error {
	return p.record("rent.server")
}

func (p *fakeProducer) PublishUserJoinedEvent(ctx context.Context, eventID, userID, clanID string, enemy bool, joinTime time.Time) error {
	return p.record("user.joined_event")
}

func (p *fakeProducer) PublishUserLeftEvent(ctx context.Context, eventID, userID string) error {
	return p.record("user.left_event")
}

func (p *fakeProducer) PublishUserRoleChanged(ctx context.Context, eventID, teamID, userID string, role domain.Role) error {
	return p.record("user.role_changed")
}

func (p *fakeProducer) PublishUserJoinedTeam(ctx context.Context, teamID, userID, userEventID string, role domain.Role) error {
	return p.record("user.joined_team")
}

func (p *fakeProducer) PublishUserLeftTeam(ctx context.Context, teamID, userID, userEventID string) error {
	return p.record("user.left_team")
}

func (p *fakeProducer) PublishTeamGameStarted(ctx context.Context, teamID string, gameNumber int64, timeStart time.Time) error {
	return p.record("team_game.started")
}

func (p *fakeProducer) PublishTeamGameFinished(ctx context.Context, teamID string, winner bool, timeFinish time.Time) error {
	return p.record("team_game.finished")
}

func (p *fakeProducer) PublishTeamMemberStatsAdded(ctx context.Context, teamID, userEventID string, kills, deaths, points, revival, destroyedVehicles int64) error {
	return p.record("team_member.stats_added")
}

// ── общая обвязка тестов ────────────────────────────────────────────────────

func newTestService(t *testing.T) (*eventService, *fakeRepo, *fakeProducer) {
	t.Helper()

	repo := newFakeRepo()
	producer := newFakeProducer()
	svc := &eventService{
		eventRepo:   repo,
		producer:    producer,
		eventLocks:  newEventLocker(),
		eventTimers: map[string]*eventTimer{},
	}

	// Цепочки таймеров, поставленные в тесте, ждут реального времени старта —
	// гасим их, чтобы горутины не жили дольше теста.
	t.Cleanup(func() {
		svc.timersMutex.Lock()
		defer svc.timersMutex.Unlock()

		for eventID, timer := range svc.eventTimers {
			timer.cancel()
			delete(svc.eventTimers, eventID)
		}
	})

	return svc, repo, producer
}

// ── помощники тестов ────────────────────────────────────────────────────────

func newID() string { return uuid.New().String() }

// futureStart — момент старта, проходящий проверку "не раньше чем через 45 минут".
func futureStart() time.Time { return time.Now().Add(2 * time.Hour) }

func kindName(kind domain.ErrorKind) string {
	switch kind {
	case domain.KindInternal:
		return "Internal"
	case domain.KindInvalidArgument:
		return "InvalidArgument"
	case domain.KindNotFound:
		return "NotFound"
	case domain.KindPermissionDenied:
		return "PermissionDenied"
	case domain.KindFailedPrecondition:
		return "FailedPrecondition"
	case domain.KindAlreadyExists:
		return "AlreadyExists"
	case domain.KindAborted:
		return "Aborted"
	case domain.KindUnavailable:
		return "Unavailable"
	}
	return "неизвестная категория"
}

// requireKind — ошибка должна быть не только "какая-то", но и нужной
// категории: именно из неё транспорт выбирает gRPC-код.
func requireKind(t *testing.T, err error, want domain.ErrorKind) {
	t.Helper()

	if err == nil {
		t.Fatalf("ожидалась ошибка категории %s, но вызов прошёл успешно", kindName(want))
	}

	if got := domain.KindOf(err); got != want {
		t.Fatalf("ожидалась категория %s, получена %s (%v)", kindName(want), kindName(got), err)
	}
}

func requireNoErr(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
}

// testEvent — созданный через публичную ручку ивент со всеми его участниками.
type testEvent struct {
	id          string
	creator     string
	creatorClan string
	enemy       string
	enemyClan   string
	timeStart   time.Time
}

func createTestEvent(t *testing.T, svc *eventService, targetGameCount int64) testEvent {
	t.Helper()

	ev := testEvent{
		id:          "",
		creator:     newID(),
		creatorClan: newID(),
		enemy:       newID(),
		enemyClan:   newID(),
		timeStart:   futureStart(),
	}

	err := svc.CreateEvent(context.Background(), ev.creator, ev.creatorClan, ev.enemy, ev.enemyClan, "test event", ev.timeStart, targetGameCount)
	requireNoErr(t, err)

	created, err := svc.GetLastEventByCreatorId(context.Background(), ev.creator)
	requireNoErr(t, err)
	if created == nil {
		t.Fatal("созданный ивент не найден")
	}

	ev.id = created.EventID
	return ev
}

// teams возвращает команды ивента по номеру игры: [своя сторона, вражеская].
func (e testEvent) teams(t *testing.T, svc *eventService, gameNumber int64) (ally, enemy string) {
	t.Helper()

	allTeams, err := svc.GetTeamsByEventID(context.Background(), e.id)
	requireNoErr(t, err)

	creatorUser, err := svc.eventRepo.GetUserByID(context.Background(), e.id, e.creator)
	requireNoErr(t, err)

	for _, team := range allTeams {
		if team.GameNumber != gameNumber {
			continue
		}
		if team.SideLeaderID == creatorUser.UserEventID {
			ally = team.TeamID
		} else {
			enemy = team.TeamID
		}
	}

	if ally == "" || enemy == "" {
		t.Fatalf("не найдены команды игры %d: ally=%q enemy=%q", gameNumber, ally, enemy)
	}

	return ally, enemy
}

// joinPlayer — игрок вступает в ивент за указанную сторону.
func joinPlayer(t *testing.T, svc *eventService, eventID, clanID string, enemy bool) string {
	t.Helper()

	userID := newID()
	requireNoErr(t, svc.JoinToEvent(context.Background(), eventID, userID, clanID, enemy))
	return userID
}

// startEvent проводит ивент тем же путём, что и таймеры: контрольная точка →
// подтверждение → старт ивента и первой игры.
func startEvent(t *testing.T, svc *eventService, repo *fakeRepo, ev testEvent) {
	t.Helper()

	event := repo.events[ev.id]
	if event.UserCount < MinPlayersRequired {
		event.UserCount = MinPlayersRequired
		repo.events[ev.id] = event
	}

	confirmed, err := svc.checkMinPlayers(context.Background(), ev.id)
	requireNoErr(t, err)
	if !confirmed {
		t.Fatal("ивент не подтвердился при достаточном количестве игроков")
	}

	requireNoErr(t, svc.startEventAndGames(context.Background(), ev.id))
}

// playGame доводит одну игру до конца: старт (если ещё не идёт) и финиш с
// указанным победителем.
func playGame(t *testing.T, svc *eventService, ally, enemy, winner string) {
	t.Helper()

	team, err := svc.GetTeamByID(context.Background(), ally)
	requireNoErr(t, err)

	if team.Status == domain.TeamStatusPending {
		requireNoErr(t, svc.StartTeamGame(context.Background(), ally, enemy))
	}

	requireNoErr(t, svc.FinishTeamGame(context.Background(), ally, enemy, winner))
}

// ── типовые сбои зависимостей ───────────────────────────────────────────────

var errDB = errors.New("база отказала")

// Сбой репозитория не должен превращаться в доменную ошибку: категории нет,
// значит транспорт отдаст Internal, а не "не найдено" или "нельзя".
func requireInternal(t *testing.T, err error) {
	t.Helper()
	requireKind(t, err, domain.KindInternal)
}

var errKafka = errors.New("kafka недоступна")

// waitFor ждёт выполнения условия: цепочка таймеров работает в своей
// горутине, поэтому результат появляется не сразу.
func waitFor(t *testing.T, cond func() bool, msg string) {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("не дождались: %s", msg)
}

func eventStatus(repo *fakeRepo, eventID string) domain.EventStatus {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	return repo.events[eventID].Status
}

func setEventFields(repo *fakeRepo, eventID string, apply func(event *domain.Event)) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	event := repo.events[eventID]
	apply(&event)
	repo.events[eventID] = event
}
