package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "auto-message-sender/docs"
	"auto-message-sender/internal/client"
	"auto-message-sender/internal/config"
	"auto-message-sender/internal/database"
	"auto-message-sender/internal/handler"
	"auto-message-sender/internal/repository"
	"auto-message-sender/internal/router"
	"auto-message-sender/internal/service"
	"auto-message-sender/internal/validator"
	"auto-message-sender/pkg/health"
	"auto-message-sender/pkg/logger"
)

const appVersion = "1.0.0"

// @title Message API
// @version 1.0
// @description This is a message sending service API
// @host localhost:8080
// @BasePath /api/v1
var appContext context.Context
var appCancel context.CancelFunc

func main() {
	logger.Init(logger.InfoLevel)

	if err := config.LoadSettings(); err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.Setup()
	if err != nil {
		logger.Fatalf("Failed to setup database: %v", err)
	}

	messageRepo := repository.NewMessageRepository(db)
	webhookClient := client.NewWebhookClient()
	redisSvc := service.NewRedisService()
	messageSvc := service.NewMessageService(messageRepo, webhookClient, redisSvc)

	messageHandler := handler.NewMessageHandler(messageSvc)

	e := echo.New()

	e.Validator = validator.NewCustomValidator()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	routerConfig := router.Config{
		MessageHandler: messageHandler,
		HealthConfig: health.Config{
			Version: appVersion,
			DB:      db,
		},
	}
	router.SetupRoutes(e, routerConfig)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	appContext, appCancel = context.WithCancel(context.Background())
	if err := messageSvc.StartSending(appContext); err != nil {
		logger.Errorf("Failed to start message sending: %v", err)
	} else {
		logger.Info("Automatic message sending started")
	}
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		logger.Info("Shutting down server...")

		if err := messageSvc.StopSending(); err != nil {
			logger.Errorf("Error stopping message sender: %v", err)
		}

		appCancel()
		if err := e.Shutdown(appContext); err != nil {
			logger.Fatalf("Error shutting down server: %v", err)
		}
	}()

	port := config.AppSettings.Server.Port
	logger.Infof("Server starting on port %s", port)
	if err := e.Start(fmt.Sprintf(":%s", port)); err != nil && err != http.ErrServerClosed {
		log.Fatal("Error starting server:", err)
	}
}
