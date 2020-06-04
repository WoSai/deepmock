package service

import (
	"github.com/wosai/deepmock/types"
	"github.com/wosai/deepmock/types/domain"
)

var (
	Mock *mockService
)

type (
	mockService struct {
		repo   types.MockRepository
		domain domain.Mock
	}
)

func (m *mockService) HandleRequest() {}
