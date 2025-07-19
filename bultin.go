package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"
)

// RequestID generates and adds a unique request ID to the context.
// The request ID is useful for tracing requests across multiple services
// and correlating log entries.
//
// Example:
//
//	chain := middleware.NewChain(
//		middleware.RequestID(),
//		middleware.Observability(logger),
//		businessLogicMiddleware,
//	)
func RequestID() MiddlewareFunc {
	return RequestIDWithGenerator(generateRequestID)
}

// RequestIDWithGenerator creates a request ID middleware with a custom ID generator.
// This allows you to use your own request ID generation logic.
//
// Example:
//
//	customGenerator := func() string {
//		return fmt.Sprintf("req-%d", time.Now().UnixNano())
//	}
//	middleware := middleware.RequestIDWithGenerator(customGenerator)
func RequestIDWithGenerator(generator func() string) MiddlewareFunc {
	return func(ctx context.Context, input any) (context.Context, any, error) {
		// Check if request ID already exists
		if _, ok := GetRequestID(ctx); !ok {
			return ctx, input, nil
		}

		requestID := generator()
		ctx = SetRequestID(ctx, requestID)

		return ctx, input, nil
	}
}

// Timeout wraps the middleware execution with a timeout context.
// If the downstream middleware takes longer than the specified duration,
// the context is cancelled and an error is returned.
//
// Example:
//
//	middleware := middleware.Timeout(30 * time.Second)
func Timeout(duration time.Duration) MiddlewareFunc {
	return func(ctx context.Context, input any) (context.Context, any, error) {
		timeoutCtx, cancel := context.WithTimeout(ctx, duration)
		defer cancel()

		// Add timeout information to metadata
		timeoutCtx = AddMetadata(timeoutCtx, "timeout", duration.String())

		return timeoutCtx, input, nil
	}
}

// Recovery provides panic recovery for middleware chains.
// If any downstream middleware panics, this middleware catches the panic,
// logs it, and returns an error instead of crashing the application.
//
// Example:
//
//	chain := middleware.NewChain(
//		middleware.Recovery(logger),
//		middleware.Observability(logger),
//		riskyBusinessLogicMiddleware,
//	)
func Recovery(logger *slog.Logger) MiddlewareFunc {
	if logger == nil {
		logger = slog.Default()
	}

	return func(ctx context.Context, input any) (context.Context, any, error) {
		defer func() {
			if r := recover(); r != nil {
				requestID, _ := GetRequestID(ctx)
				chainName, _ := GetChainName(ctx)
				stack := string(debug.Stack())

				logAttrs := []slog.Attr{
					slog.Any("panic", r),
					slog.String("stack", stack),
				}

				if requestID != "" {
					logAttrs = append(logAttrs, slog.String("request_id", requestID))
				}

				if chainName != "" {
					logAttrs = append(logAttrs, slog.String("chain_name", chainName))
				}

				logger.LogAttrs(ctx, slog.LevelError, "Panic recovered in middleware", logAttrs...)
			}
		}()

		return ctx, input, nil
	}
}

// Validation creates a middleware that validates input data using a provided
// validation function. This is useful for ensuring data integrity before
// processing requests.
//
// Example:
//
//	validator := func(input any) error {
//		if input == nil {
//			return errors.New("input cannot be nil")
//		}
//		return nil
//	}
//	middleware := middleware.Validation(validator)
func Validation(validator func(any) error) MiddlewareFunc {
	return func(ctx context.Context, input any) (context.Context, any, error) {
		if err := validator(input); err != nil {
			return ctx, nil, fmt.Errorf("validation failed: %w", err)
		}

		// Add validation success to metadata
		ctx = AddMetadata(ctx, "validated", true)

		return ctx, input, nil
	}
}

// RateLimit creates a simple rate limiting middleware using a token bucket approach.
// This is a basic implementation - for production use cases, consider using
// external rate limiting solutions like Redis-based rate limiters.
//
// Example:
//
//	// Allow 100 requests per second
//	middleware := middleware.RateLimit(100, time.Second)
type tokenBucket struct {
	tokens     int
	capacity   int
	refillRate time.Duration
	lastRefill time.Time
}

func RateLimit(requestsPerDuration int, duration time.Duration) MiddlewareFunc {
	bucket := &tokenBucket{
		tokens:     requestsPerDuration,
		capacity:   requestsPerDuration,
		refillRate: duration,
		lastRefill: time.Now(),
	}

	return func(ctx context.Context, input any) (context.Context, any, error) {
		now := time.Now()

		// Refill tokens based on elapsed time
		if now.Sub(bucket.lastRefill) >= bucket.refillRate {
			bucket.tokens = bucket.capacity
			bucket.lastRefill = now
		}

		// Check if we have tokens available
		if bucket.tokens <= 0 {
			return ctx, nil, fmt.Errorf("rate limit exceeded")
		}

		// Consume a token
		bucket.tokens--

		// Add rate limit info to metadata
		ctx = AddMetadata(ctx, "rate_limit_remaining", bucket.tokens)

		return ctx, input, nil
	}
}

// Conditional creates a middleware that only executes if a condition is met.
// This is useful for implementing feature flags or conditional processing.
//
// Example:
//
//	// Only run expensive middleware for premium users
//	condition := func(ctx context.Context, input any) bool {
//		metadata := middleware.GetMetadata(ctx)
//		userType, exists := metadata["user_type"]
//		return exists && userType == "premium"
//	}
//	middleware := middleware.Conditional(condition, expensiveMiddleware)
func Conditional(condition func(context.Context, any) bool, middleware MiddlewareFunc) MiddlewareFunc {
	return func(ctx context.Context, input any) (context.Context, any, error) {
		if !condition(ctx, input) {
			// Condition not met, skip middleware
			ctx = AddMetadata(ctx, "conditional_skipped", true)
			return ctx, input, nil
		}

		// Condition met, execute middleware
		return middleware(ctx, input)
	}
}

// generateRequestID generates a random request ID
func generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
