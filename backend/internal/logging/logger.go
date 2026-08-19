// Package logging centraliza a configuração do logger estruturado da
// aplicação (log/slog, saída em JSON) e a propagação desse logger — já
// com o correlation_id da requisição atual — via context.Context, para
// que services e repositories não precisem recebê-lo como parâmetro.
package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"os"
	"strings"

	"ecommerce/internal/config"
)

// ServiceName identifica, em cada registro de log, qual serviço o gerou.
// Hoje só existe um serviço (o monólito); o campo já fica pronto para
// quando outros serviços passarem a compor o mesmo fluxo distribuído.
const ServiceName = "ecommerce-api"

type contextKey int

const loggerContextKey contextKey = iota

var sensitiveAttrKeys = map[string]struct{}{
	"password":      {},
	"password_hash": {},
	"passwordhash":  {},
	"token":         {},
	"access_token":  {},
	"authorization": {},
}

// New monta o logger estruturado da aplicação: saída em JSON, nível vindo
// de cfg.LogLevel, e todo registro já identificado com o serviço e o
// ambiente que o geraram.
func New(cfg *config.Config) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:       parseLevel(cfg.LogLevel),
		ReplaceAttr: redactSensitive,
	})

	return slog.New(handler).With(
		slog.String("service", ServiceName),
		slog.String("environment", cfg.Environment),
	)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// redactSensitive impede que campos com nomes tipicamente sensíveis (senha,
// token etc.) cheguem ao log, mesmo que algum chamador os inclua por engano.
func redactSensitive(groups []string, a slog.Attr) slog.Attr {
	if _, sensitive := sensitiveAttrKeys[strings.ToLower(a.Key)]; sensitive {
		a.Value = slog.StringValue("[REDACTED]")
	}

	return a
}

// ContextWithLogger devolve um contexto carregando logger — normalmente já
// com o correlation_id da requisição atual — para que services e
// repositories recuperem o mesmo logger sem precisar recebê-lo como
// parâmetro explícito.
func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, logger)
}

// FromContext recupera o logger carregado pelo contexto. Se nenhum foi
// definido (por exemplo, em um teste que não passa pelo middleware de
// requisição), devolve o logger global como fallback — nunca nil.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerContextKey).(*slog.Logger); ok && logger != nil {
		return logger
	}

	return slog.Default()
}

// NewCorrelationID gera um identificador aleatório para correlacionar todos
// os logs de uma mesma operação dentro deste serviço e, futuramente, entre
// serviços diferentes do mesmo fluxo distribuído.
func NewCorrelationID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)

	return hex.EncodeToString(buf)
}
