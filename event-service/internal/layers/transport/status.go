package transport

import (
	"context"
	"errors"
	"log"

	domain "event-service/internal/core/domain"

	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// fail превращает ошибку сервисного слоя в gRPC-статус: клиент различает
// причину по коду, а не по разбору строки. Наружу уходит только текст
// доменных ошибок — всё остальное (ошибки БД и прочие внутренние сбои)
// пишется в лог и отдаётся как Internal без подробностей.
func fail(method string, err error) error {
	if err == nil {
		return nil
	}

	switch domain.KindOf(err) {
	case domain.KindInvalidArgument:
		return status.Error(codes.InvalidArgument, err.Error())
	case domain.KindNotFound:
		return status.Error(codes.NotFound, err.Error())
	case domain.KindPermissionDenied:
		return status.Error(codes.PermissionDenied, err.Error())
	case domain.KindFailedPrecondition:
		return status.Error(codes.FailedPrecondition, err.Error())
	case domain.KindAlreadyExists:
		return status.Error(codes.AlreadyExists, err.Error())
	case domain.KindAborted:
		log.Printf("%s: %v", method, cause(err))
		return status.Error(codes.Aborted, err.Error())
	case domain.KindUnavailable:
		log.Printf("%s: %v", method, cause(err))
		return status.Error(codes.Unavailable, err.Error())
	}

	// Клиент оборвал запрос или вышел его дедлайн — это не сбой сервиса.
	if errors.Is(err, context.Canceled) {
		return status.Error(codes.Canceled, "request canceled")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, "request deadline exceeded")
	}

	log.Printf("%s: %v", method, err)

	return status.Error(codes.Internal, "internal error")
}

// cause — настоящая причина доменной ошибки для лога: клиенту уходит только
// обобщённый текст ("database is unavailable"), а в логе нужна конкретика.
func cause(err error) error {
	if inner := errors.Unwrap(err); inner != nil {
		return inner
	}

	return err
}
