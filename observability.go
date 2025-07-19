package middleware

import (
	"context"
	"log/slog"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Define a custom key type to avoid collisions
type observabilityKey string

const (
	// Define your observability keuys
	StartTimeKey observabilityKey = "start_time"
)

// ObservabilityConfig configures the observability middleware behavior
type ObservabilityConfig struct {
	// Logger is the structured logger instance
	Logger *slog.Logger

	// SpanName is the name to use for distributed tracing spans
	SpanName string

	// LogInput determines whether to log the input data
	LogInput bool

	// LogOutput determines whether to log the output data
	LogOutput bool

	// LogLevel is the log level to use for middleware logs
	LogLevel slog.Level

	// SkipHealthChecks determines whether to skip logging for health check requests
	SkipHealthChecks bool
}

// DefaultObservabilityConfig returns a default configuration for observability middleware
func DefaultObservabilityConfig() *ObservabilityConfig {
	return &ObservabilityConfig{
		SpanName:         "middleware.request",
		LogInput:         true,
		LogOutput:        false,
		LogLevel:         slog.LevelInfo,
		SkipHealthChecks: true,
	}
}

// Observability creates a middleware function that provides distributed tracing
// and structured logging capabilities. It integrates with DataDog APM for
// distributed tracing and uses structured logging for observability.
//
// The middleware automatically:
//   - Creates distributed tracing spans
//   - Logs request processing with structured data
//   - Tracks request duration
//   - Adds observability metadata to context
//
// Example:
//
//	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
//	middleware := middleware.Observability(logger)
func Observability(logger *slog.Logger) MiddlewareFunc {
	config := DefaultObservabilityConfig()
	config.Logger = logger
	return ObservabilityWithConfig(config)
}

// ObservabilityWithConfig creates an observability middleware with custom configuration.
// This allows fine-grained control over logging and tracing behavior.
//
// Example:
//
//	config := &middleware.ObservabilityConfig{
//		Logger:           logger,
//		SpanName:         "user-service.request",
//		LogInput:         true,
//		LogOutput:        true,
//		LogLevel:         slog.LevelDebug,
//		SkipHealthChecks: false,
//	}
//	middleware := middleware.ObservabilityWithConfig(config)
func ObservabilityWithConfig(config *ObservabilityConfig) MiddlewareFunc {
	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	if config.SpanName == "" {
		config.SpanName = "middleware.request"
	}

	return func(ctx context.Context, input any) (context.Context, any, error) {
		startTime := time.Now()

		// Create distributed tracing span
		span := tracer.StartSpan(config.SpanName)
		defer span.Finish()

		// Add span to context for downstream middleware
		ctx = tracer.ContextWithSpan(ctx, span)

		// Store start time in context
		ctx = context.WithValue(ctx, StartTimeKey, startTime)

		// Get request ID if available
		requestID, _ := GetRequestID(ctx)
		if requestID != "" {
			span.SetTag("request.id", requestID)
		}

		// Get chain name if available
		chainName, _ := GetChainName(ctx)
		if chainName != "" {
			span.SetTag("chain.name", chainName)
		}

		// Log structured input data
		logAttrs := []slog.Attr{
			slog.Time("timestamp", startTime),
		}

		if requestID != "" {
			logAttrs = append(logAttrs, slog.String("request_id", requestID))
		}

		if chainName != "" {
			logAttrs = append(logAttrs, slog.String("chain_name", chainName))
		}

		if config.LogInput {
			logAttrs = append(logAttrs, slog.Any("input", input))
			// Set trace tag for input (convert to string for safety)
			span.SetTag("input.type", getTypeName(input))
		}

		config.Logger.LogAttrs(ctx, config.LogLevel, "Request started", logAttrs...)

		// Mark context as observed
		ctx = AddMetadata(ctx, "observed", true)
		ctx = AddMetadata(ctx, "start_time", startTime)

		// Continue with next middleware (input is passed through unchanged)
		return ctx, input, nil
	}
}

// ObservabilityComplete creates a middleware that logs both the start and completion
// of request processing. This is useful as the last middleware in a chain to
// capture the final result and duration.
//
// Example:
//
//	chain := middleware.NewChain(
//		middleware.Observability(logger),
//		businessLogicMiddleware,
//		middleware.ObservabilityComplete(logger), // Log completion
//	)
func ObservabilityComplete(logger *slog.Logger) MiddlewareFunc {
	return func(ctx context.Context, input any) (context.Context, any, error) {
		startTimeValue := ctx.Value(StartTimeKey)
		var duration time.Duration

		if startTime, ok := startTimeValue.(time.Time); ok {
			duration = time.Since(startTime)
		}

		requestID, _ := GetRequestID(ctx)
		chainName, _ := GetChainName(ctx)

		logAttrs := []slog.Attr{
			slog.Duration("duration", duration),
			slog.Time("completed_at", time.Now()),
		}

		if requestID != "" {
			logAttrs = append(logAttrs, slog.String("request_id", requestID))
		}

		if chainName != "" {
			logAttrs = append(logAttrs, slog.String("chain_name", chainName))
		}

		logAttrs = append(logAttrs, slog.Any("output", input))

		logger.LogAttrs(ctx, slog.LevelInfo, "Request completed", logAttrs...)

		// Add span tags for completion
		if span, ok := tracer.SpanFromContext(ctx); ok {
			span.SetTag("duration.ms", float64(duration.Nanoseconds())/1e6)
			span.SetTag("output.type", getTypeName(input))
		}

		return ctx, input, nil
	}
}

// getTypeName safely extracts the type name from any value
func getTypeName(v any) string {
	if v == nil {
		return "nil"
	}

	switch v.(type) {
	case string:
		return "string"
	case int, int8, int16, int32, int64:
		return "int"
	case uint, uint8, uint16, uint32, uint64:
		return "uint"
	case float32, float64:
		return "float"
	case bool:
		return "bool"
	case []byte:
		return "[]byte"
	case map[string]any:
		return "map[string]any"
	default:
		return "unknown"
	}
}
