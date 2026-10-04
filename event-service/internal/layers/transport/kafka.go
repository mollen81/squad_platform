package transport

import (
	"context"
	"encoding/json"
	"log"

	domain "event-service/internal/core/domain"
	service "event-service/internal/layers/service"
)

// Типы сообщений vps-сервиса: сервер под ивент покупает и разворачивает он.
const (
	messageTypeVPSPurchased = "vps.purchased"
	messageTypeVPSDeployed  = "vps.deployed"
)

// KafkaHandler — входящий транспорт для сообщений vps-сервиса: разбирает JSON и
// зовёт сервисный слой, как это делает gRPC-хендлер для запросов. Сам ничего не
// решает: что считать повторной доставкой и в каком статусе ивент готов принять
// сообщение — дело сервиса.
type KafkaHandler struct {
	eventService service.EventService
}

func NewKafkaHandler(eventService service.EventService) *KafkaHandler {
	return &KafkaHandler{eventService: eventService}
}

// vpsMessage — общий вид сообщений vps-сервиса:
//
//	{"type":"vps.purchased","event_id":"…","server_ip":"…","server_password":"…"}
//	{"type":"vps.deployed","event_id":"…","server_ip":"…"}
//
// Привязка к ивенту идёт по event_id (vps-сервис получает его из нашего
// server.rent), server_ip в vps.deployed нужен только для сверки с купленным.
// Времени в
// vps.deployed нет — моментом готовности сервис считает время получения.
type vpsMessage struct {
	Type           string `json:"type"`
	EventID        string `json:"event_id"`
	ServerIP       string `json:"server_ip"`
	ServerPassword string `json:"server_password"`
}

// Handle обрабатывает одно сообщение. Ошибка с категорией Unavailable/Aborted/
// Internal означает "повтори позже", остальные — "это сообщение обработать
// нельзя" (см. consumer.Run).
func (h *KafkaHandler) Handle(ctx context.Context, value []byte) error {
	var message vpsMessage
	if err := json.Unmarshal(value, &message); err != nil {
		return domain.InvalidArgument("не удалось разобрать сообщение: %v", err)
	}

	switch message.Type {
	case messageTypeVPSPurchased:
		return h.eventService.ServerPurchased(ctx, message.EventID, message.ServerIP, message.ServerPassword)

	case messageTypeVPSDeployed:
		return h.eventService.ServerDeployed(ctx, message.EventID, message.ServerIP)

	default:
		// Чужое сообщение в том же топике — не наша забота.
		log.Printf("kafka: пропущено сообщение неизвестного типа %q", message.Type)
		return nil
	}
}
