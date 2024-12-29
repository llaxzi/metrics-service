package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type middleware struct {
	Log *zap.Logger // Log Синглтон.
}

// InitializeZap инициализирует синглтон логера с необходимым уровнем логирования.
func (m *middleware) InitializeZap(level string) error {

	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zLogger, err := cfg.Build()
	if err != nil {
		return err
	}

	m.Log = zLogger
	return nil
}

func (m *middleware) WithLogging() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		start := time.Now()

		// Сначала передаем gin.Context дальше, после выполнения всей цепочки возвращаемся и логируем
		ctx.Next()

		duration := time.Since(start)

		m.Log.Info("got incoming HTTP request",
			zap.String("method", ctx.Request.Method),
			zap.String("path", ctx.Request.URL.Path),
			zap.String("duration", strconv.FormatFloat(duration.Seconds(), 'f', 3, 64)),
		)

		m.Log.Info("sending HTTP response",
			zap.String("status", strconv.Itoa(ctx.Writer.Status())),
			zap.String("length", strconv.Itoa(ctx.Writer.Size())),
		)

	}
}
