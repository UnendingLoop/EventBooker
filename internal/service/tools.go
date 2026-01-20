package service

import (
	"regexp"
	"strings"
	"time"

	"github.com/UnendingLoop/EventBooker/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func validateNormalizeUser(u *model.User) error {
	// Проверка роли
	if u.Role != model.RoleAdmin && u.Role != model.RoleUser {
		return model.ErrIncorrectUserRole
	}
	// Проверка имейл
	matchEmail := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !matchEmail.MatchString(u.Email) {
		return model.ErrIncorrectEmail
	}
	u.Email = strings.TrimSpace(u.Email)
	u.Email = strings.ToLower(u.Email)

	// Проверка телефона
	if u.Tel != "" {
		u.Tel = normalizePhone(u.Tel)
		matchPhone := regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
		if !matchPhone.MatchString(u.Tel) {
			return model.ErrIncorrectPhone
		}

	}

	// Генерация хэша из пароля
	passHash, _ := bcrypt.GenerateFromPassword([]byte(u.PassHash), bcrypt.DefaultCost)
	u.PassHash = string(passHash)

	return nil
}

func normalizePhone(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	if strings.HasPrefix(s, "00") {
		s = "+" + s[2:]
	}
	return s
}

func validateNormalizeEvent(event *model.Event) error {
	if event.Title == "" || event.TotalSeats <= 0 || event.BookWindow <= 0 {
		return model.ErrEmptyEventInfo
	}
	if event.EventDate.UTC().Before(time.Now().UTC()) {
		return model.ErrIncorrectEventTime
	}
	now := time.Now().UTC()
	event.Created = &now
	event.AvailSeats = event.TotalSeats
	event.Status = model.EventStatusActual

	return nil
}
