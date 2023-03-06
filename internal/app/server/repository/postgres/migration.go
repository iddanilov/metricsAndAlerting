package postgresql

import (
	"context"

	"github.com/iddanilov/metricsAndAlerting/internal/pkg/logger"
	"github.com/iddanilov/metricsAndAlerting/internal/pkg/repository/postgresql"
)

func CreateTable(ctx context.Context, db postgresql.DB, logger logger.Logger) error {
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
	logger.Info("DB Create")

	return nil
}
