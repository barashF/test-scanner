package gateway

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ServiceConference struct{}

func NewServiceConference() *ServiceConference { return &ServiceConference{} }

func (s *ServiceConference) CreateConference(ctx context.Context, userID uuid.UUID, startTime time.Time) (string, error) {
	return "link", nil
}
