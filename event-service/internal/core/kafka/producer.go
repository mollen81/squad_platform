package kafka

import (
	"context"
	"encoding/json"
	"time"

	"event-service/internal/core/domain"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafka.LeastBytes{},
			Compression:            kafka.Lz4,
			RequiredAcks:           kafka.RequireOne,
			Async:                  false,
			AllowAutoTopicCreation: true,
			BatchSize:              100,
			BatchTimeout:           10 * time.Millisecond,
			ReadTimeout:            10 * time.Second,
			WriteTimeout:           10 * time.Second,
		},
	}
}

func (p *Producer) PublishEventCreated(ctx context.Context, event domain.Event) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":            "event_created",
		"event_id":        event.EventID,
		"name":            event.Name,
		"user_create_id":  event.UserCreateID,
		"time_start":      event.TimeStart,
		"create_time":     event.CreateTime,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.EventID),
		Value: data,
	})
}

func (p *Producer) PublishEventDeleted(ctx context.Context, eventID, userCreateID string) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":           "event_deleted",
		"event_id":       eventID,
		"user_create_id": userCreateID,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(eventID),
		Value: data,
	})
}

func (p *Producer) PublishEventTimeUpdated(ctx context.Context, eventID, userCreateID string, newTimeStart time.Time) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":           "event_time_updated",
		"event_id":       eventID,
		"user_create_id": userCreateID,
		"new_time_start": newTimeStart,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(eventID),
		Value: data,
	})
}

func (p *Producer) PublishUserJoinedEvent(ctx context.Context, eventID, userID string, joinTime time.Time) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":      "user_joined_event",
		"event_id":  eventID,
		"user_id":   userID,
		"join_time": joinTime,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(eventID),
		Value: data,
	})
}

func (p *Producer) PublishUserLeftEvent(ctx context.Context, eventID, userID string) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":     "user_left_event",
		"event_id": eventID,
		"user_id":  userID,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(eventID),
		Value: data,
	})
}

func (p *Producer) PublishGameCreated(ctx context.Context, game domain.Game) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":       "game_created",
		"game_id":    game.GameID,
		"event_id":   game.EventID,
		"map_name":   game.MapName,
		"time_start": game.TimeStart,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(game.EventID),
		Value: data,
	})
}

func (p *Producer) PublishGameWinnerUpdated(ctx context.Context, gameID, winnerTeamID string) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":           "game_winner_updated",
		"game_id":        gameID,
		"winner_team_id": winnerTeamID,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(gameID),
		Value: data,
	})
}

func (p *Producer) PublishGameLoserUpdated(ctx context.Context, gameID, loserTeamID string) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":          "game_loser_updated",
		"game_id":       gameID,
		"loser_team_id": loserTeamID,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(gameID),
		Value: data,
	})
}

func (p *Producer) PublishGameFinished(ctx context.Context, gameID string, timeFinish time.Time) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":        "game_finished",
		"game_id":     gameID,
		"time_finish": timeFinish,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(gameID),
		Value: data,
	})
}

func (p *Producer) PublishGameDeleted(ctx context.Context, gameID string) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":    "game_deleted",
		"game_id": gameID,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(gameID),
		Value: data,
	})
}

func (p *Producer) PublishUserStatsAdded(ctx context.Context, stats domain.GameUserStats) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":      "user_stats_added",
		"stats_id":  stats.GameUserStatsID,
		"game_id":   stats.Game.GameID,
		"user_id":   stats.User.UserID,
		"kills":     stats.Kills,
		"deaths":    stats.Deaths,
		"points":    stats.Points,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(stats.Game.GameID),
		Value: data,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
