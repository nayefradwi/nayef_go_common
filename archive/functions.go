package common

import (
	"context"
)

func GetBoolFromContext(ctx context.Context, key any) bool {
	value := ctx.Value(key)
	if value == nil {
		return false
	}
	val, _ := value.(bool)
	return val
}

func SetBoolToContext(ctx context.Context, key any, value string) context.Context {
	if value == "true" {
		return context.WithValue(ctx, key, true)
	}
	return context.WithValue(ctx, key, false)
}

func GetIntFromContext(ctx context.Context, key any) int {
	value := ctx.Value(key)
	if value == nil {
		return 0
	}
	if val, ok := value.(int); ok {
		return val
	}
	return 0
}
