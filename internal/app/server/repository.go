package task

import (
	"context"

	"gitlab.corp.mail.ru/calendar/notesapi/internal/app/models"
)

//go:generate mockgen -package mock -destination repository/mock/note_mock.go . Repository

// Repository represent the Note's repository contract
type Repository interface {
	GetByID(ctx context.Context, key models.NotePrimaryKey) (models.Note, error)
}
