package transport

import (
	"log"
	"net/http"

	"github.com/UnendingLoop/EventBooker/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/ginext"
)

func (eh *EBHandlers) SimplePinger(ctx *ginext.Context) {
	rid := stringFromCtx(ctx, "request_id")
	ctx.JSON(200, gin.H{rid: "pong"})
}

func (eh *EBHandlers) SignUpUser(ctx *gin.Context) {
	var newUser model.User

	if err := ctx.ShouldBindJSON(&newUser); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user payload"})
		return
	}

	token, err := eh.svc.CreateUser(ctx.Request.Context(), &newUser)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	resp := convertUserAuthToResponse(token, &newUser)

	ctx.JSON(http.StatusCreated, resp)
}

func (eh *EBHandlers) CreateEvent(ctx *gin.Context) {
	// логируем админовые ивенты
	rid := stringFromCtx(ctx, "request_id")
	uid := stringFromCtx(ctx, "user_id")
	mail := stringFromCtx(ctx, "email")
	role := stringFromCtx(ctx, "role")

	log.Printf("rid=%q userID=%q userEmail=%q role=%q creating event", rid, uid, mail, role)

	// дальше обычная логика
	if role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var event model.Event
	if err := ctx.ShouldBindJSON(&event); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid event payload"})
		return
	}

	if err := eh.svc.CreateEvent(ctx.Request.Context(), &event); err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, event)
}

func (eh *EBHandlers) LoginUser(ctx *gin.Context) {
	var req authRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid auth payload"})
		return
	}

	token, user, err := eh.svc.LoginUser(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	resp := convertUserAuthToResponse(token, user)

	ctx.JSON(http.StatusOK, resp)
}

func (eh *EBHandlers) GetEvents(ctx *gin.Context) {
	role := stringFromCtx(ctx, "role")
	res, err := eh.svc.GetEventsList(ctx.Request.Context(), role)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func (eh *EBHandlers) DeleteEvent(ctx *gin.Context) {
	// логируем админовые ивенты
	rid := stringFromCtx(ctx, "request_id")
	uid := stringFromCtx(ctx, "user_id")
	mail := stringFromCtx(ctx, "email")
	role := stringFromCtx(ctx, "role")

	log.Printf("rid=%q userID=%q userEmail=%q role=%q deleting event", rid, uid, mail, role)

	// обычный флоу
	rawID, ok := ctx.Params.Get("id")
	if !ok {
		ctx.JSON(400, gin.H{"error": "empty event id"})
		return
	}

	eventID := stringToInt(rawID)
	err := eh.svc.DeleteEvent(ctx.Request.Context(), eventID)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

func (eh *EBHandlers) BookEvent(ctx *gin.Context) {
	var book model.Book
	if err := ctx.ShouldBindJSON(&book); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking payload"})
		return
	}
	uid := stringFromCtx(ctx, "user_id")

	err := eh.svc.BookEvent(ctx.Request.Context(), &book, stringToInt(uid))
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, book)
}

func (eh *EBHandlers) ConfirmBook(ctx *gin.Context) {
	uid := stringFromCtx(ctx, "user_id")
	bid, ok := ctx.Params.Get("id")
	if !ok {
		ctx.JSON(400, gin.H{"error": "empty book id"})
		return
	}

	if err := eh.svc.ConfirmBook(ctx.Request.Context(), stringToInt(bid), stringToInt(uid)); err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

func (eh *EBHandlers) GetUserBooks(ctx *gin.Context) {
	uid := stringFromCtx(ctx, "user_id")

	res, err := eh.svc.GetBooksListByUserID(ctx.Request.Context(), stringToInt(uid))
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func (eh *EBHandlers) CancelBook(ctx *gin.Context) {
	uid := stringFromCtx(ctx, "user_id")
	bid, ok := ctx.Params.Get("id")
	if !ok {
		ctx.JSON(400, gin.H{"error": "empty book id"})
		return
	}
	if err := eh.svc.CancelBook(ctx.Request.Context(), stringToInt(bid), stringToInt(uid)); err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}
