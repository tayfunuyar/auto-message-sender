package database

import (
	"fmt"

	"auto-message-sender/internal/config"
	"auto-message-sender/internal/entity"
	"auto-message-sender/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Setup() (*gorm.DB, error) {
	initDSN := fmt.Sprintf("host=%s user=%s password=%s port=%s sslmode=disable",
		config.AppSettings.Database.Host,
		config.AppSettings.Database.User,
		config.AppSettings.Database.Password,
		config.AppSettings.Database.Port,
	)

	logger.Debugf("Connecting to postgres with DSN: host=%s port=%s user=%s",
		config.AppSettings.Database.Host,
		config.AppSettings.Database.Port,
		config.AppSettings.Database.User)

	initDB, err := gorm.Open(postgres.Open(initDSN), &gorm.Config{})
	if err != nil {
		logger.WithError(err).Error("Failed to connect to postgres")
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	dbname := config.AppSettings.Database.Name
	var exists int64
	sqlDB, err := initDB.DB()
	if err != nil {
		logger.WithError(err).Error("Failed to get sql.DB")
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	logger.Debugf("Checking if database %s exists", dbname)
	err = sqlDB.QueryRow("SELECT 1 FROM pg_database WHERE datname = $1", dbname).Scan(&exists)
	if err != nil || exists != 1 {
		logger.Infof("Database %s does not exist, creating...", dbname)
		_, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
		if err != nil {
			logger.WithError(err).Errorf("Failed to create database %s", dbname)
			return nil, fmt.Errorf("failed to create database %s: %w", dbname, err)
		}
		logger.Infof("Database %s created successfully", dbname)
	} else {
		logger.Infof("Database %s already exists", dbname)
	}

	sqlDB.Close()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		config.AppSettings.Database.Host,
		config.AppSettings.Database.User,
		config.AppSettings.Database.Password,
		config.AppSettings.Database.Name,
		config.AppSettings.Database.Port,
	)

	logger.Debugf("Connecting to database %s", dbname)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.WithError(err).Errorf("Failed to connect to database %s", dbname)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	logger.Infof("Connected to database %s successfully", dbname)

	logger.Debug("Running database migrations")
	if err := runMigrations(db); err != nil {
		logger.WithError(err).Error("Failed to run database migrations")
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}
	logger.Info("Database migrations completed successfully")

	logger.Debug("Seeding test data")
	if err := seedTestData(db); err != nil {
		logger.WithError(err).Warn("Failed to seed test data")
	} else {
		logger.Info("Test data seeded successfully")
	}

	return db, nil
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(&entity.Message{})
}

func seedTestData(db *gorm.DB) error {
	return SeedMessages(db)
}
