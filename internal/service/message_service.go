package service

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"auto-message-sender/internal/client"
	"auto-message-sender/internal/entity"
	"auto-message-sender/internal/model/request"
	"auto-message-sender/internal/repository"
	"auto-message-sender/pkg/logger"
)

type MessageService interface {
	StartSending(ctx context.Context) error
	StopSending() error
	GetMessages(filter *request.MessageFilterRequest) ([]entity.Message, error)
	CreateMessage(ctx context.Context, to, content string) (*entity.Message, error)
}

type messageService struct {
	repo          repository.MessageRepository
	webhookClient client.WebhookClient
	redisSvc      RedisService
	stopChan      chan struct{}
	wg            sync.WaitGroup
	isRunning     bool
	runningMutex  sync.Mutex
}

func NewMessageService(repo repository.MessageRepository, webhookClient client.WebhookClient, redisSvc RedisService) MessageService {
	return &messageService{
		repo:          repo,
		webhookClient: webhookClient,
		redisSvc:      redisSvc,
		stopChan:      make(chan struct{}),
		isRunning:     false,
	}
}

func (s *messageService) StartSending(ctx context.Context) error {
	s.runningMutex.Lock()
	defer s.runningMutex.Unlock()

	if s.isRunning {
		logger.Info("Message sending service is already running")
		return nil
	}

	s.stopChan = make(chan struct{})
	s.isRunning = true
	s.wg.Add(1)

	bgCtx := context.Background()

	logger.Info("Starting automatic message sending service")
	go s.processPendingMessages(bgCtx)
	return nil
}

func (s *messageService) StopSending() error {
	s.runningMutex.Lock()
	defer s.runningMutex.Unlock()

	if !s.isRunning {
		logger.Info("Message sending service is already stopped")
		return nil
	}

	logger.Info("Stopping automatic message sending service")
	close(s.stopChan)
	s.wg.Wait()
	s.isRunning = false
	logger.Info("Message sending service stopped successfully")
	return nil
}

func (s *messageService) GetMessages(filter *request.MessageFilterRequest) ([]entity.Message, error) {
	logger.WithFields(logrus.Fields{
		"status":    filter.Status,
		"startDate": filter.StartDate,
		"endDate":   filter.EndDate,
		"page":      filter.Page,
		"pageSize":  filter.PageSize,
	}).Debug("Retrieving filtered messages")

	messages, err := s.repo.GetMessages(filter)
	if err != nil {
		logger.WithError(err).Error("Failed to retrieve sent messages")
		return nil, err
	}

	logger.WithField("count", len(messages)).Debug("Retrieved sent messages")
	return messages, nil
}

func (s *messageService) CreateMessage(ctx context.Context, to, content string) (*entity.Message, error) {
	messageID := uuid.New()
	logger.WithFields(logrus.Fields{
		"messageID": messageID.String(),
		"to":        to,
		"length":    len(content),
	}).Info("Creating new message")

	message := &entity.Message{
		ID:      messageID,
		To:      to,
		Content: content,
		Status:  entity.StatusPending,
	}

	err := s.repo.Create(message)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"messageID": messageID.String(),
			"error":     err.Error(),
		}).Error("Failed to create message")
		return nil, err
	}

	logger.WithField("messageID", messageID.String()).Info("Message created successfully")
	return message, nil
}

func (s *messageService) processPendingMessages(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	logger.Info("Message processing routine started")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Message processing stopped due to context cancellation")
			return
		case <-s.stopChan:
			logger.Info("Message processing stopped via stop channel")
			return
		case t := <-ticker.C:
			logger.WithField("time", t.Format(time.RFC3339)).Debug("Processing pending messages")

			messages, err := s.repo.GetUnsentMessages(2)
			if err != nil {
				logger.WithError(err).Error("Failed to get unsent messages")
				continue
			}

			logger.WithField("count", len(messages)).Info("Retrieved unsent messages for processing")

			for _, msg := range messages {
				logger.WithFields(logrus.Fields{
					"messageID": msg.ID.String(),
					"to":        msg.To,
				}).Info("Sending message")

				messageID, err := s.webhookClient.SendMessage(msg)
				if err != nil {
					logger.WithFields(logrus.Fields{
						"messageID": msg.ID.String(),
						"error":     err.Error(),
					}).Error("Failed to send message via webhook")
					continue
				}

				sentTime := time.Now()
				logger.WithFields(logrus.Fields{
					"messageID":    msg.ID.String(),
					"webhookMsgID": messageID,
					"sentTime":     sentTime.Format(time.RFC3339),
				}).Info("Message sent successfully via webhook")

				err = s.repo.UpdateMessageID(msg.ID, messageID)
				if err != nil {
					logger.WithFields(logrus.Fields{
						"messageID":    msg.ID.String(),
						"webhookMsgID": messageID,
						"error":        err.Error(),
					}).Error("Failed to update message ID")
					continue
				}

				err = s.repo.UpdateStatus(messageID, entity.StatusSent, sentTime)
				if err != nil {
					logger.WithFields(logrus.Fields{
						"messageID":    msg.ID.String(),
						"webhookMsgID": messageID,
						"error":        err.Error(),
					}).Error("Failed to update message status")
					continue
				}

				err = s.redisSvc.CacheMessageID(ctx, messageID, sentTime)
				if err != nil {
					logger.WithFields(logrus.Fields{
						"messageID":    msg.ID.String(),
						"webhookMsgID": messageID,
						"error":        err.Error(),
					}).Warn("Failed to cache message ID in Redis")
				} else {
					logger.WithFields(logrus.Fields{
						"messageID":    msg.ID.String(),
						"webhookMsgID": messageID,
					}).Debug("Message ID cached in Redis")
				}

				logger.WithFields(logrus.Fields{
					"messageID":    msg.ID.String(),
					"webhookMsgID": messageID,
				}).Info("Message processing completed successfully")
			}
		}
	}
}
