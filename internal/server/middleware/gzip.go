package middleware

import (
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
)

func newGzipWriter(w gin.ResponseWriter) *gzipWriter {
	return &gzipWriter{w, gzip.NewWriter(w)}
}

type gzipWriter struct {
	gin.ResponseWriter
	gzWriter *gzip.Writer
}

func (w *gzipWriter) Write(b []byte) (int, error) {
	return w.gzWriter.Write(b)
}

func newGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &gzipReader{r, zr}, nil
}

type gzipReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func (g *gzipReader) Read(p []byte) (n int, err error) {
	return g.zr.Read(p)
}

func (g *gzipReader) Close() error {
	if err := g.zr.Close(); err != nil {
		return err
	}
	return g.r.Close()
}

func (m *middleware) WithGzip() gin.HandlerFunc {
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
			}
			ctx.Request.Body = cr
			defer cr.Close()
		}

		ctx.Next()
	}
}
