package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/raywall/go-middleware"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type Payload struct {
	UserID string
	Action string
}

func main() {
	// Inicia o tracer do Datadog
	tracer.Start()
	defer tracer.Stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	payload := Payload{
		UserID: "abc123",
		Action: "login",
	}

	// Criação da cadeia de middlewares
	chain := middleware.NewChain(
		middleware.Observability(logger),
		// Aqui você poderia adicionar middleware.Validation(), Auth(), etc.
		businessLogic(logger),
	)

	ctx := context.Background()
	ctx, result, err := chain.Then(ctx, payload)
	if err != nil {
		logger.Error("Erro na execução", slog.String("err", err.Error()))
		return
	}

	logger.Info("Resultado final", slog.Any("output", result))
}

func businessLogic(logger *slog.Logger) middleware.MiddlewareFunc {
	return func(ctx context.Context, input any) (context.Context, any, error) {
		payload, ok := input.(Payload)
		if !ok {
			return ctx, nil, fmt.Errorf("payload inválido")
		}

		logger.Info("Handler executando lógica de negócio",
			slog.String("user_id", payload.UserID),
			slog.String("action", payload.Action),
		)

		// Retorna um resultado de forma genérica
		return ctx, map[string]string{
			"message": "Ação realizada com sucesso!",
			"user_id": payload.UserID,
		}, nil
	}
}
