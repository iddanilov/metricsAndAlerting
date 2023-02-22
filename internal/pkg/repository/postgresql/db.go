package postgresql

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"

	"github.com/iddanilov/metricsAndAlerting/internal/models"
)

type DB struct {
	DB     *sql.DB
	Buffer []models.Metrics
}

func NewDB(DNS string) (*DB, error) {
	db, err := sql.Open("postgres", DNS)
	if err != nil {
		return nil, err
	}
	log.Println("DB Opened")

	return &DB{
		DB:     db,
		Buffer: make([]models.Metrics, 0, 1000),
	}, nil
}
