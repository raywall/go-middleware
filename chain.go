package middleware

import (
	"context"
	"fmt"
)

// Define context keys to avoid collisions
type chainNameKey struct{}
type middlewareIndexKey struct{}

// Context keys as variables
var (
	ChainNameKey       = chainNameKey{}
	MiddlewareIndexKey = middlewareIndexKey{}
)

// MiddlewareFunc defines a middleware function that processes context and data.
// It receives a context and input data, and returns a potentially modified context,
// output data, and an error. If an error is returned, the middleware chain
// execution stops immediately.
//
// Example:
//
//	func LoggingMiddleware(logger *slog.Logger) MiddlewareFunc {
//		return func(ctx context.Context, input any) (context.Context, any, error) {
//			logger.Info("Processing request", slog.Any("input", input))
//			return ctx, input, nil
//		}
//	}
type MiddlewareFunc func(ctx context.Context, input any) (context.Context, any, error)

// Chain represents a sequence of MiddlewareFuncs that are executed in order.
// It provides a fluent interface for building and executing middleware pipelines.
//
// Example:
//
//	chain := NewChain(
//		middleware.Observability(logger),
//		middleware.RequestID(),
//		businessLogicMiddleware,
//	)
type Chain struct {
	middlewares []MiddlewareFunc
	name        string // Optional name for debugging/logging
}

// NewChain creates a new middleware Chain with the given middleware functions.
// The middleware functions will be executed in the order they are provided.
//
// Example:
//
//	chain := NewChain(
//		authMiddleware,
//		validationMiddleware,
//		businessLogicMiddleware,
//	)
func NewChain(middlewares ...MiddlewareFunc) *Chain {
	return &Chain{
		middlewares: append([]MiddlewareFunc{}, middlewares...),
	}
}

// NewNamedChain creates a new middleware Chain with a name for debugging purposes.
// The name is useful for logging and tracing to identify different chains.
//
// Example:
//
//	chain := NewNamedChain("user-service-chain",
//		authMiddleware,
//		validationMiddleware,
//	)
func NewNamedChain(name string, middlewares ...MiddlewareFunc) *Chain {
	return &Chain{
		middlewares: append([]MiddlewareFunc{}, middlewares...),
		name:        name,
	}
}

// Append adds one or more middleware functions to the end of the chain.
// This method creates a new chain and does not modify the original chain,
// ensuring immutability and thread safety.
//
// Example:
//
//	baseChain := NewChain(authMiddleware)
//	extendedChain := baseChain.Append(validationMiddleware, businessLogicMiddleware)
func (c *Chain) Append(middlewares ...MiddlewareFunc) *Chain {
	newMiddlewares := make([]MiddlewareFunc, len(c.middlewares)+len(middlewares))
	copy(newMiddlewares, c.middlewares)
	copy(newMiddlewares[len(c.middlewares):], middlewares)

	return &Chain{
		middlewares: newMiddlewares,
		name:        c.name,
	}
}

// Prepend adds one or more middleware functions to the beginning of the chain.
// This method creates a new chain and does not modify the original chain.
//
// Example:
//
//	baseChain := NewChain(businessLogicMiddleware)
//	extendedChain := baseChain.Prepend(authMiddleware, validationMiddleware)
func (c *Chain) Prepend(middlewares ...MiddlewareFunc) *Chain {
	newMiddlewares := make([]MiddlewareFunc, len(middlewares)+len(c.middlewares))
	copy(newMiddlewares, middlewares)
	copy(newMiddlewares[len(middlewares):], c.middlewares)

	return &Chain{
		middlewares: newMiddlewares,
		name:        c.name,
	}
}

// Then executes the middleware chain sequentially, passing the context and data
// through each middleware function. If any middleware returns an error, execution
// stops immediately and the error is returned along with the current context.
//
// The input data flows through each middleware and can be transformed at each step.
// The final output is the result of the last middleware in the chain.
//
// Example:
//
//	ctx := context.Background()
//	ctx, result, err := chain.Then(ctx, requestData)
//	if err != nil {
//		log.Printf("Chain execution failed: %v", err)
//		return
//	}
func (c *Chain) Then(ctx context.Context, input any) (context.Context, any, error) {
	if len(c.middlewares) == 0 {
		return ctx, input, nil
	}

	var err error
	var output any = input
	currentCtx := ctx

	// Add chain metadata to context if chain has a name
	if c.name != "" {
		currentCtx = context.WithValue(currentCtx, ChainNameKey, c.name)
	}

	for i, mw := range c.middlewares {
		// Add current middleware index to context for debugging
		currentCtx = context.WithValue(currentCtx, MiddlewareIndexKey, i)

		currentCtx, output, err = mw(currentCtx, output)
		if err != nil {
			// Wrap error with additional context information
			return currentCtx, nil, fmt.Errorf("middleware %d failed: %w", i, err)
		}
	}

	return currentCtx, output, nil
}

// GetChainName retrieves the chain name from the context.
// It returns the chain name string and a boolean indicating whether the chain name was found.
//
// Example:
//
//	chainName, ok := GetChainName(ctx)
//	if ok {
//	    log.Printf("Executing chain: %s", chainName)
//	}
func GetChainName(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(ChainNameKey).(string)
	return name, ok
}

// Len returns the number of middleware functions in the chain.
func (c *Chain) Len() int {
	return len(c.middlewares)
}

// Name returns the name of the chain, if set.
func (c *Chain) Name() string {
	return c.name
}

// Clone creates a deep copy of the chain, allowing safe modification
// without affecting the original chain.
func (c *Chain) Clone() *Chain {
	middlewares := make([]MiddlewareFunc, len(c.middlewares))
	copy(middlewares, c.middlewares)

	return &Chain{
		middlewares: middlewares,
		name:        c.name,
	}
}
