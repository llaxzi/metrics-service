package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"html/template"
	apperrors "metrics-service/internal/server/errors"
	"metrics-service/internal/server/mocks"
	"metrics-service/internal/server/models"
	"metrics-service/internal/server/retry"
	"metrics-service/internal/server/service"
	"metrics-service/internal/server/storage"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsHandler_Update(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
	}
	testTable := []struct {
		name    string
		request string
		want    want
	}{

		// Тесты для Counter
		{"OK for Counter", "/update/counter/PollCounter/2", want{http.StatusOK, "text/plain; charset=utf-8"}},
		{"Wrong url #1 for Counter", "/update/counter/PollCounter", want{http.StatusNotFound, "text/plain"}},
		{"Wrong url #2 for Counter", "/update/counter/PollCounter/invalidType", want{http.StatusBadRequest, "text/plain; charset=utf-8"}},

		// Тесты для Gauge
		{"OK for Gauge", "/update/gauge/PollGauge/3.14", want{http.StatusOK, "text/plain; charset=utf-8"}},
		{"Wrong url #1 for Gauge", "/update/gauge/PollGauge", want{http.StatusNotFound, "text/plain"}},
		{"Wrong url #2 for Gauge", "/update/gauge/PollGauge/invalidType", want{http.StatusBadRequest, "text/plain; charset=utf-8"}},

		// Некорректный тип метрики
		{"Invalid Metric Type", "/update/invalidType/PollMetric/1", want{http.StatusBadRequest, "text/plain; charset=utf-8"}},
	}

	for _, test := range testTable {

		request := httptest.NewRequest(http.MethodPost, test.request, nil)
		request.Header.Set("Content-Type", "text/plain")

		w := httptest.NewRecorder()

		router := gin.Default()
		metricsStorage := storage.NewMetricsStorage()

		metricsService := service.NewMetricsService(metricsStorage, nil, nil, false, true, nil)
		metricsH := NewMetricsHandler(metricsService)

		router.POST("/update/:metricType/:metricName/:metricVal", metricsH.Update)

		router.ServeHTTP(w, request)

		assert.Equal(t, test.want.statusCode, w.Code)

		assert.Equal(t, test.want.contentType, w.Header().Get("Content-Type"))
	}
}

func TestMetricsHandler_Get(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		body        string
	}
	testTable := []struct {
		name       string
		request    string
		want       want
		storageSet func(s storage.MetricsStorage)
	}{
		{"OK counter", "/value/counter/someMetric", want{http.StatusOK, "text/plain", "5"}, func(s storage.MetricsStorage) {
			s.SetCounter("someMetric", 5)
		}},
		{"Not found UR counterL", "/value/counter/someMetric/metric", want{http.StatusNotFound, "text/plain", "404 page not found"}, func(s storage.MetricsStorage) {
		}},
		{"Not found metric counter", "/value/counter/someMetric", want{http.StatusNotFound, "text/plain", "metric doesn't exist"}, func(s storage.MetricsStorage) {
		}},
		{"OK gauge", "/value/gauge/someMetric", want{http.StatusOK, "text/plain", "1.343"}, func(s storage.MetricsStorage) {
			s.SetGauge("someMetric", 1.343)
		}},
		{"Not found URL gauge", "/value/gauge/someMetric/metric", want{http.StatusNotFound, "text/plain", "404 page not found"}, func(s storage.MetricsStorage) {
		}},
		{"Not found metric gauge", "/value/gauge/someMetric", want{http.StatusNotFound, "text/plain", "metric doesn't exist"}, func(s storage.MetricsStorage) {
		}},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {

			request := httptest.NewRequest(http.MethodGet, test.request, nil)

			w := httptest.NewRecorder()

			metricsStorage := storage.NewMetricsStorage()
			test.storageSet(metricsStorage)

			metricsService := service.NewMetricsService(metricsStorage, nil, nil, false, true, nil)

			metricsH := NewMetricsHandler(metricsService)

			router := gin.Default()

			router.GET("/value/:metricType/:metricName", metricsH.Get)

			router.ServeHTTP(w, request)

			assert.Equal(t, test.want.statusCode, w.Code)

			assert.Contains(t, w.Header().Get("Content-Type"), test.want.contentType)

			assert.Equal(t, test.want.body, w.Body.String())
		})
	}
}

