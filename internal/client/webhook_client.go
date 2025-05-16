package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"auto-message-sender/internal/config"
	"auto-message-sender/internal/entity"
	"auto-message-sender/internal/model/request"
	"auto-message-sender/internal/model/response"
	"auto-message-sender/pkg/logger"
)

type WebhookClient interface {
	SendMessage(message entity.Message) (string, error)
}

type webhookClient struct {
	client *http.Client
}

func NewWebhookClient() WebhookClient {
	return &webhookClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *webhookClient) SendMessage(message entity.Message) (string, error) {
	webhookURL := config.AppSettings.Webhook.URL

	logger.WithFields(logrus.Fields{
		"messageID": message.ID.String(),
		"to":        message.To,
		"url":       webhookURL,
	}).Debug("Preparing webhook request")

	if webhookURL == "" {
		logger.Error("Webhook URL is not configured")
		return "", errors.New("webhook URL is not configured")
	}

	payload := request.WebhookRequest{
		To:      message.To,
		Content: message.Content,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"messageID": message.ID.String(),
			"error":     err.Error(),
		}).Error("Failed to marshal webhook request")
		return "", fmt.Errorf("failed to marshal webhook request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewReader(payloadBytes))
	if err != nil {
		logger.WithFields(logrus.Fields{
			"messageID": message.ID.String(),
			"url":       webhookURL,
			"error":     err.Error(),
		}).Error("Failed to create webhook request")
		return "", fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	startTime := time.Now()
	logger.WithFields(logrus.Fields{
		"messageID": message.ID.String(),
		"url":       webhookURL,
	}).Debug("Sending webhook request")

	resp, err := c.client.Do(req)
	requestDuration := time.Since(startTime)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"messageID": message.ID.String(),
			"url":       webhookURL,
			"duration":  requestDuration.String(),
			"error":     err.Error(),
		}).Error("Failed to send webhook request")
		return "", fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer resp.Body.Close()

	logger.WithFields(logrus.Fields{
		"messageID":  message.ID.String(),
		"statusCode": resp.StatusCode,
		"duration":   requestDuration.String(),
	}).Debug("Webhook response received")

	if resp.StatusCode != http.StatusOK {
		logger.WithFields(logrus.Fields{
			"messageID":  message.ID.String(),
			"statusCode": resp.StatusCode,
			"duration":   requestDuration.String(),
		}).Error("Webhook request failed with non-202 status code")
		return "", fmt.Errorf("webhook request failed with status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"messageID": message.ID.String(),
			"error":     err.Error(),
		}).Error("Failed to read webhook response body")
		return "", fmt.Errorf("failed to read webhook response body: %w", err)
	}

	var webhookResponse response.WebhookResponse
	if err := json.Unmarshal(bodyBytes, &webhookResponse); err != nil {
		logger.WithFields(logrus.Fields{
			"messageID": message.ID.String(),
			"error":     err.Error(),
		}).Error("Failed to decode webhook response")
		return "", fmt.Errorf("failed to decode webhook response: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"messageID":    message.ID.String(),
		"webhookMsgID": webhookResponse.MessageID,
		"duration":     requestDuration.String(),
	}).Info("Webhook request completed successfully")

	return webhookResponse.MessageID, nil
}
