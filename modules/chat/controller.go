package chat

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Gooowan/matchup/modules/core/logging"
	"github.com/Gooowan/matchup/modules/core/types"
	"github.com/Gooowan/matchup/modules/core/utils"
	"github.com/Gooowan/matchup/modules/users/auth"
)

type ChatController struct {
	svc *ChatService
}

func NewChatController(svc *ChatService) *ChatController {
	return &ChatController{svc: svc}
}

func (c *ChatController) ListChats(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	chats, err := c.svc.ListChats(ctx.Request.Context(), user.ID)
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to list chats", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to list chats"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: chats})
}

func (c *ChatController) GetMessages(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	chatID, err := utils.StringToUUID(ctx.Param("chatId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid chat ID"})
		return
	}

	// Parse cursor (defaults to far future for first page)
	cursor := time.Now().Add(24 * 365 * 100 * time.Hour)
	if cursorStr := ctx.Query("cursor"); cursorStr != "" {
		ms, err := strconv.ParseInt(cursorStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid cursor"})
			return
		}
		cursor = time.UnixMilli(ms)
	}

	limit := int32(50)
	if limitStr := ctx.Query("limit"); limitStr != "" {
		l, err := strconv.ParseInt(limitStr, 10, 32)
		if err == nil && l > 0 && l <= 100 {
			limit = int32(l)
		}
	}

	msgs, err := c.svc.GetMessages(ctx.Request.Context(), chatID, user.ID, cursor, limit)
	if err != nil {
		if err.Error() == "access denied" {
			ctx.JSON(http.StatusForbidden, types.Resp{Error: "Access denied"})
			return
		}
		logging.FromContext(ctx.Request.Context()).Error("failed to get messages", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to get messages"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: msgs})
}

func (c *ChatController) SendMessage(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	chatID, err := utils.StringToUUID(ctx.Param("chatId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid chat ID"})
		return
	}

	var req struct {
		Type    string `json:"type" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: err.Error()})
		return
	}

	msg, err := c.svc.SendMessage(ctx.Request.Context(), chatID, user.ID, req.Type, req.Content)
	if err != nil {
		if err.Error() == "access denied" {
			ctx.JSON(http.StatusForbidden, types.Resp{Error: "Access denied"})
			return
		}
		if err.Error() == "cannot send message to blocked user" {
			ctx.JSON(http.StatusForbidden, types.Resp{Error: err.Error()})
			return
		}
		logging.FromContext(ctx.Request.Context()).Error("failed to send message", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to send message"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: msg})
}

func (c *ChatController) RegisterRoutes(rg *gin.RouterGroup, userAuth gin.HandlerFunc, messageRL ...gin.HandlerFunc) {
	rg.Use(userAuth)
	rg.GET("", c.ListChats)
	rg.GET("/:chatId/messages", c.GetMessages)
	sendHandlers := []gin.HandlerFunc{c.SendMessage}
	if len(messageRL) > 0 {
		sendHandlers = append([]gin.HandlerFunc{messageRL[0]}, sendHandlers...)
	}
	rg.POST("/:chatId/messages", sendHandlers...)
}
