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
			ReadTimeout:            30 * time.Second,
			WriteTimeout:           30 * time.Second,
		},
	}
}

func (p *Producer) PublishEventCreated(ctx context.Context, event domain.Event) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":            "event.created",
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

func (p *Producer) PublishEventCanceled(ctx context.Context, eventID, userCreateID string) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":           "event.canceled",
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
		"type":           "event.time_updated",
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

func (p *Producer) PublishUserJoinedEvent(ctx context.Context, eventID, userID, clanID string, enemy bool, joinTime time.Time) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":      "user.joined_event",
		"event_id":  eventID,
		"user_id":   userID,
		"clan_id":   clanID,
		"enemy":     enemy,
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
		"type":     "user.left_event",
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

// PublishGameWinnerUpdated/PublishGameLoserUpdated/PublishGameFinished удалены:
// были завязаны на game_id из удалённой таблицы games. Их место заняли
// PublishTeamGameStarted/PublishTeamGameFinished ниже (team_id вместо game_id).

func (p *Producer) PublishTeamGameStarted(ctx context.Context, teamID string, gameNumber int64, timeStart time.Time) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":        "team_game.started",
		"team_id":     teamID,
		"game_number": gameNumber,
		"time_start":  timeStart,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(teamID),
		Value: data,
	})
}

func (p *Producer) PublishTeamGameFinished(ctx context.Context, teamID string, winner bool, timeFinish time.Time) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":        "team_game.finished",
		"team_id":     teamID,
		"winner":      winner,
		"time_finish": timeFinish,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(teamID),
		Value: data,
	})
}

func (p *Producer) PublishEventFinished(ctx context.Context, eventID string, winnerSide string, timeFinish time.Time) error {
	data, err := json.Marshal(map[string]interface{}{
		"type": "event.finished",
		"event_id": eventID,
		"winner_side": winnerSide,
		"time_finish": timeFinish,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key: []byte(eventID),
		Value: data,
	})
}

// PublishUserStatsAdded удалён: принимал domain.GameUserStats, которого больше нет
// (таблица game_user_stats убрана из схемы, статистика теперь в team_members).

func (p *Producer) PublishTeamMemberStatsAdded(ctx context.Context, teamID, userEventID string, kills, deaths, points, revival, destroyedVehicles int64) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":               "team_member.stats_added",
		"team_id":            teamID,
		"user_event_id":      userEventID,
		"kills":              kills,
		"deaths":             deaths,
		"points":             points,
		"revival":            revival,
		"destroyed_vehicles": destroyedVehicles,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(teamID),
		Value: data,
	})
}

func (p *Producer) PublishEventConfirmed(ctx context.Context, eventID string) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":     "event.confirmed",
		"event_id": eventID,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(eventID),
		Value: data,
	})
}

func (p *Producer) PublishEventDeclined(ctx context.Context, eventID string) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":     "event.declined",
		"event_id": eventID,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(eventID),
		Value: data,
	})
}

func (p *Producer) PublishEventStarted(ctx context.Context, eventID string, timeStart time.Time) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":       "event.started",
		"event_id":   eventID,
		"time_start": timeStart,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(eventID),
		Value: data,
	})
}

func (p *Producer) PublishUserJoinedTeam(ctx context.Context, teamID, userID, userEventID string, role domain.Role) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":          "user.joined_team",
		"team_id":       teamID,
		"user_id":       userID,
		"user_event_id": userEventID,
		"role":          string(role),
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(teamID),
		Value: data,
	})
}

func (p *Producer) PublishUserLeftTeam(ctx context.Context, teamID, userID, userEventID string) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":          "user.left_team",
		"team_id":       teamID,
		"user_id":       userID,
		"user_event_id": userEventID,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(teamID),
		Value: data,
	})
}

func (p *Producer) PublishUserRoleChanged(ctx context.Context, eventID, userID string, role domain.Role) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":     "user.role_changed",
		"event_id": eventID,
		"user_id":  userID,
		"role":     string(role),
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(eventID),
		Value: data,
	})
}

// PublishTeamsCreated/PublishTeamConfirmed удалены: были частью старого флоу
// CreateTeamsForEvent/подтверждения команд (team.is_confirmed), которого больше
// нет — команды на все запланированные игры создаются сразу внутри CreateEvent.

func (p *Producer) PublishRentServer(ctx context.Context, eventID string, playersList []string, timeStart time.Time) error {
	data, err := json.Marshal(map[string]interface{}{
		"type":         "rent.server",
		"event_id":     eventID,
		"players_list": playersList,
		"time_start":   timeStart,
	})
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(eventID),
		Value: data,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
