package service

import (
	"strings"
	"testing"

	domain "event-service/internal/core/domain"
)

func TestValidateID(t *testing.T) {
	if err := validateID("event_id", newID()); err != nil {
		t.Errorf("корректный uuid отвергнут: %v", err)
	}

	for _, value := range []string{"", "   ", "не-uuid", "123", "00000000-0000-0000-0000-00000000000"} {
		err := validateID("event_id", value)
		requireKind(t, err, domain.KindInvalidArgument)

		if !strings.Contains(err.Error(), "event_id") {
			t.Errorf("в тексте ошибки нет имени поля: %v", err)
		}
	}
}

func TestNormalizeEventName(t *testing.T) {
	name, err := normalizeEventName("  Клановая заруба  ")
	if err != nil {
		t.Fatalf("корректное название отвергнуто: %v", err)
	}
	if name != "Клановая заруба" {
		t.Errorf("название не обрезано по краям: %q", name)
	}

	for _, value := range []string{"", "   ", "\t\n"} {
		_, err := normalizeEventName(value)
		requireKind(t, err, domain.KindInvalidArgument)
	}

	// Длина считается в символах, а не в байтах: кириллица не должна
	// отвергаться раньше времени.
	longEnough := strings.Repeat("я", maxEventNameLength)
	if _, err := normalizeEventName(longEnough); err != nil {
		t.Errorf("название из %d символов отвергнуто: %v", maxEventNameLength, err)
	}

	_, err = normalizeEventName(strings.Repeat("я", maxEventNameLength+1))
	requireKind(t, err, domain.KindInvalidArgument)
}

func TestValidateSearchName(t *testing.T) {
	if err := validateSearchName("турнир"); err != nil {
		t.Errorf("корректный поисковый запрос отвергнут: %v", err)
	}

	for _, value := range []string{"", "  "} {
		requireKind(t, validateSearchName(value), domain.KindInvalidArgument)
	}
}

func TestValidateStat(t *testing.T) {
	for _, value := range []int64{0, 1, maxStatValue} {
		if err := validateStat("kills", value); err != nil {
			t.Errorf("значение %d отвергнуто: %v", value, err)
		}
	}

	for _, value := range []int64{-1, -100, maxStatValue + 1} {
		err := validateStat("kills", value)
		requireKind(t, err, domain.KindInvalidArgument)

		if !strings.Contains(err.Error(), "kills") {
			t.Errorf("в тексте ошибки нет имени показателя: %v", err)
		}
	}
}

func TestValidateAssignableRole(t *testing.T) {
	for _, role := range []domain.Role{domain.RolePlayer, domain.RoleSideLeader} {
		if err := validateAssignableRole(role); err != nil {
			t.Errorf("роль %s должна назначаться: %v", role, err)
		}
	}

	// squad_leader принадлежит сайд-лидеру стороны и через API не выдаётся.
	requireKind(t, validateAssignableRole(domain.RoleSquadLeader), domain.KindInvalidArgument)
	requireKind(t, validateAssignableRole(domain.Role("")), domain.KindInvalidArgument)
	requireKind(t, validateAssignableRole(domain.Role("captain")), domain.KindInvalidArgument)
}
