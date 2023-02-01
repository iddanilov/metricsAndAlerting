package postgresql

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/lib/pq"

	"github.com/iddanilov/metricsAndAlerting/internal/models"
)

type DB struct {
	DB     *sql.DB
	buffer []models.Metrics
}

func NewDB(DNS string) (*DB, error) {
	db, err := sql.Open("postgres", DNS)
	if err != nil {
		return nil, err
	}
	log.Println("DB Opened")

	return &DB{
		DB:     db,
		buffer: make([]models.Metrics, 0, 1000),
	}, nil
}

func (db *DB) DBPing(ctx context.Context) error {
	return db.DB.PingContext(ctx)

}