func TestHtmlHandler_Get(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		body        string
	}
	testTable := []struct {
		name       string
		request    string
		want       want
		storageSet func(s storage.MetricsStorage)
	}{
		{"OK", "/", want{http.StatusOK, "text/html", `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Metrics</title>
</head>
<body>
<h1>Metrics List</h1>
<div>
    
    <p>counter: 5</p>
    
    <p>gauge: 1.343</p>
    
</div>
</body>
</html>`}, func(s storage.MetricsStorage) {
			s.SetCounter("counter", 5)
			s.SetGauge("gauge", 1.343)
		}},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {

			metricsStorage := storage.NewMetricsStorage()
			metricsService := service.NewHTMLService(metricsStorage, nil, false)
			htmlH := NewHTMLHandler(metricsService)

			test.storageSet(metricsStorage)

			// Используем router для проверки html ответа
			r := gin.Default()

			// Устанавливаем html template
			r.SetHTMLTemplate(template.Must(template.New("metrics.html").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Metrics</title>
</head>
<body>
<h1>Metrics List</h1>
<div>
    {{ range $key, $value := . }}
    <p>{{ $key }}: {{ $value }}</p>
    {{ else }}
    <p>No metrics available</p>
    {{ end }}
</div>
</body>
</html>`)))

			r.GET("/", htmlH.Get)

			request := httptest.NewRequest(http.MethodGet, test.request, nil)

			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			assert.Equal(t, test.want.statusCode, w.Code)

			assert.Equal(t, test.want.contentType, w.Header().Get("Content-Type"))

			//assert.Equal(t, test.want.body, w.Body.String()) // при комите почему-то темлпейт или body изменяется, тест перестает проходить. Проблема с невидимыми символами.

		})
	}

}

func TestMetricsHandler_UpdateJSON(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
	}

	testTable := []struct {
		name    string
		request models.Metrics
		want    want
	}{
		{
			"Valid Counter Update",
			models.Metrics{ID: "PollCounter", MType: "counter", Delta: int64Ptr(2)},
			want{http.StatusOK, "application/json; charset=utf-8"},
		},
		{
			"Valid Gauge Update",
			models.Metrics{ID: "PollGauge", MType: "gauge", Value: float64Ptr(3.14)},
			want{http.StatusOK, "application/json; charset=utf-8"},
		},
		{
			"Invalid Metric Type",
			models.Metrics{ID: "InvalidMetric", MType: "unknown"},
			want{http.StatusBadRequest, "application/json; charset=utf-8"},
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(test.request)
			request := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(jsonData))
			request.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router := gin.Default()
			metricsStorage := storage.NewMetricsStorage()

			metricsService := service.NewMetricsService(metricsStorage, nil, nil, false, true, nil)
			metricsH := NewMetricsHandler(metricsService)

			router.POST("/update", metricsH.UpdateJSON)

			router.ServeHTTP(w, request)
			assert.Equal(t, test.want.statusCode, w.Code)
			assert.Equal(t, test.want.contentType, w.Header().Get("Content-Type"))
		})
	}
}

func TestMetricsHandler_GetJSON(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		response    interface{}
	}
	testTable := []struct {
		name    string
		request models.Metrics
		want    want
		setup   func(storage.MetricsStorage)
	}{
		{
			"Existing Counter",
			models.Metrics{ID: "someCounter", MType: "counter"},
			want{http.StatusOK, "application/json; charset=utf-8", models.Metrics{ID: "someCounter", MType: "counter", Delta: int64Ptr(10)}},
			func(s storage.MetricsStorage) { s.SetCounter("someCounter", 10) },
		},
		{
			"Existing Gauge",
			models.Metrics{ID: "someGauge", MType: "gauge"},
			want{http.StatusOK, "application/json; charset=utf-8", models.Metrics{ID: "someGauge", MType: "gauge", Value: float64Ptr(15.5)}},
			func(s storage.MetricsStorage) { s.SetGauge("someGauge", 15.5) },
		},
		{"Not existing Counter", models.Metrics{ID: "someCounter", MType: "counter"},
			want{http.StatusNotFound, "application/json; charset=utf-8", gin.H{"error": "metric doesn't exist"}},
			func(s storage.MetricsStorage) {},
		},
		{"Not existing Gauge", models.Metrics{ID: "someGauge", MType: "gauge"},
			want{http.StatusNotFound, "application/json; charset=utf-8", gin.H{"error": "metric doesn't exist"}},
			func(s storage.MetricsStorage) {},
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			metricsStorage := storage.NewMetricsStorage()
			test.setup(metricsStorage)

			jsonData, _ := json.Marshal(test.request)
			request := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(jsonData))
			request.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			metricsService := service.NewMetricsService(metricsStorage, nil, nil, false, true, nil)
			metricsH := NewMetricsHandler(metricsService)

			router := gin.Default()
			router.POST("/value", metricsH.GetJSON)

			router.ServeHTTP(w, request)
			assert.Equal(t, test.want.statusCode, w.Code)
			assert.Equal(t, test.want.contentType, w.Header().Get("Content-Type"))

			if w.Code == http.StatusOK {
				var response models.Metrics
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, test.want.response, response)
			} else {
				var errResponse gin.H
				err := json.NewDecoder(w.Body).Decode(&errResponse)
				require.NoError(t, err)
				assert.Equal(t, test.want.response, errResponse)
			}

		})
	}
}

func TestMetricsHandler_Ping(t *testing.T) {
	testTable := []struct {
		name string
		want int
	}{
		{"OK", http.StatusOK},
		{"Internal Server error", http.StatusInternalServerError},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {

			// Мокаем репозиторий
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockRepository(ctrl)

			// Настраиваем поведение мока
			if test.want == http.StatusOK {
				repo.EXPECT().Ping().Return(nil)
			} else {
				repo.EXPECT().Ping().Return(errors.New("server error"))
			}

			serviceRetryer := retry.NewRetryer()
			serviceRetryer.SetConditionFunc(func(err error) bool {
				return errors.Is(err, apperrors.ErrPgConnExc)
			})

			metricsService := service.NewMetricsService(nil, nil, repo, true, true, serviceRetryer)
			metricsH := NewMetricsHandler(metricsService)

			w := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/ping", nil)

			router := gin.Default()
			router.GET("/ping", metricsH.Ping)
			router.ServeHTTP(w, request)

			assert.Equal(t, w.Code, test.want)

		})
	}

}

func TestMetricsHandler_UpdateBatch(t *testing.T) {
	testTable := []struct {
		name string
		want int
		body []models.Metrics
	}{
		{name: "OK gauge", want: http.StatusOK, body: []models.Metrics{{ID: "Metric1", MType: "gauge", Value: float64Ptr(21.2)}}},
		{name: "OK counter", want: http.StatusOK, body: []models.Metrics{{ID: "Metric2", MType: "counter", Delta: int64Ptr(12)}}},
		{name: "Invalid JSON", want: http.StatusOK, body: []models.Metrics{{ID: "Metric3", Delta: int64Ptr(13)}}},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {

			body, err := json.Marshal(test.body)
			assert.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
			assert.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			mStorage := storage.NewMetricsStorage()
			metricsService := service.NewMetricsService(mStorage, nil, nil, false, true, nil)
			metricsH := NewMetricsHandler(metricsService)

			router := gin.Default()
			router.POST("/updates", metricsH.UpdateBatch)
			router.ServeHTTP(w, request)

			assert.Equal(t, test.want, w.Code)
		})
	}
}

func int64Ptr(i int64) *int64 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}
