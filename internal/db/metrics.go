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
	row, err := db.DB.Query(checkMetricDB)
	if err != nil {
		if err.Error() == `pq: relation "metrics" does not exist` {
			_, err = db.DB.ExecContext(ctx, createTable)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if row != nil {
		if row.Err() != nil {
			return err
		}
		defer row.Close()
	}
	log.Println("DB Create")

	return nil
}

func (db *DB) UpdateMetric(ctx context.Context, metrics models.Metrics) error {
	if metrics.MType == "counter" {
		value, err := db.GetCounterMetric(ctx, metrics.ID)
		log.Println(sql.ErrNoRows)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
		} else {
			*metrics.Delta = *metrics.Delta + *value
		}
	}
	_, err := db.DB.ExecContext(ctx, queryUpdateMetrics, metrics.ID, metrics.MType, metrics.Delta, metrics.Value)
	if err != nil {
		log.Println("Can't Update Metric")
	}
	return err
}

func (db *DB) UpdateMetrics(metrics []models.Metrics) error {
	if db.DB == nil {
		return errors.New("you haven`t opened the database connection")
	}
	tx, err := db.DB.Begin()
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
		if m.MType == "counter" {
			value, err := db.GetCounterMetric(context.Background(), m.ID)
			log.Println(sql.ErrNoRows)
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					return err
				}
			} else {
				*m.Delta = *m.Delta + *value
			}
		}
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
	row := db.DB.QueryRowContext(ctx, queryGetMetric, metricID)
	err := row.Scan(&dbMetric.ID, &dbMetric.MType, &dbMetric.Delta, &dbMetric.Value)
	if err != nil {
		return models.Metrics{}, err
	}
	return dbMetric, nil
}

func (db *DB) GetMetricNames(ctx context.Context) ([]string, error) {
	var result []string
	rows, err := db.DB.QueryContext(ctx, queryGetMetricNames)
	if err != nil {
		return nil, err
	}
	// обязательно закрываем перед возвратом функции
	defer rows.Close()

	// пробегаем по всем записям
	for rows.Next() {
		var v string
		err = rows.Scan(&v)
		if err != nil {
			return nil, err
		}

		result = append(result, v)
	}

	// проверяем на ошибки
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *DB) GetCounterMetric(ctx context.Context, metricID string) (*int64, error) {
	var result int64
	row := db.DB.QueryRowContext(ctx, queryGetCounterMetricValue, metricID)
	err := row.Scan(&result)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &result, nil
}

func (db *DB) GetGaugeMetric(ctx context.Context, metricID string) (*float64, error) {
	var result float64
	row := db.DB.QueryRowContext(ctx, queryGetGaugeMetricValue, metricID)
	err := row.Scan(&result)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &result, nil
}
