package kafka

import (
	"context"
	"errors"
	"log"
	"time"

	"event-service/internal/core/domain"

	kafka "github.com/segmentio/kafka-go"
)

const (
	// retryDelay — пауза перед повторной обработкой сообщения, если помешала
	// временная причина (недоступна БД, конфликт транзакций).
	retryDelay = 2 * time.Second
	// maxRetryDelay — предел этой паузы: если зависимость лежит долго, чаще
	// чем раз в полминуты дёргать её незачем.
	maxRetryDelay = 30 * time.Second
)

// Handler обрабатывает одно сообщение. Ошибку возвращает так, чтобы по её
// категории (domain.KindOf) было понятно, имеет ли смысл повтор.
type Handler func(ctx context.Context, value []byte) error

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
			// FirstOffset, а не LastOffset: если сервис лежал в момент, когда
			// vps-сервис доложил о покупке, сообщение нужно прочитать, а не
			// пропустить.
			StartOffset: kafka.FirstOffset,
			MinBytes:    1,
			MaxBytes:    10e6,
			MaxWait:     500 * time.Millisecond,
			// Оффсет коммитим сами и только после успешной обработки —
			// иначе сообщение о купленном сервере может потеряться.
			CommitInterval:   0,
			ReadBatchTimeout: 10 * time.Second,
		}),
	}
}

// Run читает сообщения до отмены ctx. Оффсет двигается только после успешной
// обработки: временная ошибка приводит к повтору того же сообщения (сообщений
// тут единицы, поэтому лучше задержать очередь, чем потерять покупку сервера),
// а сообщение, которое не обработается никогда — неизвестный тип, битый JSON,
// ивент не в том статусе, — пропускается с записью в лог.
func (c *Consumer) Run(ctx context.Context, handle Handler) {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("kafka consumer: fetch: %v", err)
			if !sleep(ctx, retryDelay) {
				return
			}
			continue
		}

		if !c.handleWithRetries(ctx, message, handle) {
			return
		}

		if err := c.reader.CommitMessages(ctx, message); err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("kafka consumer: commit offset %d: %v", message.Offset, err)
		}
	}
}

// handleWithRetries возвращает false, только если пора останавливаться
// (отменён ctx).
func (c *Consumer) handleWithRetries(ctx context.Context, message kafka.Message, handle Handler) bool {
	delay := retryDelay

	for {
		err := handle(ctx, message.Value)
		if err == nil {
			return true
		}

		if ctx.Err() != nil {
			return false
		}

		if !retryable(err) {
			// Повтор не поможет: сообщение не для нас, испорчено или пришло
			// не вовремя. Пропускаем, иначе оно навсегда заблокирует очередь.
			log.Printf("kafka consumer: сообщение с оффсетом %d пропущено: %v", message.Offset, err)
			return true
		}

		log.Printf("kafka consumer: повтор сообщения с оффсетом %d через %v: %v", message.Offset, delay, err)
		if !sleep(ctx, delay) {
			return false
		}

		if delay < maxRetryDelay {
			delay *= 2
		}
	}
}

// retryable — помешала временная причина: недоступна БД, конфликт транзакций
// или неизвестный сбой внутри сервиса.
func retryable(err error) bool {
	switch domain.KindOf(err) {
	case domain.KindUnavailable, domain.KindAborted, domain.KindInternal:
		return true
	default:
		return false
	}
}

func sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}

func (c *Consumer) Close() error {
	if err := c.reader.Close(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
