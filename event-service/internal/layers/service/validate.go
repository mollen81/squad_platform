package service

import (
	"strings"
	"unicode/utf8"

	domain "event-service/internal/core/domain"

	uuid "github.com/google/uuid"
)

const (
	// Максимальная длина названия ивента
	maxEventNameLength = 100
	// Потолок для одного показателя статистики за игру — защита от опечаток
	// и мусора: столько не набивают даже за очень длинную игру
	maxStatValue = 100000
	// Максимальная длина адреса сервера: с запасом на IPv6 и порт
	maxServerIPLength = 64
)

// Проверки запроса живут в сервисном слое, а не полагаются на БД: иначе
// клиенту вместо внятной ошибки прилетало бы "invalid input syntax for type
// uuid" или испорченные ими же данные (например, отрицательные киллы в
// суммарной статистике команды).

// validateID — все идентификаторы в API это UUID, приходящие из других
// сервисов.
func validateID(field, value string) error {
	if value == "" {
		return domain.InvalidArgument("%s is required", field)
	}

	if err := uuid.Validate(value); err != nil {
		return domain.InvalidArgument("%s must be a valid uuid", field)
	}

	return nil
}

// normalizeEventName приводит название к тому виду, в котором оно ляжет в БД.
func normalizeEventName(name string) (string, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return "", domain.InvalidArgument("event_name is required")
	}

	if utf8.RuneCountInString(name) > maxEventNameLength {
		return "", domain.InvalidArgument("event_name must be at most %d characters", maxEventNameLength)
	}

	return name, nil
}

// validateSearchName — название как параметр поиска: пустую строку не ищем,
// но длину не ограничиваем.
func validateSearchName(name string) error {
	if strings.TrimSpace(name) == "" {
		return domain.InvalidArgument("event_name is required")
	}

	return nil
}

// normalizeServerIP — адрес сервера из vps-сервиса. Это просто строка (IPv4,
// IPv6, возможно с портом), поэтому проверяем только непустоту и разумную
// длину: разбирать формат здесь — значит рисковать отказом на адресе, который
// для vps-сервиса валиден. Обрезка по краям нужна, чтобы адрес из
// vps.purchased и vps.deployed сравнивался одинаково.
func normalizeServerIP(value string) (string, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return "", domain.InvalidArgument("server_ip is required")
	}

	if utf8.RuneCountInString(value) > maxServerIPLength {
		return "", domain.InvalidArgument("server_ip must be at most %d characters", maxServerIPLength)
	}

	return value, nil
}

// validateStat — один показатель статистики игрока за игру.
func validateStat(field string, value int64) error {
	if value < 0 {
		return domain.InvalidArgument("%s must not be negative", field)
	}

	if value > maxStatValue {
		return domain.InvalidArgument("%s must be at most %d", field, maxStatValue)
	}

	return nil
}

// validateAssignableRole — роли, которые вообще можно выдать игроку:
// squad_leader в команде один, и это сайд-лидер стороны (его ставит
// CreateEvent), поэтому выдать эту роль через API нельзя.
func validateAssignableRole(role domain.Role) error {
	switch role {
	case domain.RolePlayer, domain.RoleSideLeader:
		return nil
	case domain.RoleSquadLeader:
		return domain.InvalidArgument("squad_leader role cannot be assigned: it belongs to the side leader")
	default:
		return domain.InvalidArgument("invalid role %q: only side_leader or player can be set", role)
	}
}
