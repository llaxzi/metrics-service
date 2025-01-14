package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	apperrors "metrics-service/internal/server/errors"
	"metrics-service/internal/server/models"
	"time"
)

type Repository interface {
	Close() error
	Ping() error
	Save(metrics []models.Metrics) error
	Bootstrap() error
	Get(metricType, metricName string) (models.Metrics, error)
	GetSlice() ([][]string, error)
}

const (
	timeout = 1
)

type repository struct {
	db *sql.DB
}

func NewRepository(dsn string) (Repository, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	return &repository{db}, nil
}

func (r *repository) Close() error {
	return r.db.Close()
}

func (r *repository) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	err := r.db.PingContext(ctx)
	if err != nil {
		if r.isPgConnErr(err) {
			return apperrors.ErrPgConnExc
		}
	}
	return err
}

// SaveAll выполняет batch вставку в бд одним sql запросом
func (r *repository) SaveAll(metrics []models.Metrics) error {
	if len(metrics) < 1 {
		return nil
	}

	// Формируем UPSERT sql запрос
	query := "INSERT INTO public.metrics(metric_id, metric_type, delta, value) VALUES "
	values := make([]interface{}, 0, len(metrics)*4)
	paramIdx := 1
	for i, metric := range metrics {
		values = append(values, metric.ID, metric.MType, metric.Delta, metric.Value)
		query += fmt.Sprintf("($%d, $%d, $%d, $%d)", paramIdx, paramIdx+1, paramIdx+2, paramIdx+3)
		paramIdx += 4
		if i != len(metrics)-1 {
			query += ", "
		}
	}
	query += " ON CONFLICT (metric_id) DO UPDATE SET delta = metrics.delta + EXCLUDED.delta, value = EXCLUDED.value;"

	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		if r.isPgConnErr(err) {
			return apperrors.ErrPgConnExc
		}
		return fmt.Errorf("failed to start tx: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		tx.Rollback()
		if r.isPgConnErr(err) {
			return apperrors.ErrPgConnExc
		}
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	result, err := stmt.ExecContext(ctx, values...)
	if err != nil {
		tx.Rollback()
		if r.isPgConnErr(err) {
			return apperrors.ErrPgConnExc
		}
		return fmt.Errorf("failed to execute statement: %w", err)
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		if r.isPgConnErr(err) {
			return apperrors.ErrPgConnExc
		}
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAff != int64(len(metrics)) {
		tx.Rollback()
		if r.isPgConnErr(err) {
			return apperrors.ErrPgConnExc
		}
		return errors.New("rows affected != len(metrics)")
	}

	return tx.Commit()
}

// Save выполняет batch вставку в бд
func (r *repository) Save(metrics []models.Metrics) error {
	if len(metrics) < 1 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
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
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, metric := range metrics {
		_, err = stmt.ExecContext(ctx, metric.ID, metric.MType, metric.Delta, metric.Value)
		if err != nil {
			tx.Rollback()
			if r.isPgConnErr(err) {
				return apperrors.ErrPgConnExc
			}
			return fmt.Errorf("failed to execute statement for metric %s: %w", metric.ID, err)
		}
	}
	return tx.Commit()
}

// Bootstrap устанавливает бд в debug окружении
func (r *repository) Bootstrap() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
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

func (r *repository) Get(metricType, metricName string) (models.Metrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	var metric models.Metrics

	query := `SELECT metric_id, metric_type, delta, value FROM public.metrics WHERE metric_type = $1 AND metric_id = $2`
	row := r.db.QueryRowContext(ctx, query, metricType, metricName)
	err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Metrics{}, fmt.Errorf("metric not found: %w", err)
		}
		if r.isPgConnErr(err) {
			return models.Metrics{}, apperrors.ErrPgConnExc
		}
		return models.Metrics{}, fmt.Errorf("failed to retrieve metric: %w", err)
	}
	return metric, nil
}

func (r *repository) GetSlice() ([][]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	query := `SELECT metric_id, metric_type, delta, value FROM public.metrics`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		if r.isPgConnErr(err) {
			return nil, apperrors.ErrPgConnExc
		}
		return nil, fmt.Errorf("failed to query metrics: %w", err)
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
			return nil, fmt.Errorf("failed to scan row: %w", err)
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
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return metrics, nil
}

func (r *repository) isPgConnErr(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code)
}
