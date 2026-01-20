// Package model holds shared data-structures of the app
package model

import (
	"context"
	"strings"
	"time"
)

const (
	RoleAdmin = "admin"
	RoleUser  = "user"

	BookStatusCreated   = "created"
	BookStatusConfirmed = "confirmed"
	BookStatusCancelled = "cancelled"

	EventStatusActual    = "actual"
	EventStatusExpired   = "expired"
	EventStatusCancelled = "cancelled"
)

type (
	Event struct {
		ID         int        `json:"id,omitempty"`
		Title      string     `json:"title"`
		Descr      string     `json:"descr,omitempty"`
		Created    *time.Time `json:"created,omitempty"`
		Status     string     `json:"status,omitempty"`
		EventDate  CustomTime `json:"eventdate"`
		TotalSeats int        `json:"total"`            // общее кол-во мест у события для бронирования
		AvailSeats int        `json:"avail,omitempty"`  // доступное кол-во мест у события для бронирования
		BookWindow int        `json:"period,omitempty"` // период жизни неподтвержденной брони в секундах
	}
	Book struct {
		ID              int        `json:"id,omitempty"`
		EventID         int        `json:"eventid"`
		UserID          int        `json:"userid"`
		Status          string     `json:"status,omitempty"`
		Created         *time.Time `json:"created_at,omitempty"`
		ConfirmDeadline *time.Time `json:"confirm_deadline,omitempty"`
	}
	User struct {
		ID       int        `json:"id,omitempty"`
		Role     string     `json:"role,omitempty"`
		Created  *time.Time `json:"created,omitempty"`
		Name     string     `json:"name"`
		Surname  string     `json:"surname,omitempty"`
		Tel      string     `json:"tel,omitempty"`
		Email    string     `json:"email"`
		PassHash string     `json:"password,omitempty"`
	}

	CustomTime struct {
		time.Time
	}
)

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" || s == "" {
		ct.Time = time.Time{}
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	ct.Time = t
	return nil
}

func (ct *CustomTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + ct.Time.Format("2006-01-02") + `"`), nil
}

func RequestIDFromCtx(ctx context.Context) string {
	if v := ctx.Value("request_id"); v != nil {
		return v.(string)
	}
	return ""
}
