package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"auto-message-sender/internal/config"
	"auto-message-sender/pkg/logger"

	"github.com/go-redis/redis/v8"
)

type RedisService interface {
	CacheMessageID(ctx context.Context, messageID string, sentTime time.Time) error
	GetMessageSentTime(ctx context.Context, messageID string) (time.Time, error)
}

type redisService struct {
	client *redis.Client
}

func NewRedisService() RedisService {
	redisHost := config.AppSettings.Redis.Host
	redisPort := config.AppSettings.Redis.Port

	logger.WithFields(logrus.Fields{
		"host": redisHost,
		"port": redisPort,
	}).Debug("Initializing Redis client")

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: "",
		DB:       0,
	})

	return &redisService{
		client: client,
	}
}

func (s *redisService) CacheMessageID(ctx context.Context, messageID string, sentTime time.Time) error {
	key := fmt.Sprintf("message:%s", messageID)

	logger.WithFields(logrus.Fields{
		"key":      key,
		"sentTime": sentTime.Format(time.RFC3339),
	}).Debug("Caching message ID in Redis")

	err := s.client.Set(ctx, key, sentTime.Format(time.RFC3339), 0).Err()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"key":   key,
			"error": err.Error(),
		}).Error("Failed to cache message ID in Redis")
		return err
	}

	logger.WithField("key", key).Debug("Successfully cached message ID in Redis")
	return nil
}

func (s *redisService) GetMessageSentTime(ctx context.Context, messageID string) (time.Time, error) {
	key := fmt.Sprintf("message:%s", messageID)

	logger.WithField("key", key).Debug("Retrieving message sent time from Redis")

	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			logger.WithField("key", key).Debug("Message ID not found in Redis")
		} else {
			logger.WithFields(logrus.Fields{
				"key":   key,
				"error": err.Error(),
			}).Error("Failed to get message sent time from Redis")
		}
		return time.Time{}, err
	}

	sentTime, err := time.Parse(time.RFC3339, val)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"key":   key,
			"value": val,
			"error": err.Error(),
		}).Error("Failed to parse message sent time from Redis")
		return time.Time{}, err
	}

	logger.WithFields(logrus.Fields{
		"key":      key,
		"sentTime": sentTime.Format(time.RFC3339),
	}).Debug("Successfully retrieved message sent time from Redis")

	return sentTime, nil
}
