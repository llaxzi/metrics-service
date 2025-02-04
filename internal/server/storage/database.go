package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"log"
	apperrors "metrics-service/internal/server/errors"
	"metrics-service/internal/server/models"
	"strconv"
	"time"
)

const (
	timeout = 1
)

type repository struct {
	db *sql.DB
}

func (r *repository) Update(ctx context.Context, metricType, metricName, metricValStr string) error {
	var delta *int64
	var value *float64
	switch metricType {
	case "counter":
		metricVal, err := strconv.ParseInt(metricValStr, 10, 64)
		if err != nil {
			return apperrors.ErrWrongMetricValue
		}
		delta = &metricVal
		value = nil
	case "gauge":
		metricVal, err := strconv.ParseFloat(metricValStr, 64)
		if err != nil {
			return apperrors.ErrWrongMetricValue
		}
		value = &metricVal

	}
	metrics := []models.Metrics{
		{
			ID:    metricName,
			MType: metricType,
			Delta: delta,
			Value: value,
		},
	}
	return r.UpdateBatch(ctx, metrics)
}

func (r *repository) UpdateJSON(ctx context.Context, metric *models.Metrics) error {
	metrics := []models.Metrics{*metric}
	return r.UpdateBatch(ctx, metrics)
}

// UpdateBatch выполняет batch вставку в бд
func (r *repository) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
	if len(metrics) < 1 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, timeout*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		if r.isPgConnErr(err) {
			return apperrors.ErrPgConnExc
		}
		return fmt.Errorf("failed to start tx: %w", err)
	}

	query := "INSERT INTO public.metrics(metric_id, metric_type, delta, value) VALUES ($1, $2, $3, $4)"
	query += " ON CONFLICT (metric_id) DO UPDATE SET delta = metrics.delta + EXCLUDED.delta, value = EXCLUDED.value;"

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		tx.Rollback()
		if r.isPgConnErr(err) {
			return apperrors.ErrPgConnExc
		}
		log.Printf("failed to prepare statement: %v", err)
		return apperrors.ErrServer
	}
	defer stmt.Close()

	for _, metric := range metrics {
		_, err = stmt.ExecContext(ctx, metric.ID, metric.MType, metric.Delta, metric.Value)
		if err != nil {
			tx.Rollback()
			if r.isPgConnErr(err) {
				return apperrors.ErrPgConnExc
			}
			log.Printf("failed to execute statement for metric %s: %v", metric.ID, err)
			return apperrors.ErrServer
		}
	}
	return tx.Commit()
}

func (r *repository) Get(ctx context.Context, metricType, metricName string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout*time.Second)
	defer cancel()

	var delta *int64
	var value *float64

	query := `SELECT metric_id, metric_type, delta, value FROM public.metrics WHERE metric_type = $1 AND metric_id = $2`
	row := r.db.QueryRowContext(ctx, query, metricType, metricName)
	err := row.Scan(metricName, metricType, &delta, &value)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("metric not found: %w", err)
		}
		if r.isPgConnErr(err) {
			return "", apperrors.ErrPgConnExc
		}
		log.Printf("failed to scan metric: %v", err)
		return "", apperrors.ErrServer
	}

	switch metricType {
	case "counter":
		if delta == nil {
			return "", apperrors.ErrWrongMetricValue
		}
		return strconv.FormatInt(*delta, 10), nil
	case "gauge":
		if value == nil {
			return "", apperrors.ErrWrongMetricValue
		}
		return strconv.FormatFloat(*value, 'f', -1, 64), nil
	default:
		return "", apperrors.ErrInvalidMetricType
	}

}

func (r *repository) GetJSON(ctx context.Context, metric *models.Metrics) error {
	ctx, cancel := context.WithTimeout(ctx, timeout*time.Second)
	defer cancel()

	query := `SELECT metric_id, metric_type, delta, value FROM public.metrics WHERE metric_type = $1 AND metric_id = $2`
	row := r.db.QueryRowContext(ctx, query, metric.MType, metric.ID)
	err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.ErrMetricNotExist
		}
		if r.isPgConnErr(err) {
			return apperrors.ErrPgConnExc
		}
		log.Printf("failed to scan metric: %v", err)
		return apperrors.ErrServer
	}
	return nil
}

