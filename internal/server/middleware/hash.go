package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"io"
	apperrors "metrics-service/internal/server/errors"
	"net/http"
)

type responseWriter struct {
	gin.ResponseWriter
	body []byte
}

func (rw *responseWriter) Write(p []byte) (n int, err error) {
	rw.body = append(rw.body, p...)
	return rw.ResponseWriter.Write(p)
}

func (m *middleware) WithHMAC() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if len(m.hashKey) < 1 {
			ctx.Next()
			return
		}

		hashHeader := ctx.GetHeader("HashSHA256")
		if hashHeader == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": apperrors.ErrHashHeaderMissing})
			ctx.Abort()
			return
		}

		body, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": apperrors.ErrServer})
			ctx.Abort()
			return
		}

		// Возвращаем тело запроса обратно в поток
		ctx.Request.Body = io.NopCloser(bytes.NewReader(body))

		hashBody, err := m.generateHash(body)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": apperrors.ErrServer})
			ctx.Abort()
			return
		}

		if hashBody != hashHeader {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": apperrors.ErrHashHeaderInvalid})
			ctx.Abort()
			return
		}

		// Перехватываем ответ
		writer := &responseWriter{ResponseWriter: ctx.Writer}
		ctx.Writer = writer

		ctx.Next()

		hashResponse, err := m.generateHash(writer.body)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": apperrors.ErrServer})
			return
		}

		ctx.Header("HashSHA256", hashResponse)
	}
}

func (m *middleware) generateHash(src []byte) (string, error) {
	h := hmac.New(sha256.New, m.hashKey)
	_, err := h.Write(src)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
