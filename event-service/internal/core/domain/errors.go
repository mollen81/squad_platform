package domain

import (
	"errors"
	"fmt"
)

// ErrorKind — категория доменной ошибки. Транспорт по ней выбирает gRPC-код
// (см. toStatus), поэтому сервисный слой ничего не знает про gRPC, а наружу
// не утекают внутренние ошибки: всё без категории считается внутренним и
// отдаётся клиенту как Internal без подробностей.
type ErrorKind int

const (
	// KindInternal — сбой, о котором клиенту знать нечего (ошибка БД и т.п.)
	KindInternal ErrorKind = iota
	// KindInvalidArgument — запрос неверен сам по себе: пустой или не-UUID
	// идентификатор, недопустимая роль, отрицательная статистика
	KindInvalidArgument
	// KindNotFound — объекта нет: ивент, команда, участник
	KindNotFound
	// KindPermissionDenied — вызывающий не имеет права на это действие
	KindPermissionDenied
	// KindFailedPrecondition — запрос корректен, но состояние не позволяет:
	// не тот статус ивента, игра уже началась, команда заполнена
	KindFailedPrecondition
	// KindAlreadyExists — объект уже есть: игрок уже в ивенте или в команде
	KindAlreadyExists
	// KindAborted — конкурентный конфликт, который сервис не смог разрешить
	// сам (транзакция не прошла даже после повторов): запрос имеет смысл
	// повторить целиком
	KindAborted
	// KindUnavailable — недоступна зависимость (база): сбой временный,
	// запрос имеет смысл повторить
	KindUnavailable
)

// Error — доменная ошибка с категорией. Текст уходит клиенту как есть,
// поэтому в нём не должно быть внутренних подробностей: настоящая причина
// (ошибка драйвера БД и т.п.) прячется в cause и видна только в логе.
type Error struct {
	Kind  ErrorKind
	msg   string
	cause error
}

func (e *Error) Error() string { return e.msg }

// Unwrap отдаёт исходную причину — для логов, наружу она не уходит.
func (e *Error) Unwrap() error { return e.cause }

func newError(kind ErrorKind, format string, a ...any) error {
	return &Error{Kind: kind, msg: fmt.Sprintf(format, a...)}
}

// Wrap помечает ошибку категорией, сохраняя исходную причину.
func Wrap(kind ErrorKind, cause error, format string, a ...any) error {
	return &Error{Kind: kind, msg: fmt.Sprintf(format, a...), cause: cause}
}

func InvalidArgument(format string, a ...any) error {
	return newError(KindInvalidArgument, format, a...)
}

func NotFound(format string, a ...any) error {
	return newError(KindNotFound, format, a...)
}

func PermissionDenied(format string, a ...any) error {
	return newError(KindPermissionDenied, format, a...)
}

func FailedPrecondition(format string, a ...any) error {
	return newError(KindFailedPrecondition, format, a...)
}

func AlreadyExists(format string, a ...any) error {
	return newError(KindAlreadyExists, format, a...)
}

func Aborted(format string, a ...any) error {
	return newError(KindAborted, format, a...)
}

func Unavailable(format string, a ...any) error {
	return newError(KindUnavailable, format, a...)
}

// KindOf — категория ошибки; всё, что не доменная ошибка, считается
// внутренним сбоем.
func KindOf(err error) ErrorKind {
	var domainErr *Error
	if errors.As(err, &domainErr) {
		return domainErr.Kind
	}

	return KindInternal
}
