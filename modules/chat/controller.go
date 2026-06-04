package chat

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Gooowan/matchup/modules/core/logging"
	"github.com/Gooowan/matchup/modules/core/types"
	corehttp "github.com/Gooowan/matchup/modules/core/http"
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

// CreateOrGetChat creates a new direct-message chat between the caller and target user.
// Requires a mutual match (both users swiped LIKE on each other).
func (c *ChatController) CreateOrGetChat(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if !corehttp.BindJSON(ctx, &req) {
		return
	}

	targetID, err := utils.StringToUUID(req.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid user_id"})
		return
	}

	chatID, err := c.svc.CreateChatForUsers(ctx.Request.Context(), user.ID, targetID)
	if err != nil {
		if errors.Is(err, ErrNoMutualMatch) {
			ctx.JSON(http.StatusForbidden, types.Resp{Error: "No mutual match with this user"})
			return
		}
		logging.FromContext(ctx.Request.Context()).Error("failed to create chat", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to create chat"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: gin.H{"chat_id": utils.UUIDToString(chatID)}})
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

	// cursor: fetch messages BEFORE this Unix-ms timestamp (pagination, default = far future).
	cursor := time.Now().Add(100 * 365 * 24 * time.Hour)
	if cursorStr := ctx.Query("cursor"); cursorStr != "" {
		ms, err := strconv.ParseInt(cursorStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid cursor"})
			return
		}
		cursor = time.UnixMilli(ms)
	}

	// after: fetch messages AFTER this Unix-ms timestamp (polling new messages).
	var afterTime time.Time
	if afterStr := ctx.Query("after"); afterStr != "" {
		ms, err := strconv.ParseInt(afterStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid after"})
			return
		}
		afterTime = time.UnixMilli(ms)
	}

	limit := int32(50)
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil && l > 0 && l <= 100 {
			limit = int32(l)
		}
	}

	msgs, err := c.svc.GetMessages(ctx.Request.Context(), chatID, user.ID, cursor, afterTime, limit)
	if err != nil {
		if errors.Is(err, ErrChatNotFound) {
			ctx.JSON(http.StatusNotFound, types.Resp{Error: "Chat not found"})
			return
		}
		if errors.Is(err, ErrAccessDenied) {
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
	if !corehttp.BindJSON(ctx, &req) {
		return
	}

	msg, err := c.svc.SendMessage(ctx.Request.Context(), chatID, user.ID, req.Type, req.Content)
	if err != nil {
		switch {
		case errors.Is(err, ErrChatNotFound):
			ctx.JSON(http.StatusNotFound, types.Resp{Error: "Chat not found"})
		case errors.Is(err, ErrAccessDenied):
			ctx.JSON(http.StatusForbidden, types.Resp{Error: "Access denied"})
		case errors.Is(err, ErrUserBlocked):
			ctx.JSON(http.StatusForbidden, types.Resp{Error: "Cannot send message to blocked user"})
		case errors.Is(err, ErrContentBlocked):
			ctx.JSON(http.StatusUnprocessableEntity, types.Resp{Error: err.Error()})
		default:
			logging.FromContext(ctx.Request.Context()).Error("failed to send message", "error", err)
			ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to send message"})
		}
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: msg})
}

// GetChatMeta returns the peer's profile data for the chat thread header.
func (c *ChatController) GetChatMeta(ctx *gin.Context) {
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

	peer, err := c.svc.GetChatMeta(ctx.Request.Context(), chatID, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, ErrChatNotFound):
			ctx.JSON(http.StatusNotFound, types.Resp{Error: "Chat not found"})
		case errors.Is(err, ErrAccessDenied):
			ctx.JSON(http.StatusForbidden, types.Resp{Error: "Access denied"})
		default:
			logging.FromContext(ctx.Request.Context()).Error("failed to get chat meta", "error", err)
			ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to get chat metadata"})
		}
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: peer})
}

