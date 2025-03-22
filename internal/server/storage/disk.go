package storage

import (
	"encoding/json"
	"os"

	"metrics-service/internal/server/models"
)

type DiskWriter interface {
	Save(metrics []models.Metrics) error
}

type diskWriter struct {
	filePath string
	encoder  *json.Encoder
}

func NewDiskWriter(filePath string) (DiskWriter, error) {
	return &diskWriter{filePath: filePath}, nil
}

func (w *diskWriter) Save(metrics []models.Metrics) error {
	file, err := os.OpenFile(w.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	w.encoder = json.NewEncoder(file)
	defer file.Close()
	return w.encoder.Encode(&metrics)
}

type DiskReader interface {
	Load(m *metricsStorage) error
	Close() error
}

type diskReader struct {
	file    *os.File
	decoder *json.Decoder
}

func NewDiskReader(filePath string) (DiskReader, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &diskReader{file, json.NewDecoder(file)}, nil
}

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

func (r *diskReader) Close() error {
	return r.file.Close()
}
