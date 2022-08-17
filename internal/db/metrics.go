package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/metricsAndAlerting/internal/models"
	"log"
)

type Metrics struct {
	ID    string           `db:"id"`
	MType string           `db:"m_type"`
	Delta *sql.NullInt64   `db:"delta"`
	Value *sql.NullFloat64 `db:"value"`
}

func (db *DB) CreateTable(ctx context.Context) error {
	_, err := db.db.Query(`select * from metrics;`)
	if err != nil {
		_, err = db.db.ExecContext(ctx, createTable)
		return err
	}
	return err
}

func (db *DB) UpdateMetric(ctx context.Context, metrics models.Metrics) error {
	_, err := db.db.ExecContext(ctx, queryUpdateMetrics, metrics.ID, metrics.MType, metrics.Delta, metrics.Value)
	return err
}

func (db *DB) UpdateMetrics(metrics []models.Metrics) error {
	if db.db == nil {
		return errors.New("You haven`t opened the database connection")
	}
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(queryUpdateMetrics)
	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, m := range metrics {
		if _, err = stmt.Exec(m.ID, m.MType, m.Delta, m.Value); err != nil {
			if err = tx.Rollback(); err != nil {
				log.Fatalf("update drivers: unable to rollback: %v", err)
			}
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("update drivers: unable to commit: %v", err)
		return err
	}

	db.buffer = db.buffer[:0]
	return nil

}

func (db *DB) GetMetric(ctx context.Context, metricId string) (models.Metrics, error) {
	var dbMetric models.Metrics
	row := db.db.QueryRowContext(ctx, queryGetMetric, metricId)
	err := row.Scan(&dbMetric.ID, &dbMetric.MType, &dbMetric.Delta, &dbMetric.Value)
	if err != nil {
		return models.Metrics{}, err
	}
	return dbMetric, nil
}

func (db *DB) GetCounterMetric(ctx context.Context, metricId string) (int64, error) {
	var result int64
	row := db.db.QueryRowContext(ctx, queryGetCounterMetricValue, metricId)
	err := row.Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	return result, err
}

func (db *DB) GetGaugeMetric(ctx context.Context, metricId string) (float64, error) {
	var result float64
	row := db.db.QueryRowContext(ctx, queryGetGaugeMetricValue, metricId)
	err := row.Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	return result, err
}
