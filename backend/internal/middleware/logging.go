package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ecommerce/internal/logging"
)

// CorrelationIDHeader é o cabeçalho usado para receber (de um chamador ou de
// outro serviço) e devolver o identificador de correlação da requisição.
const CorrelationIDHeader = "X-Correlation-ID"

// RequestLogger gera (ou reaproveita) um correlation_id por requisição,
// disponibiliza um logger com esse campo no contexto da requisição — para
// handlers e services logarem suas próprias operações correlacionadas — e
// emite um registro estruturado com o resultado da requisição ao final.
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := strings.TrimSpace(c.GetHeader(CorrelationIDHeader))
		if correlationID == "" {
			correlationID = logging.NewCorrelationID()
		}

		c.Writer.Header().Set(CorrelationIDHeader, correlationID)

		requestLogger := logger.With(slog.String("correlation_id", correlationID))
		c.Request = c.Request.WithContext(logging.ContextWithLogger(c.Request.Context(), requestLogger))

		start := time.Now()
		c.Next()
		duration := time.Since(start)

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		attrs := []any{
			slog.String("operation", "http_request"),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", duration),
			slog.String("client_ip", c.ClientIP()),
		}

		if user, ok := GetAuthenticatedUser(c); ok {
			attrs = append(attrs, slog.String("customer_id", user.ID))
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("error", c.Errors.String()))
		}

		status := c.Writer.Status()

		switch {
		case status >= http.StatusInternalServerError:
			requestLogger.Error("requisição HTTP concluída", attrs...)
		case status >= http.StatusBadRequest:
			requestLogger.Warn("requisição HTTP concluída", attrs...)
		default:
			requestLogger.Info("requisição HTTP concluída", attrs...)
		}
	}
}

// Recovery captura panics durante o processamento da requisição, registra o
// ocorrido em formato estruturado (com stack trace) e responde 500 — sem
// isso, um panic num handler derrubaria o processo inteiro.
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			logging.FromContext(c.Request.Context()).Error(
				"panic recuperado durante o processamento da requisição",
				slog.String("operation", "http_request"),
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
				slog.Any("panic", recovered),
				slog.String("stack", string(debug.Stack())),
			)

			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "erro interno do servidor",
			})
		}()

		c.Next()
	}
}