// MarkChatRead marks all messages in a chat as read for the calling user.
func (c *ChatController) MarkChatRead(ctx *gin.Context) {
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

	if err := c.svc.MarkChatRead(ctx.Request.Context(), chatID, user.ID); err != nil {
		if errors.Is(err, ErrChatNotFound) {
			ctx.JSON(http.StatusNotFound, types.Resp{Error: "Chat not found"})
			return
		}
		if errors.Is(err, ErrAccessDenied) {
			ctx.JSON(http.StatusForbidden, types.Resp{Error: "Access denied"})
			return
		}
		logging.FromContext(ctx.Request.Context()).Error("failed to mark chat read", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to mark chat as read"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: gin.H{"ok": true}})
}

// ReportMessage lets a chat participant report a specific message for admin review.
func (c *ChatController) ReportMessage(ctx *gin.Context) {
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
	messageID, err := utils.StringToUUID(ctx.Param("messageId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid message ID"})
		return
	}

	var req struct {
		Category string `json:"category" binding:"required"`
		Comment  string `json:"comment"`
	}
	if !corehttp.BindJSON(ctx, &req) {
		return
	}

	report, err := c.svc.ReportMessage(ctx.Request.Context(), chatID, messageID, user.ID, req.Category, req.Comment)
	if err != nil {
		switch {
		case errors.Is(err, ErrChatNotFound):
			ctx.JSON(http.StatusNotFound, types.Resp{Error: "Chat or message not found"})
		case errors.Is(err, ErrAccessDenied):
			ctx.JSON(http.StatusForbidden, types.Resp{Error: "Access denied"})
		default:
			logging.FromContext(ctx.Request.Context()).Error("failed to report message", "error", err)
			ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to report message"})
		}
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: gin.H{"id": utils.UUIDToString(report.ID)}})
}

// AdminListMessageReports returns open message reports (admin only).
func (c *ChatController) AdminListMessageReports(ctx *gin.Context) {
	status := ctx.DefaultQuery("status", "open")
	limit := int32(50)
	offset := int32(0)
	if l, err := strconv.ParseInt(ctx.DefaultQuery("limit", "50"), 10, 32); err == nil && l > 0 && l <= 200 {
		limit = int32(l)
	}
	if o, err := strconv.ParseInt(ctx.DefaultQuery("offset", "0"), 10, 32); err == nil && o >= 0 {
		offset = int32(o)
	}

	reports, err := c.svc.AdminListMessageReports(ctx.Request.Context(), status, limit, offset)
	if err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to list message reports", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to list message reports"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: reports})
}

// AdminHideMessage soft-deletes a message (sets moderation_status=hidden). Admin only.
func (c *ChatController) AdminHideMessage(ctx *gin.Context) {
	admin, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	messageID, err := utils.StringToUUID(ctx.Param("messageId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid message ID"})
		return
	}

	if err := c.svc.AdminHideMessage(ctx.Request.Context(), messageID); err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to hide message", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to hide message"})
		return
	}

	logging.FromContext(ctx.Request.Context()).Info("admin_audit: hide_message",
		"event", "admin_audit",
		"action", "hide_message",
		"target_type", "message",
		"target_id", ctx.Param("messageId"),
		"admin_id", admin.ID.String(),
	)

	ctx.JSON(http.StatusOK, types.Resp{Data: gin.H{"ok": true}})
}

// AdminResolveMessageReport marks a message report as resolved or dismissed. Admin only.
func (c *ChatController) AdminResolveMessageReport(ctx *gin.Context) {
	user, ok := auth.GetUserFromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Resp{Error: "Unauthorized"})
		return
	}

	reportID, err := utils.StringToUUID(ctx.Param("reportId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.Resp{Error: "Invalid report ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=resolved dismissed"`
	}
	if !corehttp.BindJSON(ctx, &req) {
		return
	}

	if err := c.svc.AdminResolveMessageReport(ctx.Request.Context(), reportID, user.ID, req.Status); err != nil {
		logging.FromContext(ctx.Request.Context()).Error("failed to resolve message report", "error", err)
		ctx.JSON(http.StatusInternalServerError, types.Resp{Error: "Failed to resolve report"})
		return
	}

	ctx.JSON(http.StatusOK, types.Resp{Data: gin.H{"ok": true}})
}

func (c *ChatController) RegisterRoutes(rg *gin.RouterGroup, userAuth gin.HandlerFunc, messageRL ...gin.HandlerFunc) {
	rg.Use(userAuth)
	rg.GET("", c.ListChats)
	rg.POST("", c.CreateOrGetChat)
	rg.GET("/:chatId/meta", c.GetChatMeta)
	rg.GET("/:chatId/messages", c.GetMessages)
	rg.POST("/:chatId/read", c.MarkChatRead)
	rg.POST("/:chatId/messages/:messageId/report", c.ReportMessage)

	sendHandlers := []gin.HandlerFunc{c.SendMessage}
	if len(messageRL) > 0 {
		sendHandlers = append([]gin.HandlerFunc{messageRL[0]}, sendHandlers...)
	}
	rg.POST("/:chatId/messages", sendHandlers...)
}

// RegisterAdminRoutes registers admin-only moderation endpoints for messages.
// Must be called with an admin auth middleware applied.
func (c *ChatController) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.GET("/message-reports", c.AdminListMessageReports)
	rg.PATCH("/message-reports/:reportId", c.AdminResolveMessageReport)
	rg.DELETE("/messages/:messageId", c.AdminHideMessage)
}
