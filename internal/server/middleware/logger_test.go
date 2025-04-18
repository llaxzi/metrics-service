package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func BenchmarkWithLogging(b *testing.B) {
	gin.SetMode(gin.TestMode)

	m := NewMiddleware([]byte("some-secret"))
	_ = m.InitializeZap("debug")

	loggingMiddleware := m.WithLogging()

	router := gin.New()
	router.Use(loggingMiddleware)
	router.GET("/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "pong")
	})

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
