package config

import (
	"github.com/kelseyhightower/envconfig"

	"log-processor/internal/domain"
)

type Service struct {
	config *domain.Config
}

func NewConfigService() *Service {
	return &Service{}
}

func (s *Service) GetConfig() *domain.Config {
	return s.config
}

func (s *Service) LoadConfig() error {
	if s.config != nil {
		return nil
	}

	var config domain.Config
	err := envconfig.Process("LOG_PROCESSOR", &config)
	if err != nil {
		return err
	}

	s.config = &config

	return nil
}
