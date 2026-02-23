package health

import (
	"time"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Database  string    `json:"database"`
	Timestamp time.Time `json:"timestamp"`
}

func (s *Service) Check() (*HealthResponse, error) {
	sqlDB, err := s.db.DB()
	if err != nil {
		return &HealthResponse{
			Status:    "ERROR",
			Database:  "disconnected",
			Timestamp: time.Now(),
		}, err
	}

	if err := sqlDB.Ping(); err != nil {
		return &HealthResponse{
			Status:    "ERROR",
			Database:  "disconnected",
			Timestamp: time.Now(),
		}, err
	}

	return &HealthResponse{
		Status:    "OK",
		Database:  "connected",
		Timestamp: time.Now(),
	}, nil
}
