package db

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/metricsAndAlerting/internal/models"
)

type Metrics struct {
	ID    string           `db:"id"`
	MType string           `db:"m_type"`
	Delta *sql.NullInt64   `db:"delta"`
	Value *sql.NullFloat64 `db:"value"`
}

func (db *DB) CreateTable(ctx context.Context) error {
	row, err := db.db.Query(checkMetricDB)
	if err != nil {
		if err.Error() == `pq: relation "metrics" does not exist` {
			_, err := db.db.ExecContext(ctx, createTable)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if row.Err() != nil {
		return err
	}
	defer row.Close()
	return nil
}

func (db *DB) UpdateMetric(ctx context.Context, metrics models.Metrics) error {
	_, err := db.db.ExecContext(ctx, queryUpdateMetrics, metrics.ID, metrics.MType, metrics.Delta, metrics.Value)
	log.Println("Can't Update Metric")
	return err
}

func (db *DB) UpdateMetrics(metrics []models.Metrics) error {
	if db.db == nil {
		return errors.New("you haven`t opened the database connection")
	}
	tx, err := db.db.Begin()
	if err != nil {
		log.Println("Can't create tx", err)
		return err
	}

	stmt, err := tx.Prepare(queryUpdateMetrics)
	if err != nil {
		log.Println("Can't create stmt", err)
		return err
	}

	defer stmt.Close()

	for _, m := range metrics {
		if _, err = stmt.Exec(m.ID, m.MType, m.Delta, m.Value); err != nil {
			log.Println("Can't make Exec", err)
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

func (db *DB) GetMetric(ctx context.Context, metricID string) (models.Metrics, error) {
	var dbMetric models.Metrics
	row := db.db.QueryRowContext(ctx, queryGetMetric, metricID)
	err := row.Scan(&dbMetric.ID, &dbMetric.MType, &dbMetric.Delta, &dbMetric.Value)
	if err != nil {
		return models.Metrics{}, err
	}
	return dbMetric, nil
}

func (db *DB) GetCounterMetric(ctx context.Context, metricID string) (int64, error) {
	var result int64
	row := db.db.QueryRowContext(ctx, queryGetCounterMetricValue, metricID)
	err := row.Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	return result, err
}

func (db *DB) GetGaugeMetric(ctx context.Context, metricID string) (float64, error) {
	var result float64
	row := db.db.QueryRowContext(ctx, queryGetGaugeMetricValue, metricID)
	err := row.Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	return result, err
}
