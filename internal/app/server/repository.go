package task

import (
	"context"
)

// Repository represent the metric and server repository contract
type Repository interface {
	CreateTable(ctx context.Context) error
}
