package idgeneration

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/exp/rand"
)

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) GenerateId(ctx context.Context, entity string) string {
	return fmt.Sprintf("%s-%d", entity, rand.New(rand.NewSource(uint64(time.Now().UnixNano()))).Int())
}
