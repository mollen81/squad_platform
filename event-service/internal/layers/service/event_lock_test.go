package service

import (
	"context"
	"sync"
	"testing"
	"time"

	domain "event-service/internal/core/domain"
)

func TestEventLockerSerializesSameEvent(t *testing.T) {
	locker := newEventLocker()
	eventID := newID()

	var (
		mu      sync.Mutex
		busy    bool
		overlap bool
		done    sync.WaitGroup
	)

	for i := 0; i < 20; i++ {
		done.Add(1)
		go func() {
			defer done.Done()

			unlock, err := locker.lock(context.Background(), eventID)
			if err != nil {
				t.Errorf("не удалось взять мьютекс: %v", err)
				return
			}
			defer unlock()

			mu.Lock()
			if busy {
				overlap = true
			}
			busy = true
			mu.Unlock()

			time.Sleep(time.Millisecond)

			mu.Lock()
			busy = false
			mu.Unlock()
		}()
	}

	done.Wait()

	if overlap {
		t.Error("две операции над одним ивентом выполнялись одновременно")
	}
}

func TestEventLockerDoesNotBlockOtherEvents(t *testing.T) {
	locker := newEventLocker()

	unlockFirst, err := locker.lock(context.Background(), newID())
	if err != nil {
		t.Fatalf("не удалось взять мьютекс: %v", err)
	}
	defer unlockFirst()

	// Другой ивент не должен ждать: если бы ждал, тест упал бы по таймауту.
	done := make(chan struct{})
	go func() {
		unlockSecond, err := locker.lock(context.Background(), newID())
		if err == nil {
			unlockSecond()
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Error("операция над другим ивентом заблокирована чужим мьютексом")
	}
}

func TestEventLockerRespectsContext(t *testing.T) {
	locker := newEventLocker()
	eventID := newID()

	unlock, err := locker.lock(context.Background(), eventID)
	if err != nil {
		t.Fatalf("не удалось взять мьютекс: %v", err)
	}
	defer unlock()

	// Ожидание прерывается по дедлайну запроса, а не висит до победного.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	if _, err := locker.lock(ctx, eventID); err == nil {
		t.Error("ожидание мьютекса должно прерываться отменой контекста")
	}
}

func TestEventLockerCleansUpReleasedLocks(t *testing.T) {
	locker := newEventLocker()
	eventID := newID()

	unlock, err := locker.lock(context.Background(), eventID)
	if err != nil {
		t.Fatalf("не удалось взять мьютекс: %v", err)
	}
	unlock()

	locker.mu.Lock()
	defer locker.mu.Unlock()

	// Иначе мапа росла бы на каждый когда-либо созданный ивент.
	if len(locker.locks) != 0 {
		t.Errorf("после освобождения осталось %d записей", len(locker.locks))
	}
}

func TestEventLockerReusesLockForWaiters(t *testing.T) {
	locker := newEventLocker()
	eventID := newID()

	unlock, err := locker.lock(context.Background(), eventID)
	if err != nil {
		t.Fatalf("не удалось взять мьютекс: %v", err)
	}

	waiting := make(chan func(), 1)
	go func() {
		secondUnlock, err := locker.lock(context.Background(), eventID)
		if err != nil {
			t.Errorf("ожидающий не получил мьютекс: %v", err)
			return
		}
		waiting <- secondUnlock
	}()

	// Пока первый держит мьютекс, запись жива и на неё ссылается ожидающий.
	time.Sleep(20 * time.Millisecond)
	unlock()

	select {
	case secondUnlock := <-waiting:
		secondUnlock()
	case <-time.After(2 * time.Second):
		t.Fatal("ожидающий так и не получил мьютекс")
	}

	locker.mu.Lock()
	defer locker.mu.Unlock()

	if len(locker.locks) != 0 {
		t.Errorf("после освобождения осталось %d записей", len(locker.locks))
	}
}

// Ожидание мьютекса ивента прерывается вместе с запросом: клиент, который уже
// отвалился, не должен занимать очередь.
func TestCallsAbortWhenRequestCanceledWhileWaitingForLock(t *testing.T) {
	cases := []struct {
		name string
		call func(ctx context.Context, svc *eventService, ev testEvent, teamID string) error
	}{
		{"CancelEvent", func(ctx context.Context, svc *eventService, ev testEvent, _ string) error {
			return svc.CancelEvent(ctx, ev.id, ev.creator)
		}},
		{"JoinToEvent", func(ctx context.Context, svc *eventService, ev testEvent, _ string) error {
			return svc.JoinToEvent(ctx, ev.id, newID(), newID(), false)
		}},
		{"LeaveEvent", func(ctx context.Context, svc *eventService, ev testEvent, _ string) error {
			return svc.LeaveEvent(ctx, newID(), ev.id)
		}},
		{"UpdateTimeEvent", func(ctx context.Context, svc *eventService, ev testEvent, _ string) error {
			return svc.UpdateTimeEvent(ctx, ev.id, ev.creator, futureStart())
		}},
		{"JoinUserToTeam", func(ctx context.Context, svc *eventService, ev testEvent, teamID string) error {
			return svc.JoinUserToTeam(ctx, teamID, newID(), domain.RolePlayer)
		}},
		{"RemoveUserFromTeam", func(ctx context.Context, svc *eventService, ev testEvent, teamID string) error {
			return svc.RemoveUserFromTeam(ctx, teamID, newID())
		}},
		{"SetRole", func(ctx context.Context, svc *eventService, ev testEvent, teamID string) error {
			return svc.SetRole(ctx, teamID, ev.creator, newID(), domain.RolePlayer)
		}},
		{"StartTeamGame", func(ctx context.Context, svc *eventService, ev testEvent, teamID string) error {
			_, enemy := ev.teamsQuiet(svc, 1)
			return svc.StartTeamGame(ctx, teamID, enemy)
		}},
		{"FinishTeamGame", func(ctx context.Context, svc *eventService, ev testEvent, teamID string) error {
			_, enemy := ev.teamsQuiet(svc, 1)
			return svc.FinishTeamGame(ctx, teamID, enemy, teamID)
		}},
		{"AddTeamMemberStats", func(ctx context.Context, svc *eventService, ev testEvent, teamID string) error {
			return svc.AddTeamMemberStats(ctx, teamID, ev.creator, 1, 1, 1, 1, 1)
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			svc, _, _ := newTestService(t)
			ev := createTestEvent(t, svc, 1)
			ally, _ := ev.teams(t, svc, 1)

			// Ивент занят другой операцией.
			unlock, err := svc.eventLocks.lock(context.Background(), ev.id)
			requireNoErr(t, err)
			defer unlock()

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			done := make(chan error, 1)
			go func() { done <- c.call(ctx, svc, ev, ally) }()

			select {
			case err := <-done:
				if err == nil {
					t.Error("вызов с отменённым запросом должен вернуть ошибку, а не ждать мьютекс")
				}
			case <-time.After(2 * time.Second):
				t.Error("вызов завис на мьютексе, хотя запрос уже отменён")
			}
		})
	}
}

// teams в этом тесте вызывается из горутины без *testing.T — отдельный
// вариант без падения теста.
func (e testEvent) teamsQuiet(svc *eventService, gameNumber int64) (ally, enemy string) {
	allTeams, err := svc.GetTeamsByEventID(context.Background(), e.id)
	if err != nil {
		return "", ""
	}

	creatorUser, err := svc.eventRepo.GetUserByID(context.Background(), e.id, e.creator)
	if err != nil {
		return "", ""
	}

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
	return ally, enemy
}
