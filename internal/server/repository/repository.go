package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"metrics-service/internal/server/storage"
	"time"
)

type Repository interface {
	Close() error
	Ping() error
	Save() error
	CreateMetricsTable() error
}

const (
	pingTimeout   = 1
	insertTimeout = 1
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
	return r.db.PingContext(ctx)
}

// Save выполняет batch вставку в бд
func (r *repository) Save() error {
	/*
		Warning: лимит параметров запроса в postgres = 65.535 параметров.
		Метод может обновить до 65535 / 4 = 16383 метрик.
		Для большего количества метрик (возможно ли такое даже в крупном проекте!?) нужно разбивать на чанки в сервисе.
	*/

	metrics := r.mStorage.GetMetricsJSON()
	if len(metrics) < 1 {
		return nil
	}

	// Формируем UPSERT sql запрос
	query := "INSERT INTO public.metrics(metric_id,metric_type,delta,value) VALUES "
	var values []interface{}
	paramIdx := 1
	for _, metric := range metrics {
		query += fmt.Sprintf("($%d,$%d,$%d,$%d),", paramIdx, paramIdx+1, paramIdx+2, paramIdx+3)
		paramIdx += 4
		values = append(values, metric.ID, metric.MType, metric.Delta, metric.Value)
	}
	query = query[:len(query)-1] + " " // обрезаем последнюю ','
	// обновление существующих записей
	query += "ON CONFLICT (metric_id) DO UPDATE SET delta = EXCLUDED.delta, value = EXCLUDED.value;"

	ctx, cancel := context.WithTimeout(context.Background(), insertTimeout*time.Second)
	defer cancel()
	execContext, err := r.db.ExecContext(ctx, query, values...)
	if err != nil {
		return err
	}

	rowsAff, err := execContext.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff != int64(len(metrics)) {
		return errors.New("rows affected != len(metrics)")
	}

	return nil
}

// CreateMetricsTable устанавливает бд в debug окружении
func (r *repository) CreateMetricsTable() error {
	//Удаляем enum тип и таблицу, если существуют
	dropTableQuery := `DROP TABLE IF EXISTS public.metrics;`
	_, err := r.db.Exec(dropTableQuery)
	if err != nil {
		return fmt.Errorf("failed to drop table metrics: %v", err)
	}
	dropTypeQuery := `DROP TYPE IF EXISTS MType;`
	_, err = r.db.Exec(dropTypeQuery)
	if err != nil {
		return fmt.Errorf("failed to drop type MType: %v", err)
	}

	// Создаем enum тип
	createTypeQuery := `CREATE TYPE MType AS ENUM ('gauge', 'counter');`
	_, err = r.db.Exec(createTypeQuery)
	if err != nil {
		return fmt.Errorf("failed to create enum MType: %v", err)
	}

	// Создаем таблицу
	createTableQuery := `CREATE TABLE IF NOT EXISTS public.metrics (
		metric_id VARCHAR(100) PRIMARY KEY,
		metric_type MType,
		delta INT,
		value DOUBLE PRECISION);`
	_, err = r.db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create table metrics: %v", err)
	}
	return nil
}
