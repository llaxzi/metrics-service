package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func generateHMAC(data, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func BenchmarkWithHMAC(b *testing.B) {
	gin.SetMode(gin.TestMode)
	m := &middleware{hashKey: []byte("secret-key")}
	r := gin.New()
	r.Use(m.WithHMAC())
	r.POST("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, World!")
	})

	body := []byte("test data")
	hash := generateHMAC(body, m.hashKey)

	b.Run("WithValidHMAC", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
			req.Header.Set("HashSHA256", hash)

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
		}
	})

	b.Run("WithInvalidHMAC", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
			req.Header.Set("HashSHA256", "invalidhash")

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
		}
	})
}
