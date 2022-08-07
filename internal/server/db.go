package server

import (
	"database/sql"
	_ "github.com/lib/pq"
)

func New(DNS string) (*sql.DB, error) {
	db, err := sql.Open("postgres", DNS)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
