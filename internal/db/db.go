package db

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/metricsAndAlerting/internal/models"
)

type DB struct {
	db     *sql.DB
	buffer []models.Metrics
}

func NewDB(DNS string) (*DB, error) {
	db, err := sql.Open("postgres", DNS)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &DB{
		db:     db,
		buffer: make([]models.Metrics, 0, 1000),
	}, nil
}

func (db *DB) DBPing(ctx context.Context) error {
	return db.db.PingContext(ctx)

}
