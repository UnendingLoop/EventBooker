package service

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/UnendingLoop/EventBooker/internal/model"
)

var (
	ErrIncorrectUserRole  = errors.New("incorrect user role is provided")
	ErrIncorrectEmail     = errors.New("incorrect email provided")
	ErrIncorrectPhone     = errors.New("incorrect telephone number provided")
	ErrCommon500          = errors.New("something went wrong. Try again later")
	ErrEmptyBookInfo      = errors.New("incomplete data to book event")
	ErrNoSeatsAvailable   = errors.New("no more seats to book for this event")
	ErrExpiredEvent       = errors.New("the event you are trying to book has expired")
	ErrIncorrectEventTime = errors.New("event date cannot be in the past")
	ErrEmptyEventInfo     = errors.New("incomplete data to create event")
	ErrBookIsCancelled    = errors.New("the book is already cancelled")
)

func validateNormalizeUser(u *model.User) error {
	/*	User struct {
		ID       int
		Role     string
		Created  time.Time
		Name     string
		Surname  string
		Tel      string
		Email    string
		PassHash string
	}*/
	// Проверка роли
	if u.Role != model.RoleAdmin && u.Role != model.RoleUser {
		return ErrIncorrectUserRole
	}
	// Проверка имейл
	matchEmail := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !matchEmail.MatchString(u.Email) {
		return ErrIncorrectEmail
	}
	u.Email = strings.TrimSpace(u.Email)
	u.Email = strings.ToLower(u.Email)

	// Проверка телефона
	if u.Tel != "" {
		u.Tel = normalizePhone(u.Tel)
		matchPhone := regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
		if !matchPhone.MatchString(u.Tel) {
			return ErrIncorrectPhone
		}

	}

	// Генерация хэша из пароля?

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
		return ErrEmptyEventInfo
	}
	if event.EventDate.UTC().Before(time.Now().UTC()) {
		return ErrIncorrectEventTime
	}
	now := time.Now().UTC()
	event.Created = &now
	event.AvailSeats = event.TotalSeats
	event.Status = model.EventStatusActual

	return nil
}
