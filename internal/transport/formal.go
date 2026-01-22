// Package transport provides processing for incoming requests and preparing info for service-layer
package transport

import (
	"context"
	"strconv"

	"github.com/UnendingLoop/EventBooker/internal/model"
	"github.com/gin-gonic/gin"
)

type EBHandlers struct {
	svc HService
}

type HService interface {
	BookEvent(ctx context.Context, book *model.Book) error
	CancelBook(ctx context.Context, bid int, uid int) error
	ConfirmBook(ctx context.Context, bid int, uid int) error
	CreateEvent(ctx context.Context, event *model.Event) error
	CreateUser(ctx context.Context, user *model.User) (string, error)
	DeleteEvent(ctx context.Context, eid int, role string) error
	GetBooksListByUserID(ctx context.Context, uid int) ([]*model.Book, error)
	LoginUser(ctx context.Context, email string, password string) (string, *model.User, error)
	GetEventsList(ctx context.Context, role string) ([]*model.Event, error)
}

func NewEBHandlers(svc HService) *EBHandlers {
	return &EBHandlers{svc: svc}
}

// ---------------------------------------------------------------
type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	User userPublic `json:"user"`
}

type userPublic struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func convertUserAuthToResponse(user *model.User) *authResponse {
	return &authResponse{User: userPublic{ID: user.ID, Email: user.Email, Role: user.Role}}
}

// ----------------------------------------------------------
func stringFromCtx(ctx *gin.Context, key string) string {
	if v := ctx.Value(key); v != nil {
		return v.(string)
	}
	return ""
}

func intFromCtx(ctx *gin.Context, key string) int {
	if v := ctx.Value(key); v != nil {
		return v.(int)
	}
	return 0
}

func stringToInt(input string) int {
	output, err := strconv.Atoi(input)
	if err != nil {
		return -1
	}
	return output
}
