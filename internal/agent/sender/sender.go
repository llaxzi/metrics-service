// Package sender предоставляет функциональность для отправки метрик на сервер.
// Поддерживает отправку метрик в формате JSON, а также пакетную отправку метрик и их сжатие.
package sender

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"

	"metrics-service/internal/server/models"
)

// ISender представляет интерфейс для отправки метрик на сервер.
type ISender interface {
	// SendJSON отправляет одну метрику в формате JSON.
	SendJSON(metricName string, metricValI interface{}) error
	// SendBatch отправляет несколько метрик за один запрос.
	SendBatch(metricsMap map[string]interface{}) error
}

// sender реализует интерфейс Sender.
type sender struct {
	client        *resty.Client
	baseURL       string
	hashKey       []byte
	cryptoKeyPath string
	publicKey     *rsa.PublicKey
}

// NewSender создает новый экземпляр sender с заданным базовым URL и ключом для хеширования.
func NewSender(baseURL string, hashKey []byte, cryptoKeyPath string) (ISender, error) {
	client := resty.New()
	// Настройка retry
	client.SetRetryCount(3)
	client.SetRetryAfter(func(client *resty.Client, response *resty.Response) (time.Duration, error) {
		retryCount := response.Request.Attempt
		switch retryCount {
		case 1:
			return 1 * time.Second, nil
		case 2:
			return 3 * time.Second, nil
		case 3:
			return 5 * time.Second, nil
		default:
			return 0, nil
		}
	})
	client.AddRetryCondition(func(response *resty.Response, err error) bool {
		if response.StatusCode() == http.StatusServiceUnavailable || response.StatusCode() == http.StatusInternalServerError {
			return true
		}
		return false
	}) // retry только в случае, если сервер недоступен (maintenance или перегрузка) или внут. ошибка

	s := &sender{
		client:        client,
		baseURL:       baseURL,
		hashKey:       hashKey,
		cryptoKeyPath: cryptoKeyPath,
	}

	if err := s.loadPublicKey(); err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	return s, nil

}

// Send отправляет метрики на сервер. Каждая метрика может быть типа "counter" или "gauge".
func (s *sender) Send(metricsMap map[string]interface{}) error {
	for metricName, metricVal := range metricsMap {

		var metricType string
		if metricName == "PollCount" {
			metricType = "counter"
		} else {
			metricType = "gauge"
		}

		var metricValStr string
		switch v := metricVal.(type) {
		case uint64:
			metricValStr = strconv.FormatUint(metricVal.(uint64), 10)
		case uint32:
			metricValStr = strconv.FormatUint(uint64(metricVal.(uint32)), 10)
		case float64:
			metricValStr = strconv.FormatFloat(metricVal.(float64), 'f', -1, 64)
		default:
			return fmt.Errorf("wrong metric value type: %v", v)
		}

		url := s.baseURL + "/update/" + metricType + "/" + metricName + "/" + metricValStr

		client := resty.New()
		client.SetHeader("Content-type", "text/plain")

		resp, err := client.R().Post(url)
		if err != nil {
			return fmt.Errorf("failed to send request: %v", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("request %v failed: %v", url, err)
		}

	}
	fmt.Println("All metrics send to server")
	return nil
}

// SendJSON отправляет метрику в формате JSON.
func (s *sender) SendJSON(metricName string, metricValI interface{}) error {

	var body models.Metrics

	if metricName == "PollCount" {
		metricType := "counter"

		metricVal, ok := metricValI.(int64)
		if !ok {
			return fmt.Errorf("invalid type for counter metric: %v", metricName)
		}

		body = models.Metrics{ID: metricName, MType: metricType, Delta: &metricVal}
	} else {
		metricType := "gauge"

		metricVal, ok := metricValI.(float64)
		if !ok {
			return fmt.Errorf("invalid type for gauge metric: %v", metricName)
		}

		body = models.Metrics{ID: metricName, MType: metricType, Value: &metricVal}
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// Хеш в заголовке
	var hash string
	if len(s.hashKey) > 0 {
		hash, err = s.generateHash(jsonData)
		if err != nil {
			return fmt.Errorf("failed to generate hash: %w", err)
		}
		s.client.SetHeader("HashSHA256", hash)
	}

	// Сжатие данных в gzip
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	if _, err = gzipWriter.Write(jsonData); err != nil {
		return fmt.Errorf("failed to gzip data: %v", err)
	}
	if err = gzipWriter.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %v", err)
	}

	url := s.baseURL + "/update"

	client := resty.New()
	client.SetHeader("Content-type", "application/json")
	client.SetHeader("Content-Encoding", "gzip")

	resp, err := client.R().SetBody(buf.Bytes()).Post(url)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("request %v failed: %v", url, err)
	}

	fmt.Println("Metric send to server")
	return nil
}

