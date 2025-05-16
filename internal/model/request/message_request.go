package request

import (
	"auto-message-sender/internal/validator"
)

type SendMessageRequest struct {
	To      string `json:"to" validate:"required,e164"`
	Content string `json:"content" validate:"required,max=160"`
}

func (r *SendMessageRequest) Validate() error {
	if err := validator.ValidateMessageContent(r.Content); err != nil {
		return err
	}

	if err := validator.ValidatePhoneNumber(r.To); err != nil {
		return err
	}

	return nil
}

type MessageFilterRequest struct {
	Status    string `query:"status" validate:"omitempty,oneof=pending sent failed"`
	StartDate string `query:"start_date" validate:"omitempty,datetime=2006-01-02"`
	EndDate   string `query:"end_date" validate:"omitempty,datetime=2006-01-02"`
	Page      int    `query:"page" validate:"min=1"`
	PageSize  int    `query:"page_size" validate:"min=1,max=100"`
}

func (r *MessageFilterRequest) Validate() error {
	if err := validator.ValidateStatus(r.Status); err != nil {
		return err
	}
	if err := validator.ValidateDate(r.StartDate); err != nil {
		return err
	}
	if err := validator.ValidateDate(r.EndDate); err != nil {
		return err
	}
	if err := validator.ValidateDateRange(r.StartDate, r.EndDate); err != nil {
		return err
	}
	if err := validator.ValidatePageParams(r.Page, r.PageSize); err != nil {
		return err
	}

	return nil
}
