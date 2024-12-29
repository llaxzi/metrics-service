package storage

import (
	"encoding/json"
	"metrics-service/internal/server/models"
	"os"
)

type DiskWriter interface {
	Save() error
}

type diskWriter struct {
	m        MetricsStorage
	filePath string
	encoder  *json.Encoder
}

func NewDiskWriter(metricsStorage MetricsStorage, filePath string) (DiskWriter, error) {
	return &diskWriter{m: metricsStorage, filePath: filePath}, nil
}

func (w *diskWriter) Save() error {
	file, err := os.OpenFile(w.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	w.encoder = json.NewEncoder(file)
	defer file.Close()

	metrics := w.m.GetMetricsJSON()
	return w.encoder.Encode(&metrics)
}

type DiskReader interface {
	Load() error
	Close() error
}

type diskReader struct {
	m       MetricsStorage
	file    *os.File
	decoder *json.Decoder
}

func NewDiskReader(metricsStorage MetricsStorage, filePath string) (DiskReader, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &diskReader{metricsStorage, file, json.NewDecoder(file)}, nil
}

func (r *diskReader) Load() error {
	var metrics []models.Metrics

	if err := r.decoder.Decode(&metrics); err != nil {
		if err.Error() == "EOF" {
			return nil
		}
		return err
	}
	r.m.SetMetricsJSON(metrics)
	return nil
}

func (r *diskReader) Close() error {
	return r.file.Close()
}
