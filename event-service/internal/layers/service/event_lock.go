package service

import (
	"context"
	"sync"
)

// eventLocker — по мьютексу на каждый event_id. Все изменяющие операции над
// одним ивентом (ручки и шаги таймеров: контроль, аренда сервера, старт)
// выполняются под ним строго по очереди: транзакция в БД, затем публикация в
// Kafka. За счёт этого:
//   - сообщения по одному ивенту уходят в Kafka в том же порядке, в каком
//     изменения закоммичены (join → leave не превратится в leave → join);
//   - шаг таймера не может выполниться посреди CancelEvent/UpdateTimeEvent —
//     те успевают отменить цепочку таймеров до того, как шаг начнёт работу.
//
// Мьютекс живёт только внутри процесса. Согласованность данных в БД (в том
// числе при нескольких инстансах сервиса) держат SERIALIZABLE-транзакции с
// SELECT ... FOR UPDATE по строке events — см. lockEventRow.
type eventLocker struct {
	mu    sync.Mutex
	locks map[string]*eventLock
}

type eventLock struct {
	// ch с буфером 1: занятый слот — мьютекс захвачен. Канал вместо
	// sync.Mutex, чтобы ожидание можно было прервать через ctx.
	ch chan struct{}
	// refs — сколько горутин держат или ждут этот мьютекс; когда их не
	// остаётся, запись удаляется из locks, и мапа не растёт бесконечно.
	refs int
}

func newEventLocker() *eventLocker {
	return &eventLocker{locks: make(map[string]*eventLock)}
}

// lock захватывает мьютекс ивента. Ожидание прерывается при отмене ctx
// (дедлайн запроса или отмена цепочки таймеров) — тогда возвращается
// ctx.Err(). Возвращённый unlock нужно вызвать ровно один раз.
func (l *eventLocker) lock(ctx context.Context, eventID string) (unlock func(), err error) {
	l.mu.Lock()
	el, ok := l.locks[eventID]
	if !ok {
		el = &eventLock{ch: make(chan struct{}, 1)}
		l.locks[eventID] = el
	}
	el.refs++
	l.mu.Unlock()

	select {
	case el.ch <- struct{}{}:
		return func() {
			<-el.ch
			l.release(eventID, el)
		}, nil
	case <-ctx.Done():
		l.release(eventID, el)
		return nil, ctx.Err()
	}
}

func (l *eventLocker) release(eventID string, el *eventLock) {
	l.mu.Lock()
	defer l.mu.Unlock()

	el.refs--
	if el.refs == 0 {
		delete(l.locks, eventID)
	}
}
