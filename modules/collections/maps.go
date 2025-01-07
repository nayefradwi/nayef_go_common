package collections

import "encoding/json"

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

func MergeMapsWithFunc[K comparable, V any](map1 map[K]V, map2 map[K]V, mergeFunc func(value1, value2 V) V) map[K]V {
	result := make(map[K]V)

	for key, value := range map1 {
		result[key] = value
	}

	for key, value := range map2 {
		if _, ok := result[key]; ok {
			result[key] = mergeFunc(result[key], value)
		} else {
			result[key] = value
		}
	}
	return result
}
