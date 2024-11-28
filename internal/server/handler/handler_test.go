package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"html/template"
	"metrics-service/internal/server/storage"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCounterHandler_Update(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
	}
	testTable := []struct {
		name    string
		request string
		want    want
	}{
		{"OK", "/update/counter/PollCounter/2", want{http.StatusOK, "text/plain; charset=utf-8"}},
		{"Wrong url #1", "/update/counter/PollCounter", want{http.StatusNotFound, "text/plain; charset=utf-8"}},
		{"Wrong url #2", "/update/counter/PollCounter/invalidType", want{http.StatusBadRequest, "text/plain; charset=utf-8"}},
	}

	for _, test := range testTable {
		request := httptest.NewRequest(http.MethodPost, test.request, nil)
		request.Header.Set("Content-Type", "text/plain")

		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = request

		metricsStorage := storage.NewMetricsStorage()
		counterH := NewCounterHandler(metricsStorage)

		counterH.Update(ctx)

		assert.Equal(t, test.want.statusCode, w.Code)

		assert.Equal(t, test.want.contentType, w.Header().Get("Content-Type"))

	}

}

func TestGaugeHandler_Update(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
	}
	testTable := []struct {
		name    string
		request string
		want    want
	}{
		{"OK", "/update/gauge/PollCounter/2", want{http.StatusOK, "text/plain; charset=utf-8"}},
		{"Wrong url #1", "/update/gauge/PollCounter", want{http.StatusNotFound, "text/plain; charset=utf-8"}},
		{"Wrong url #2", "/update/gauge/PollCounter/invalidType", want{http.StatusBadRequest, "text/plain; charset=utf-8"}},
	}

	for _, test := range testTable {
		request := httptest.NewRequest(http.MethodPost, test.request, nil)
		request.Header.Set("Content-Type", "text/plain")

		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = request

		metricsStorage := storage.NewMetricsStorage()
		gaugeH := NewGaugeHandler(metricsStorage)

		gaugeH.Update(ctx)

		assert.Equal(t, test.want.statusCode, w.Code)

		assert.Equal(t, test.want.contentType, w.Header().Get("Content-Type"))

	}

}

func TestCounterHandler_Get(t *testing.T) {

	type want struct {
		statusCode  int
		contentType string
		body        string
	}
	testTable := []struct {
		name               string
		request            string
		requestContentType string
		want               want
		storageSet         func(s storage.MetricsStorage)
	}{
		{"OK", "/value/counter/someMetric", "text/plain", want{http.StatusOK, "text/plain", "5"}, func(s storage.MetricsStorage) {
			s.SetCounter("someMetric", 5)
		}},
		{"Invalid content type", "/value/counter/someMetric", "application/json", want{http.StatusUnsupportedMediaType, "text/plain", "unsupported content type"}, func(s storage.MetricsStorage) {
		}},
		{"Not found URL", "/value/counter/someMetric/metric", "text/plain", want{http.StatusNotFound, "text/plain", "Not found"}, func(s storage.MetricsStorage) {
		}},
		{"Not found metric", "/value/counter/someMetric", "text/plain", want{http.StatusNotFound, "text/plain", "metric doesn't exist"}, func(s storage.MetricsStorage) {
		}},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.request, nil)
			request.Header.Set("Content-Type", test.requestContentType)

			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = request

			metricsStorage := storage.NewMetricsStorage()
			counterH := NewCounterHandler(metricsStorage)

			// подготавливаем storage в соответствии с тестом
			test.storageSet(metricsStorage)

			counterH.Get(ctx)

			assert.Equal(t, test.want.statusCode, w.Code)

			assert.Contains(t, w.Header().Get("Content-Type"), test.want.contentType)

			assert.Equal(t, test.want.body, w.Body.String())

		})
	}

}

func TestGaugeHandler_Get(t *testing.T) {

	type want struct {
		statusCode  int
		contentType string
		body        string
	}
	testTable := []struct {
		name               string
		request            string
		requestContentType string
		want               want
		storageSet         func(s storage.MetricsStorage)
	}{
		{"OK", "/value/gauge/someMetric", "text/plain", want{http.StatusOK, "text/plain", "1.343"}, func(s storage.MetricsStorage) {
			s.SetGauge("someMetric", 1.343)
		}},
		{"Invalid content type", "/value/gauge/someMetric", "application/json", want{http.StatusUnsupportedMediaType, "text/plain", "unsupported content type"}, func(s storage.MetricsStorage) {

		}},
		{"Not found URL", "/value/gauge/someMetric/metric", "text/plain", want{http.StatusNotFound, "text/plain", "Not found"}, func(s storage.MetricsStorage) {
		}},
		{"Not found metric", "/value/gauge/someMetric", "text/plain", want{http.StatusNotFound, "text/plain", "metric doesn't exist"}, func(s storage.MetricsStorage) {
		}},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.request, nil)
			request.Header.Set("Content-Type", test.requestContentType)

			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = request

			metricsStorage := storage.NewMetricsStorage()
			gaugeH := NewGaugeHandler(metricsStorage)

			// подготавливаем storage в соответствии с тестом
			test.storageSet(metricsStorage)

			gaugeH.Get(ctx)

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
		name               string
		request            string
		requestContentType string
		want               want
		storageSet         func(s storage.MetricsStorage)
	}{
		{"OK", "/", "text/plain", want{http.StatusOK, "text/html; charset=utf-8", `<!DOCTYPE html>
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
		{"Invalid content type", "/", "application/json", want{http.StatusUnsupportedMediaType, "text/plain; charset=utf-8", "unsupported content type"}, func(s storage.MetricsStorage) {

		}},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {

			metricsStorage := storage.NewMetricsStorage()
			htmlH := NewHTMLHandler(metricsStorage)

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
			request.Header.Set("Content-Type", test.requestContentType)

			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			assert.Equal(t, test.want.statusCode, w.Code)

			assert.Equal(t, test.want.contentType, w.Header().Get("Content-Type"))

			assert.Equal(t, test.want.body, w.Body.String())

		})
	}

}
