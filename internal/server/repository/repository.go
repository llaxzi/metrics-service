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
	"metrics-service/internal/server/storage"

	"time"
)

type Repository interface {
	Close() error
	Ping() error
	Save() error
	Bootstrap() error
}

const (
	pingTimeout   = 1
	insertTimeout = 1
	bootstrapTimeout
)

type repository struct {
	db       *sql.DB
	mStorage storage.MetricsStorage
}

func NewRepository(dsn string, mStorage storage.MetricsStorage) (Repository, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	return &repository{db, mStorage}, nil
}

func (r *repository) Close() error {
	return r.db.Close()
}

func (r *repository) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout*time.Second)
	defer cancel()
	err := r.db.PingContext(ctx)
	if err != nil {
		if r.isPgConnErr(err) {
			return apperrors.ErrPgConnExc
		}
	}
	return err
}

// Save выполняет batch вставку в бд
func (r *repository) Save() error {
	/*
		Warning: лимит параметров запроса в postgres = 65.535 параметров.
		Метод может обновить до 65535 / 4 = 16383 метрик.
		Для большего количества метрик (возможно ли такое даже в крупном проекте!? - сомневаюсь) нужно разбивать на чанки в сервисе.
	*/

	metrics := r.mStorage.GetMetricsJSON()
	if len(metrics) < 1 {
		return nil
	}

	// Формируем UPSERT sql запрос
	query := "INSERT INTO public.metrics(metric_id,metric_type,delta,value) VALUES "
	values := make([]interface{}, 0, len(metrics)*4)
	paramIdx := 1
	for i, metric := range metrics {
		values = append(values, metric.ID, metric.MType, metric.Delta, metric.Value)
		query += fmt.Sprintf("($%d,$%d,$%d,$%d)", paramIdx, paramIdx+1, paramIdx+2, paramIdx+3)
		paramIdx += 4
		if i != len(metrics)-1 {
			query += ","
		}
	}
	// обновление существующих записей
	query += "ON CONFLICT (metric_id) DO UPDATE SET delta = EXCLUDED.delta, value = EXCLUDED.value;"

	ctx, cancel := context.WithTimeout(context.Background(), insertTimeout*time.Second)
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

// Bootstrap устанавливает бд в debug окружении
func (r *repository) Bootstrap() error {
	ctx, cancel := context.WithTimeout(context.Background(), bootstrapTimeout*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
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

func (r *repository) isPgConnErr(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code)
}
