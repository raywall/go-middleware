// Package middleware provides a robust, chainable middleware framework for Go microservices.
//
// This library enables developers to create reusable middleware pipelines that can process
// requests through multiple stages such as observability, authentication, validation,
// rate limiting, and business logic.
//
// # Core Concepts
//
// The middleware package is built around three main concepts:
//   - MiddlewareFunc: A function that processes context and data
//   - Chain: A sequence of middleware functions executed in order
//   - Built-in middlewares: Pre-built middleware for common use cases
//
// # Basic Usage
//
// Creating and using a middleware chain:
//
//	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
//
//	chain := middleware.NewChain(
//		middleware.Observability(logger),
//		middleware.RequestID(),
//		customBusinessLogic,
//	)
//
//	ctx := context.Background()
//	ctx, result, err := chain.Then(ctx, inputData)
//	if err != nil {
//		// Handle error
//	}
//
// # Creating Custom Middleware
//
// Custom middleware functions must follow the MiddlewareFunc signature:
//
//	func CustomMiddleware(config Config) middleware.MiddlewareFunc {
//		return func(ctx context.Context, input any) (context.Context, any, error) {
//			// Pre-processing logic
//
//			// Modify context or input as needed
//			ctx = context.WithValue(ctx, "custom", "value")
//
//			// Post-processing logic
//			return ctx, input, nil
//		}
//	}
//
// # Error Handling
//
// When any middleware in the chain returns an error, the execution stops immediately
// and the error is propagated to the caller:
//
//	func ValidationMiddleware() middleware.MiddlewareFunc {
//		return func(ctx context.Context, input any) (context.Context, any, error) {
//			if input == nil {
//				return ctx, nil, errors.New("input cannot be nil")
//			}
//			return ctx, input, nil
//		}
//	}
//
// # Built-in Middleware
//
// The package includes several pre-built middleware:
//   - Observability: Distributed tracing and structured logging
//   - RequestID: Generates and tracks unique request identifiers
//   - Timeout: Adds timeout control to request processing
//   - Recovery: Panic recovery with graceful error handling
//
// # Context Values
//
// Middleware can store and retrieve values from the context using well-defined keys:
//
//	// Storing values
//	ctx = context.WithValue(ctx, middleware.RequestIDKey, "req-123")
//
//	// Retrieving values
//	requestID := middleware.GetRequestID(ctx)
//
// # Performance Considerations
//
// - Middleware functions are executed sequentially, so keep them lightweight
// - Use context cancellation for long-running operations
// - Consider the order of middleware - place faster operations first
//
// # Thread Safety
//
// All middleware functions should be thread-safe as they may be called concurrently
// across multiple goroutines. The Chain itself is thread-safe for read operations
// but should not be modified after creation.
//
// # Examples
//
// For complete examples and advanced usage patterns, see the examples directory
// and the test files in this package.
package middleware
