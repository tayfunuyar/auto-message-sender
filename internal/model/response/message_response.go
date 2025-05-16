package response

type MessageResponse struct {
	Message   string `json:"message"`
	MessageID string `json:"messageId"`
}

type MessageListResponse struct {
	Messages   []MessageItem `json:"messages"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

type MessageItem struct {
	ID        string `json:"id"`
	To        string `json:"to"`
	Content   string `json:"content"`
	Status    string `json:"status"`
	MessageID string `json:"message_id,omitempty"`
	SentAt    string `json:"sent_at,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
