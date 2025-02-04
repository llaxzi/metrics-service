package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Middleware interface {
	WithLogging() gin.HandlerFunc
	InitializeZap(level string) error
	WithGzip() gin.HandlerFunc
}

type middleware struct {
	Log     *zap.Logger // Log Синглтон.
	hashKey []byte
}

func NewMiddleware(hashKey []byte) Middleware {
	return &middleware{zap.NewNop(), hashKey}
}
