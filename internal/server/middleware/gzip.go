package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// newGzipWriter создает новый gzipWriter, оборачивающий gin.ResponseWriter.
func newGzipWriter(w gin.ResponseWriter) *gzipWriter {
	return &gzipWriter{w, gzip.NewWriter(w)}
}

// gzipWriter оборачивает gin.ResponseWriter и выполняет сжатие данных перед отправкой клиенту.
type gzipWriter struct {
	gin.ResponseWriter
	gzWriter *gzip.Writer
}

// Write выполняет сжатие данных перед их записью в ResponseWriter.
func (w *gzipWriter) Write(b []byte) (int, error) {
	return w.gzWriter.Write(b)
}

// newGzipReader создает новый gzipReader для разжатия тела запроса.
func newGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &gzipReader{r, zr}, nil
}

// gzipReader оборачивает io.ReadCloser и выполняет разжатие входных данных.
type gzipReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// Read считывает разжатые данные.
func (g *gzipReader) Read(p []byte) (n int, err error) {
	return g.zr.Read(p)
}

// Close закрывает gzipReader и исходный поток данных.
func (g *gzipReader) Close() error {
	if err := g.zr.Close(); err != nil {
		return err
	}
	return g.r.Close()
}

// WithGzip добавляет middleware для сжатия и декомпрессии данных с помощью gzip.
//
// Если клиент поддерживает gzip (заголовок "Accept-Encoding" содержит "gzip"),
// то ответы сервера будут сжиматься перед отправкой.
//
// Если запрос от клиента сжат с использованием gzip (заголовок "Content-Encoding" содержит "gzip"),
// то middleware разожмет тело запроса перед его обработкой.
func (m *Middleware) WithGzip() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		acceptGzip := strings.Contains(ctx.Request.Header.Get("Accept-Encoding"), "gzip")

		if acceptGzip {
			gzW := newGzipWriter(ctx.Writer)
			defer gzW.gzWriter.Close()

			ctx.Writer = gzW
			ctx.Header("Content-Encoding", "gzip")
		}

		sendsGzip := strings.Contains(ctx.Request.Header.Get("Content-Encoding"), "gzip")

		if sendsGzip {
			cr, err := newGzipReader(ctx.Request.Body)
			if err != nil {
				ctx.Writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			ctx.Request.Body = cr
			defer cr.Close()
		}

		ctx.Next()
	}
}
