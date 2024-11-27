package sender

import (
	"github.com/stretchr/testify/assert"
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
				if len(pathParts) != 4 {
					t.Errorf("Invalid URL path format: %s", r.URL.Path)
				}

				// Базово проверяем метрики

				metricType := pathParts[1]
				metricName := pathParts[2]

				if metricName == "PollCount" {
					assert.Equal(t, "counter", metricType)
				} else {
					assert.Equal(t, "gauge", metricType)
				}

				w.WriteHeader(test.wantStatus)

			}))
			defer server.Close()

			s := &sender{
				server.Client(),
				server.URL,
			}

			s.Send(test.metricsMap)
		})
	}

}
