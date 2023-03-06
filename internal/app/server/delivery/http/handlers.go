package http

import (
	serverApp "github.com/iddanilov/metricsAndAlerting/internal/app/server"
)

type handlers struct {
	uc serverApp.Usecase
}

func NewHandlers(serverUseCase serverApp.Usecase) handlers {
	return handlers{
		uc: serverUseCase,
	}
}
