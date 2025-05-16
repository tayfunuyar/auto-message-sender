package health

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type Status string

const (
	StatusUp       Status = "UP"
	StatusDown     Status = "DOWN"
	StatusDisabled Status = "DISABLED"
)

type Response struct {
	Status    Status            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
}

type Config struct {
	Version string
	DB      *gorm.DB
}

func RegisterRoutes(e *echo.Echo, config Config) {
	e.GET("/health", Handler(config))
	e.GET("/health/ready", Handler(config))
	e.GET("/health/live", Handler(config))
}

func Handler(config Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		health := Response{
			Status:    StatusUp,
			Timestamp: time.Now().Format(time.RFC3339),
			Version:   config.Version,
			Services: map[string]string{
				"api": string(StatusUp),
			},
		}

		if config.DB != nil {
			sqlDB, err := config.DB.DB()
			if err != nil {
				health.Status = StatusDown
				health.Services["database"] = string(StatusDown)
				return c.JSON(http.StatusServiceUnavailable, health)
			}

			if err := sqlDB.Ping(); err != nil {
				health.Status = StatusDown
				health.Services["database"] = string(StatusDown)
				return c.JSON(http.StatusServiceUnavailable, health)
			}

			health.Services["database"] = string(StatusUp)
		} else {
			health.Services["database"] = string(StatusDisabled)
		}

		statusCode := http.StatusOK
		if health.Status == StatusDown {
			statusCode = http.StatusServiceUnavailable
		}

		return c.JSON(statusCode, health)
	}
}
