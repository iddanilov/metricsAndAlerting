package task

import (
	"context"
	"gitlab.corp.mail.ru/calendar/notesapi/internal/app/models"
)

//go:generate mockgen -package mock -destination usecase/mock/task_mock.go . Usecase

// Usecase represent the tasks's usecases
type Usecase interface {
	GetByID(ctx context.Context, key models.NotePrimaryKey) (models.Note, error)
}