func (r *repository) GetMetrics(ctx context.Context) ([][]string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout*time.Second)
	defer cancel()

	query := `SELECT metric_id, metric_type, delta, value FROM public.metrics`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		if r.isPgConnErr(err) {
			return nil, apperrors.ErrPgConnExc
		}
		log.Printf("failed to query metrics: %v", err)
		return nil, apperrors.ErrServer
	}
	defer rows.Close()

	var metrics [][]string

	for rows.Next() {
		var metricID string
		var metricType string
		var delta sql.NullInt64
		var value sql.NullFloat64

		err = rows.Scan(&metricID, &metricType, &delta, &value)
		if err != nil {
			log.Printf("failed to scan row: %v", err)
			return nil, apperrors.ErrServer
		}

		var metricValue string
		if metricType == "gauge" && value.Valid {
			metricValue = fmt.Sprintf("%f", value.Float64)
		} else if metricType == "counter" && delta.Valid {
			metricValue = fmt.Sprintf("%d", delta.Int64)
		} else {
			metricValue = "null"
		}

		metrics = append(metrics, []string{metricID, metricValue})
	}

	if err = rows.Err(); err != nil {
		log.Printf("row iteration error: %v", err)
		return nil, apperrors.ErrServer
	}

	return metrics, nil
}

func (r *repository) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, timeout*time.Second)
	defer cancel()
	err := r.db.PingContext(ctx)
	if err != nil {
		if r.isPgConnErr(err) {
			return apperrors.ErrPgConnExc
		}
		log.Printf("err on ping: %v", err)
		return apperrors.ErrServer
	}
	return nil
}

// Bootstrap устанавливает бд в debug окружении
func (r *repository) Bootstrap(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, timeout*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		if r.isPgConnErr(err) {
			return apperrors.ErrPgConnExc
		}
		return fmt.Errorf("failed to start tx: %w", err)
	}
	//Удаляем enum тип и таблицу, если существуют
	dropTableQuery := `DROP TABLE IF EXISTS public.metrics;`
	_, err = tx.ExecContext(ctx, dropTableQuery)
	if err != nil {
		tx.Rollback()
		if r.isPgConnErr(err) {
			return apperrors.ErrPgConnExc
		}
		return fmt.Errorf("failed to drop table metrics: %w", err)
	}
	dropTypeQuery := `DROP TYPE IF EXISTS MType;`
	_, err = tx.ExecContext(ctx, dropTypeQuery)
	if err != nil {
		tx.Rollback()
		if r.isPgConnErr(err) {
			return apperrors.ErrPgConnExc
		}
		return fmt.Errorf("failed to drop type MType: %w", err)
	}

	// Создаем enum тип
	createTypeQuery := `CREATE TYPE MType AS ENUM ('gauge', 'counter');`
	_, err = tx.ExecContext(ctx, createTypeQuery)
	if err != nil {
		tx.Rollback()
		if r.isPgConnErr(err) {
			return apperrors.ErrPgConnExc
		}
		return fmt.Errorf("failed to create enum MType: %w", err)
	}

	// Создаем таблицу
	createTableQuery := `CREATE TABLE IF NOT EXISTS public.metrics (
		metric_id VARCHAR(100) PRIMARY KEY,
		metric_type MType,
		delta BIGINT,
		value DOUBLE PRECISION);`
	_, err = tx.ExecContext(ctx, createTableQuery)
	if err != nil {
		tx.Rollback()
		if r.isPgConnErr(err) {
			return apperrors.ErrPgConnExc
		}
		return fmt.Errorf("failed to create table metrics: %w", err)
	}
	return tx.Commit()
}

func (r *repository) Save() error {
	return nil
}

func (r *repository) Close() error {
	return r.db.Close()
}

// internal

func (r *repository) isPgConnErr(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code)
}
