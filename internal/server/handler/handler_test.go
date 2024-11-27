package handler

import (
	"github.com/stretchr/testify/assert"
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
		{"OK", "/update/counter/PollCounter/2", want{http.StatusOK, "text/plain"}},
		{"Wrong url #1", "/update/counter/PollCounter", want{http.StatusNotFound, "text/plain; charset=utf-8"}},
		{"Wrong url #2", "/update/counter/PollCounter/invalidType", want{http.StatusBadRequest, "text/plain; charset=utf-8"}},
	}

	for _, test := range testTable {
		request := httptest.NewRequest(http.MethodPost, test.request, nil)
		request.Header.Set("Content-Type", "text/plain")

		w := httptest.NewRecorder()

		metricsStorage := storage.NewMetricsStorage()
		counterH := NewCounterHandler(metricsStorage)

		counterH.Update(w, request)

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
		{"OK", "/update/gauge/PollCounter/2", want{http.StatusOK, "text/plain"}},
		{"Wrong url #1", "/update/gauge/PollCounter", want{http.StatusNotFound, "text/plain; charset=utf-8"}},
		{"Wrong url #2", "/update/gauge/PollCounter/invalidType", want{http.StatusBadRequest, "text/plain; charset=utf-8"}},
	}

	for _, test := range testTable {
		request := httptest.NewRequest(http.MethodPost, test.request, nil)
		request.Header.Set("Content-Type", "text/plain")

		w := httptest.NewRecorder()

		metricsStorage := storage.NewMetricsStorage()
		gaugeH := NewGaugeHandler(metricsStorage)

		gaugeH.Update(w, request)

		assert.Equal(t, test.want.statusCode, w.Code)

		assert.Equal(t, test.want.contentType, w.Header().Get("Content-Type"))

	}

}
