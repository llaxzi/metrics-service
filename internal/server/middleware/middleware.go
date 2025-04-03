// Package middleware содержит интерфейс middleware и его реализацию.
package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// IMiddleware определяет интерфейс для middleware сервиса.
type IMiddleware interface {
	WithLogging() gin.HandlerFunc
	InitializeZap(level string) error
	WithGzip() gin.HandlerFunc
}

// Middleware реализует интерфейс  IMiddleware.
type Middleware struct {
	Log     *zap.Logger // Log Синглтон.
	hashKey []byte
}

// NewMiddleware создает новый экземпляр IMiddleware
func NewMiddleware(hashKey []byte) IMiddleware {
	return &Middleware{zap.NewNop(), hashKey}
}
