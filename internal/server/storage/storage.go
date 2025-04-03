// Package storage содержит интерфейс хранилища метрик и его реализации.
package storage

import (
	"context"
	"database/sql"
	"log"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"

	"metrics-service/internal/server/models"
)

// Storage определяет интерфейс хранилища метрик.
type Storage interface {
	// Update обновляет метрику в хранилище на основе переданных параметров.
	Update(ctx context.Context, metricType, metricName, metricValStr string) error
	// UpdateJSON обновляет метрику в хранилище, получая её в виде структуры models.Metrics.
	UpdateJSON(ctx context.Context, metric *models.Metrics) error
	// UpdateBatch выполняет пакетное обновление метрик.
	UpdateBatch(ctx context.Context, metrics []models.Metrics) error
	// Get получает значение метрики по её имени и типу.
	Get(ctx context.Context, metricType, metricName string) (string, error)
	// GetJSON получает значение метрики по структуре models.Metrics
	GetJSON(ctx context.Context, metric *models.Metrics) error
	// GetMetrics получает все метрики.
	GetMetrics(ctx context.Context) ([][]string, error)
	// Ping проверяет соединение с базой данных.
	Ping(ctx context.Context) error
	// Bootstrap подготавливает окружение хранилища (используется в debug-окружении).
	Bootstrap(ctx context.Context) error
	// Save сохраняет метрики, если имплементация хранилища требует ручного сохранения.
	Save() error
	// Close закрывает соединение с хранилищем, если имплементация хранилища требует закрытия соединения.
	Close() error
}

// NewStorage создает новый экземпляр Storage
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
