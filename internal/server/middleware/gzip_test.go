package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware_WithGzip(t *testing.T) {
	r := gin.Default()
	m := NewMiddleware([]byte(""))

	r.Use(m.WithGzip())

	// Эндпоинт для теста сжатия
	r.GET("/gzip", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello, World!")
	})

	t.Run("Gzip compression", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/gzip", nil)
		assert.NoError(t, err)

		req.Header.Set("Accept-Encoding", "gzip")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

		gr, err := gzip.NewReader(w.Body)
		assert.NoError(t, err)

		data, err := io.ReadAll(gr)
		assert.NoError(t, err)

		assert.Equal(t, "Hello, World!", string(data))

	})

	// Эндпоинт для теста сжатого запроса
	r.POST("/gzip", func(ctx *gin.Context) {
		body, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "failed to read body")
			return
		}
		ctx.String(http.StatusOK, string(body))
	})
	t.Run("Gzip decompression", func(t *testing.T) {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)

		_, err := gz.Write([]byte("Compressed data"))
		assert.NoError(t, err)

		err = gz.Close()
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", "/gzip", &buf)
		assert.NoError(t, err)

		req.Header.Set("Content-Encoding", "gzip")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Compressed data", w.Body.String())

	})

	t.Run("No compression or decompression", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/gzip", nil)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Content-Encoding"))

	})
}

func BenchmarkWithGzip(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	m := &middleware{}
	r := gin.New()
	r.Use(m.WithGzip())
	r.POST("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, World!")
	})

	b.Run("WithoutGzip", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString("test data"))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
		}
	})

	b.Run("WithGzip", func(b *testing.B) {
		var compressedData bytes.Buffer
		gz := gzip.NewWriter(&compressedData)
		_, _ = gz.Write([]byte("test data"))
		gz.Close()

		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest(http.MethodPost, "/test", &compressedData)
			req.Header.Set("Content-Encoding", "gzip")
			req.Header.Set("Accept-Encoding", "gzip")

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
		}
	})
}
