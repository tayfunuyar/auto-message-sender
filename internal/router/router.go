package router

import (
	"auto-message-sender/internal/handler"
	"auto-message-sender/pkg/health"

	"github.com/labstack/echo/v4"
)

type Config struct {
	MessageHandler handler.MessageHandler
	HealthConfig   health.Config
}

func SetupRoutes(e *echo.Echo, config Config) {
	health.RegisterRoutes(e, config.HealthConfig)

	v1 := e.Group("/api/v1")
	registerV1Routes(e, v1, config)
}

func registerV1Routes(e *echo.Echo, v1 *echo.Group, config Config) {
	messages := v1.Group("/messages")
	config.MessageHandler.RegisterRoutes(messages)
}
