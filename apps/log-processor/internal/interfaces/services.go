package interfaces

import (
	"log-processor/internal/domain"
)

type ConfigService interface {
	LoadConfig() error
	GetConfig() *domain.Config
}
