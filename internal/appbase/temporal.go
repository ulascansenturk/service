package appbase

import (
	"context"
	"go.temporal.io/sdk/client"
)

type TemporalService struct {
	Client client.Client
}

func NewTemporalService(options client.Options) (*TemporalService, error) {
	c, err := client.Dial(options)
	if err != nil {
		return nil, err
	}

	return &TemporalService{Client: c}, nil
}

func (s *TemporalService) HealthCheck() error {
	_, err := s.Client.CheckHealth(context.Background(), nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *TemporalService) Shutdown() error {
	s.Client.Close()

	return nil
}
