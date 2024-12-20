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

func NewMiddleware() Middleware {
	return &middleware{zap.NewNop()}
}
