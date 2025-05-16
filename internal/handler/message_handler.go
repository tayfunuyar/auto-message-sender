package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"auto-message-sender/internal/model/request"
	"auto-message-sender/internal/model/response"
	"auto-message-sender/internal/service"

	"github.com/labstack/echo/v4"
)

type MessageHandler interface {
	StartSending(c echo.Context) error
	StopSending(c echo.Context) error
	GetMessages(c echo.Context) error
	CreateMessage(c echo.Context) error
	RegisterRoutes(group *echo.Group)
}

type messageHandler struct {
	svc service.MessageService
}

func NewMessageHandler(svc service.MessageService) MessageHandler {
	return &messageHandler{svc: svc}
}

func (h *messageHandler) RegisterRoutes(group *echo.Group) {
	group.POST("/start", h.StartSending)
	group.POST("/stop", h.StopSending)
	group.GET("", h.GetMessages)
	group.POST("", h.CreateMessage)
}

// StartSending @Summary Start automatic message sending
// @Description Start the automatic message sending process
// @Tags messages
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /messages/start [post]
func (h *messageHandler) StartSending(c echo.Context) error {
	if err := h.svc.StartSending(context.Background()); err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response.SuccessResponse{
		Message: "Message sending started",
	})
}

// StopSending @Summary Stop automatic message sending
// @Description Stop the automatic message sending process
// @Tags messages
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /messages/stop [post]
func (h *messageHandler) StopSending(c echo.Context) error {
	if err := h.svc.StopSending(); err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response.SuccessResponse{
		Message: "Message sending stopped",
	})
}

// GetMessages @Summary Get messages
// @Description Get a list of messages with optional filtering
// @Tags messages
// @Accept json
// @Produce json
// @Param status query string false "Message status (pending/sent/failed)"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1) minimum(1)
// @Param page_size query int false "Page size" default(10) minimum(1) maximum(100)
// @Success 200 {object} response.MessageListResponse
// @Failure 400 {object} response.ValidationErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /messages [get]
func (h *messageHandler) GetMessages(c echo.Context) error {
	filter := new(request.MessageFilterRequest)
	if err := c.Bind(filter); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request format",
		})
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	if err := filter.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
	}

	messages, err := h.svc.GetMessages(filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: err.Error(),
		})
	}

	messageItems := make([]response.MessageItem, len(messages))
	for i, msg := range messages {
		messageItems[i] = response.MessageItem{
			ID:        msg.ID.String(),
			To:        msg.To,
			Content:   msg.Content,
			Status:    msg.Status,
			MessageID: msg.MessageID,
			SentAt:    msg.SentAt.Format(time.RFC3339),
		}
	}

	totalPages := 0
	if len(messageItems) > 0 {
		totalPages = (len(messageItems) + filter.PageSize - 1) / filter.PageSize
	}

	return c.JSON(http.StatusOK, response.MessageListResponse{
		Messages:   messageItems,
		Total:      len(messageItems),
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	})
}

// CreateMessage @Summary Create a new message
// @Description Create a new message to be sent
// @Tags messages
// @Accept json
// @Produce json
// @Param message body request.SendMessageRequest true "Message details"
// @Success 201 {object} response.MessageResponse
// @Failure 400 {object} response.ValidationErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /messages [post]
func (h *messageHandler) CreateMessage(c echo.Context) error {
	req := new(request.SendMessageRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request format",
		})
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: fmt.Sprintf("Validation error: %s", err.Error()),
		})
	}

	message, err := h.svc.CreateMessage(c.Request().Context(), req.To, req.Content)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, response.MessageResponse{
		Message:   "Message created successfully",
		MessageID: message.ID.String(),
	})
}
