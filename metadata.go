package middleware

import (
	"context"
)

// Define a custom key type to avoid collisions
type metadataKey string

const (
	// Define your metadata keys
	userKey    metadataKey = "user"
	sessionKey metadataKey = "session"
	requestKey metadataKey = "request_id"
)

// AddMetadata adds a key-value pair to the context as metadata.
// It returns a new context that contains the metadata.
//
// Example:
//
//	ctx = AddMetadata(ctx, "validated", true)
//	ctx = AddMetadata(ctx, "user_id", "12345")
func AddMetadata(ctx context.Context, key string, value interface{}) context.Context {
	return context.WithValue(ctx, key, value)
}

// GetMetadata retrieves metadata by key from the context.
// It returns the value and a boolean indicating whether the key was found.
//
// Example:
//
//	value, ok := GetMetadata(ctx, "validated")
//	if ok {
//	    fmt.Printf("Found value: %v\n", value)
//	}
func GetMetadata(ctx context.Context, key string) (interface{}, bool) {
	value := ctx.Value(key)
	return value, value != nil
}

// GetMetadataString retrieves a string value by key from the context metadata.
// It returns the string value and a boolean indicating whether the key was found
// and the value is actually a string type.
//
// Example:
//
//	userID, ok := GetMetadataString(ctx, "user_id")
//	if ok {
//	    fmt.Printf("User ID: %s\n", userID)
//	}
func GetMetadataString(ctx context.Context, key string) (string, bool) {
	value := ctx.Value(key)
	str, ok := value.(string)
	return str, ok
}

// GetMetadataBool retrieves a boolean value by key from the context metadata.
// It returns the boolean value and a boolean indicating whether the key was found
// and the value is actually a bool type.
//
// Example:
//
//	validated, ok := GetMetadataBool(ctx, "validated")
//	if ok && validated {
//	    fmt.Println("Request is validated")
//	}
func GetMetadataBool(ctx context.Context, key string) (bool, bool) {
	value := ctx.Value(key)
	b, ok := value.(bool)
	return b, ok
}

// SetUserID sets the user ID in the context using a type-safe approach.
// It demonstrates the recommended pattern for setting specific metadata types.
//
// Example:
//
//	ctx = SetUserID(ctx, "user123")
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userKey, userID)
}

// GetUserID retrieves the user ID from the context in a type-safe manner.
// It returns the user ID string and a boolean indicating whether the key was found
// and the value is actually a string type.
//
// Example:
//
//	userID, ok := GetUserID(ctx)
//	if ok {
//	    fmt.Printf("Current user: %s\n", userID)
//	}
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userKey).(string)
	return userID, ok
}

// SetRequestID sets the request ID in the context using a type-safe approach.
// This is useful for request tracing and logging purposes.
//
// Example:
//
//	ctx = SetRequestID(ctx, "req_abc123")
func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestKey, requestID)
}

// GetRequestID retrieves the request ID from the context in a type-safe manner.
// It returns the request ID string and a boolean indicating whether the key was found
// and the value is actually a string type.
//
// Example:
//
//	requestID, ok := GetRequestID(ctx)
//	if ok {
//	    log.Printf("Processing request: %s", requestID)
//	}
func GetRequestID(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestKey).(string)
	return requestID, ok
}
