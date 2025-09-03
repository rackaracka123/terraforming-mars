package service

import "terraforming-mars-backend/internal/repository"

// HelloService handles hello business logic
type HelloService struct {
	helloRepo *repository.HelloRepository
}

// NewHelloService creates a new hello service
func NewHelloService(helloRepo *repository.HelloRepository) *HelloService {
	return &HelloService{
		helloRepo: helloRepo,
	}
}

// GetMessage returns a hello message
func (s *HelloService) GetMessage() string {
	return s.helloRepo.GetMessage()
}