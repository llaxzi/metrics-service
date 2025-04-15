package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func BenchmarkWithLogging(b *testing.B) {
	gin.SetMode(gin.TestMode)

	keyPath := "test_private.pem"
	err := generateTestPrivateKeyPKCS8(keyPath)
	assert.NoError(b, err)
	defer os.Remove(keyPath)

	m, err := NewMiddleware([]byte(""), keyPath)
	assert.NoError(b, err)

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
