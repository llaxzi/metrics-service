// Package middleware содержит интерфейс middleware и его реализацию.
package middleware

import (
	"crypto/rsa"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// IMiddleware определяет интерфейс для middleware сервиса.
type IMiddleware interface {
	WithLogging() gin.HandlerFunc
	InitializeZap(level string) error
	WithGzip() gin.HandlerFunc
	WithDecryptRSA() gin.HandlerFunc
}

// Middleware реализует интерфейс  IMiddleware.
type Middleware struct {
	Log           *zap.Logger // Log Синглтон.
	hashKey       []byte
	cryptoKeyPath string
	privateKey    *rsa.PrivateKey
}

// NewMiddleware создает новый экземпляр IMiddleware
func NewMiddleware(hashKey []byte, cryptoKeyPath string) (IMiddleware, error) {
	m := &Middleware{Log: zap.NewNop(), hashKey: hashKey, cryptoKeyPath: cryptoKeyPath}
	return m, m.loadPrivateKey()
}
