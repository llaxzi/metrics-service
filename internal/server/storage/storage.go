package storage

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"metrics-service/internal/server/models"
	"sync"
)

// TODO: внешний контекст к методам

type Storage interface {
	Update(ctx context.Context, metricType, metricName, metricValStr string) error
	UpdateJSON(ctx context.Context, metric *models.Metrics) error
	UpdateBatch(ctx context.Context, metrics []models.Metrics) error
	Get(ctx context.Context, metricType, metricName string) (string, error)
	GetJSON(ctx context.Context, metric *models.Metrics) error
	GetMetrics(ctx context.Context) ([][]string, error)
	Ping(ctx context.Context) error
	Bootstrap(ctx context.Context) error
	Save() error
	Close() error
}

func NewStorage(flagDatabaseDSN, flagFileStoragePath string, flagRestore bool, flagStoreInterval int) (Storage, error) {

	if flagDatabaseDSN != "" {
		db, err := sql.Open("pgx", flagDatabaseDSN)
		if err != nil {
			return nil, err
		}
		return &repository{db}, nil
	}

	var diskW DiskWriter
	if flagFileStoragePath != "" {
		var err error
		diskW, err = NewDiskWriter(flagFileStoragePath)
		if err != nil {
			return nil, err
		}
	}

	memoryStorage := &metricsStorage{sync.RWMutex{}, sync.RWMutex{}, make(map[string]float64), make(map[string]int64), diskW}

	// Загружаем storage из файла, если необходимо
	if flagRestore && flagFileStoragePath != "" {
		reader, err := NewDiskReader(flagFileStoragePath)
		if err != nil {
			log.Fatalf("Failed to restore metrics from %v: %v", flagFileStoragePath, err)
		}
		err = reader.Load(memoryStorage)
		if err != nil {
			log.Fatalf("Failed to restore metrics from %v: %v", flagFileStoragePath, err)
		}
		err = reader.Close()
		if err != nil {
			log.Fatalf("Failed to close diskReader: %v", err)
		}
		log.Printf("Read metrics from file: %v\n", flagFileStoragePath)
	}

	return memoryStorage, nil
}
