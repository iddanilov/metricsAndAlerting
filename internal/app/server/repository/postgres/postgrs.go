package postgresql

import (
	"context"
	"errors"

	server "github.com/iddanilov/metricsAndAlerting/internal/app/server"
	"github.com/iddanilov/metricsAndAlerting/internal/models"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/logger"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/repository/postgresql"
)

type serverRepository struct {
	db     postgresql.DB
	logger logger.Logger
}

func NewServerRepository(db postgresql.DB, logger logger.Logger) server.Repository {
	return &serverRepository{
		logger: logger,
		db:     db,
	}
}

func (r *serverRepository) Ping() error {
	return r.db.DB.Ping()
}

func (r *serverRepository) CreateTable(ctx context.Context) error {
	row, err := r.db.DB.Query(checkMetricDB)
	if err != nil {
		if err.Error() == `pq: relation "metrics" does not exist` {
			_, err = r.db.DB.ExecContext(ctx, createTable)
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

	}
	r.logger.Info("DB Create")

	return nil
}

func (r *serverRepository) UpdateMetric(ctx context.Context, metrics models.Metrics) error {
	_, err := r.db.DB.ExecContext(ctx, queryUpdateMetrics, metrics.ID, metrics.MType, metrics.Delta, metrics.Value)
	if err != nil {
		r.logger.Error("Can't Update Metric")
	}
	return err
}

func (r *serverRepository) UpdateMetrics(metrics []models.Metrics) error {
	if r.db.DB == nil {
		return errors.New("you haven`t opened the database connection")
	}
	tx, err := r.db.DB.Begin()
	if err != nil {
		r.logger.Error("can't create tx", err)
		return err
	}

	stmt, err := tx.Prepare(queryUpdateMetrics)
	if err != nil {
		r.logger.Error("can't create stmt", err)
		return err
	}

	for _, m := range metrics {
		if _, err = stmt.Exec(m.ID, m.MType, m.Delta, m.Value); err != nil {
			r.logger.Error("can't make Exec", err)
			if err = tx.Rollback(); err != nil {
				r.logger.Fatalf("update drivers: unable to rollback: %v", err)
			}
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		r.logger.Fatalf("update drivers: unable to commit: %v", err)
		return err
	}

	r.db.Buffer = r.db.Buffer[:0]
	return nil

}

func (r *serverRepository) GetMetric(ctx context.Context, metricID string) (models.Metrics, error) {
	var dbMetric models.Metrics
	row := r.db.DB.QueryRowContext(ctx, queryGetMetric, metricID)
	err := row.Scan(&dbMetric.ID, &dbMetric.MType, &dbMetric.Delta, &dbMetric.Value)
	if err != nil {
		return models.Metrics{}, err
	}
	return dbMetric, nil
}

func (r *serverRepository) GetMetricNames(ctx context.Context) ([]string, error) {
	var result []string
	rows, err := r.db.DB.QueryContext(ctx, queryGetMetricNames)
	if err != nil {
		return nil, err
	}

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

func (r *serverRepository) GetCounterMetric(ctx context.Context, metricID string) (*int64, error) {
	var result int64
	row := r.db.DB.QueryRowContext(ctx, queryGetCounterMetricValue, metricID)
	err := row.Scan(&result)
	if err != nil {
		r.logger.Error(err)
		return nil, err
	}
	return &result, nil
}

func (r *serverRepository) GetGaugeMetric(ctx context.Context, metricID string) (*float64, error) {
	var result float64
	row := r.db.DB.QueryRowContext(ctx, queryGetGaugeMetricValue, metricID)
	err := row.Scan(&result)
	if err != nil {
		r.logger.Error(err)
		return nil, err
	}
	return &result, nil
}
