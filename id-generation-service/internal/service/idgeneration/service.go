package idgeneration

import (
	"fmt"
	"time"

	"golang.org/x/exp/rand"
)

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) GenerateId(entity string) string {
	return fmt.Sprintf("%s-%d", entity, rand.New(rand.NewSource(uint64(time.Now().UnixNano()))).Int())
}
