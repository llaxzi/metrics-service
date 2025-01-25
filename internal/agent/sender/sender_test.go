package sender

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"metrics-service/internal/server/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSender_Send(t *testing.T) {
	testTable := []struct {
		name       string
		metricsMap map[string]interface{}
		wantStatus int
		wantErr    bool
	}{
		{"OK", map[string]interface{}{
			"PollCount": uint64(40),
			"Alloc":     uint64(1234),
		}, http.StatusOK, false},

		{"Invalid metric type", map[string]interface{}{
			"PollCount": "invalid_type",
			"Alloc":     uint64(1234),
		}, http.StatusBadRequest, true},

		{
			name:       "Empty metrics map",
			metricsMap: map[string]interface{}{},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			// Создаем сервер
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				assert.Equal(t, http.MethodPost, r.Method)

				assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))

				// Проверяем URL
				pathParts := strings.Split(r.URL.Path, "/")
				if len(pathParts) != 5 {
					t.Errorf("Invalid URL path format: %s", r.URL.Path)
				}

				// Базово проверяем метрики

				metricType := pathParts[2]
				metricName := pathParts[3]

				if metricName == "PollCount" {
					assert.Equal(t, "counter", metricType)
				} else {
					assert.Equal(t, "gauge", metricType)
				}

				w.WriteHeader(test.wantStatus)

			}))
			defer server.Close()

			s := &sender{
				nil,
				server.URL,
				nil,
			}

			s.Send(test.metricsMap)
		})
	}

}

func TestSender_SendBatch(t *testing.T) {
	hashTestKey := []byte("test-key")
	// Создаём тестовый сервер
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

		hashHeader := r.Header.Get("HashSHA256")
		assert.NotEmpty(t, hashHeader)

		gz, err := gzip.NewReader(r.Body)
		assert.NoError(t, err)

		var receivedMetrics []models.Metrics
		err = json.NewDecoder(gz).Decode(&receivedMetrics)
		assert.NoError(t, err)

		wantMetrics := []models.Metrics{
			{ID: "PollCount", MType: "counter", Delta: func(v int64) *int64 { return &v }(42)},
			{ID: "RandomMetric", MType: "gauge", Value: func(v float64) *float64 { return &v }(3.14)},
		}
		assert.ElementsMatch(t, wantMetrics, receivedMetrics)

		jsonData, err := json.Marshal(wantMetrics)
		assert.NoError(t, err)

		wantHash := generateTestHash(jsonData, hashTestKey)
		assert.Equal(t, wantHash, hashHeader)

		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	metricsMap := map[string]interface{}{
		"PollCount":    int64(42),
		"RandomMetric": float64(3.14),
	}

	client := resty.New()
	sender := &sender{
		client:  client,
		baseURL: testServer.URL,
		hashKey: hashTestKey,
	}

	err := sender.SendBatch(metricsMap)
	assert.NoError(t, err)
}

func generateTestHash(src []byte, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(src)
	return hex.EncodeToString(h.Sum(nil))
}