// SendBatch отправляет несколько метрик за один запрос.
func (s *sender) SendBatch(metricsMap map[string]interface{}) error {
	if len(metricsMap) < 1 {
		return nil
	}
	var metrics []models.Metrics
	for metricName, metricValI := range metricsMap {
		var metric models.Metrics
		if metricName == "PollCount" {
			metricType := "counter"

			metricVal, ok := metricValI.(int64)
			if !ok {
				return fmt.Errorf("invalid type for counter metric: %v", metricName)
			}

			metric = models.Metrics{ID: metricName, MType: metricType, Delta: &metricVal}
		} else {
			metricType := "gauge"

			metricVal, ok := metricValI.(float64)
			if !ok {
				return fmt.Errorf("invalid type for gauge metric: %v", metricName)
			}

			metric = models.Metrics{ID: metricName, MType: metricType, Value: &metricVal}
		}
		metrics = append(metrics, metric)
	}

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Хеш в заголовке
	var hash string
	if len(s.hashKey) > 0 {
		hash, err = s.generateHash(jsonData)
		if err != nil {
			return fmt.Errorf("failed to generate hash: %w", err)
		}
		s.client.SetHeader("HashSHA256", hash)
	}

	// Сжатие данных в gzip
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	if _, err = gzipWriter.Write(jsonData); err != nil {
		return fmt.Errorf("failed to gzip data: %w", err)
	}
	if err = gzipWriter.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	// шифруем gzip-данные
	var data []byte

	if s.publicKey == nil {
		data = buf.Bytes()
		s.client.SetHeader("Content-Encoding", "gzip")
	} else {
		encryptedData, err := s.encryptOAEPChunks(buf.Bytes())
		if err != nil {
			return fmt.Errorf("failed to encrypt data chunks: %w", err)
		}
		data = encryptedData
		s.client.SetHeader("Content-type", "application/octet-stream")
		s.client.SetHeader("Content-Encoding", "encrypted")
	}

	url := s.baseURL + "/updates"

	resp, err := s.client.R().SetBody(data).Post(url)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("request %v failed: %w", url, err)
	}
	return nil
}

func (s *sender) generateHash(src []byte) (string, error) {
	h := hmac.New(sha256.New, s.hashKey)
	_, err := h.Write(src)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (s *sender) loadPublicKey() error {

	if s.cryptoKeyPath == "" {
		s.publicKey = nil
		return nil
	}

	keyData, err := os.ReadFile(s.cryptoKeyPath)
	if err != nil {
		return err
	}
	block, _ := pem.Decode(keyData)
	if block == nil {
		return errors.New("failed to decode PEM block")
	}

	pubKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	pubKey, ok := pubKeyInterface.(*rsa.PublicKey)
	if !ok {
		return errors.New("not RSA public key")
	}

	s.publicKey = pubKey
	return nil
}

func (s *sender) encryptOAEPChunks(data []byte) ([]byte, error) {
	hash := sha256.New()
	keySize := s.publicKey.Size()
	maxChunkSize := keySize - 2*hash.Size() - 2 // RSA-OAEP ограничение

	var encrypted []byte
	for start := 0; start < len(data); start += maxChunkSize {
		end := start + maxChunkSize
		if end > len(data) {
			end = len(data)
		}

		chunk := data[start:end]
		encChunk, err := rsa.EncryptOAEP(hash, rand.Reader, s.publicKey, chunk, nil)
		if err != nil {
			return nil, fmt.Errorf("chunk encryption failed: %w", err)
		}

		encrypted = append(encrypted, encChunk...)
	}

	return encrypted, nil
}
