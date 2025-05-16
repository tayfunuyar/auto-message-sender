package request

type WebhookRequest struct {
	To      string `json:"to"`
	Content string `json:"content"`
}
