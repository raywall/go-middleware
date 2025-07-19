# Middleware Package

This Go package provides a reusable, generic, and composable middleware chain system.

## Features

- Chainable middleware functions
- Generic input/output handling via `any`
- Context propagation using `context.Context`
- Middleware composition for observability, validation, auth, etc.

## Example Usage

```go
import (
    "context"
    "fmt"
    "log/slog"
    "os"

    "yourmodule/middleware"
)

type Payload struct {
    UserID string
    Action string
}

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    payload := Payload{UserID: "abc123", Action: "login"}

    chain := middleware.NewChain(
        Observability(logger),
        BusinessLogic(logger),
    )

    ctx := context.Background()
    ctx, result, err := chain.Then(ctx, payload)
    if err != nil {
        logger.Error("error", err)
        return
    }

    logger.Info("result", result)
}
```
