package validator

import (
	"fmt"
	"strings"
	"time"

	"auto-message-sender/internal/entity"

	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func ValidateMessageContent(content string) error {
	if len(content) == 0 {
		return fmt.Errorf("message content is required")
	}

	if len(content) > 160 {
		return fmt.Errorf("message content too long, maximum length is 160 characters")
	}

	return nil
}

func ValidatePhoneNumber(phone string) error {
	if len(phone) == 0 {
		return fmt.Errorf("phone number is required")
	}
	if !strings.HasPrefix(phone, "+") {
		return fmt.Errorf("phone number must begin with '+'")
	}

	for i := 1; i < len(phone); i++ {
		if phone[i] < '0' || phone[i] > '9' {
			return fmt.Errorf("phone number must contain only digits after '+' prefix")
		}
	}

	return nil
}

func ValidateStatus(status string) error {
	if status == "" {
		return nil
	}

	validStatuses := []string{entity.StatusPending, entity.StatusSent, entity.StatusFailed}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}

	return fmt.Errorf("status must be one of: %s", strings.Join(validStatuses, ", "))
}

func ValidateDate(date string) error {
	if date == "" {
		return nil
	}

	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("date must be in YYYY-MM-DD format")
	}

	return nil
}

func ValidateDateRange(startDate, endDate string) error {
	if startDate == "" || endDate == "" {
		return nil
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return fmt.Errorf("start date must be in YYYY-MM-DD format")
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return fmt.Errorf("end date must be in YYYY-MM-DD format")
	}

	if end.Before(start) {
		return fmt.Errorf("end date must be after start date")
	}

	return nil
}

func ValidatePageParams(page, pageSize int) error {
	if page < 1 {
		return fmt.Errorf("page number must be at least 1")
	}

	if pageSize < 1 || pageSize > 100 {
		return fmt.Errorf("page size must be between 1 and 100")
	}

	return nil
}
