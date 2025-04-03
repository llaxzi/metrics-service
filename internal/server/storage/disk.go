package storage

import (
	"encoding/json"
	"os"

	"metrics-service/internal/server/models"
)

// DiskWriter определяет интерфейс для сохранения метрик на диск.
type DiskWriter interface {
	Save(metrics []models.Metrics) error
}

// diskWriter реализует интерфейс DiskWriter и сохраняет метрики в JSON-файл.
type diskWriter struct {
	filePath string
	encoder  *json.Encoder
}

// NewDiskWriter создает новый экземпляр DiskWriter.
func NewDiskWriter(filePath string) (DiskWriter, error) {
	return &diskWriter{filePath: filePath}, nil
}

// Save сохраняет массив метрик в файл в формате JSON.
func (w *diskWriter) Save(metrics []models.Metrics) error {
	file, err := os.OpenFile(w.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	w.encoder = json.NewEncoder(file)
	defer file.Close()
	return w.encoder.Encode(&metrics)
}

// DiskReader определяет интерфейс для загрузки метрик с диска.
type DiskReader interface {
	Load(m *metricsStorage) error
	Close() error
}

// diskReader реализует интерфейс DiskReader и загружает метрики из JSON-файла.
type diskReader struct {
	file    *os.File
	decoder *json.Decoder
}

// NewDiskReader создает новый экземпляр diskReader.
func NewDiskReader(filePath string) (DiskReader, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &diskReader{file, json.NewDecoder(file)}, nil
}

// Load загружает метрики из файла и записывает их в хранилище.
func (r *diskReader) Load(m *metricsStorage) error {
	var metrics []models.Metrics

	if err := r.decoder.Decode(&metrics); err != nil {
		if err.Error() == "EOF" {
			return nil
		}
		return err
	}
	m.setMetricsJSON(metrics)
	return nil
}

// Close закрывает файл после чтения.
func (r *diskReader) Close() error {
	return r.file.Close()
}
