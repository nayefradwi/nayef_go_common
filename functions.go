package common

import (
	"context"
	"encoding/json"
	"time"
)

func HasAllValues[T any, P string | int](values []T, allValues []T, getKey func(value T) P) bool {
	mapValues := make(map[P]bool)
	for _, value := range allValues {
		key := getKey(value)
		mapValues[key] = true
	}
	for _, value := range values {
		key := getKey(value)
		if !mapValues[key] {
			return false
		}
	}
	return true
}

func GetUtcDateOnlyString() string {
	y, m, d := time.Now().UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
}

func GetUtcDateOnlyStringFromTime(t time.Time) string {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
}

func GetValues[K comparable, V any](input map[K]V) []V {
	values := make([]V, 0)
	for _, value := range input {
		values = append(values, value)
	}
	return values
}

func StructToMap(input interface{}) (map[string]interface{}, error) {
	encoded, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = json.Unmarshal(encoded, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetBoolFromContext(ctx context.Context, key any) bool {
	value := ctx.Value(key)
	if value == nil {
		return false
	}
	if val, ok := value.(bool); ok {
		return val
	}
	return false
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

func MergeSlice[T any](slice1 []T, slice2 []T) []T {
	result := make([]T, 0)
	result = append(result, slice1...)
	result = append(result, slice2...)
	return result
}

func MergeMaps[K comparable, V any](map1 map[K]V, map2 map[K]V) map[K]V {
	result := make(map[K]V)
	for key, value := range map1 {
		result[key] = value
	}
	for key, value := range map2 {
		result[key] = value
	}
	return result
}

func Map[F any, T any](input []F, fn func(F) T) []T {
	result := make([]T, len(input))
	for i, value := range input {
		result[i] = fn(value)
	}
	return result
}

func FirstWhere[T any](input []T, fn func(T) bool) *T {
	for _, value := range input {
		if fn(value) {
			return &value
		}
	}
	return nil
}
