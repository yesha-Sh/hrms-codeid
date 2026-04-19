package db

import (
	"database/sql"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"hrms/backend/internal/config"
)

func OpenGORM(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open gorm database: %w", err)
	}
	return db, nil
}

func OpenSQL(cfg config.Config) (*sql.DB, error) {
	sqlDB, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open sql database: %w", err)
	}
	return sqlDB, nil
}
