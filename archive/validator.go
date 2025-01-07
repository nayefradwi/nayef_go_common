package common

// func ValidateTime(dateTime time.Time, field string) ErrorDetails {
// 	if dateTime.IsZero() {
// 		return ErrorDetails{
// 			Message: field + " cannot be empty",
// 			Field:   field,
// 		}
// 	}
// 	return ErrorDetails{}
// }

// func ValidateIdPtr(id *int, field string) ErrorDetails {
// 	if id == nil {
// 		return ErrorDetails{
// 			Message: field + " cannot be empty",
// 			Field:   field,
// 		}
// 	}
// 	if *id <= 0 {
// 		return ErrorDetails{
// 			Message: field + " must be greater than 0",
// 			Field:   field,
// 		}
// 	}
// 	return ErrorDetails{}
// }

// func ValidateId(id int, field string) ErrorDetails {
// 	return ValidateIdPtr(&id, field)
// }

// func ValidateNotZero[T float64 | int](amount T, field string) ErrorDetails {
// 	if amount <= 0 {
// 		return ErrorDetails{
// 			Message: field + " must be greater than 0",
// 			Field:   field,
// 		}
// 	}
// 	return ErrorDetails{}
// }

// func ValidateStringLength(value string, field string, min int, max int) ErrorDetails {
// 	if len(value) < min || len(value) > max {
// 		return ErrorDetails{
// 			Message: field + " must be between " + strconv.Itoa(min) + " and " + strconv.Itoa(max),
// 			Field:   field,
// 		}
// 	}
// 	return ErrorDetails{}
// }

// func ValidateAlphaNuemericPtr(value *string, field string) ErrorDetails {
// 	if value == nil {
// 		return ErrorDetails{
// 			Message: field + " cannot be empty",
// 			Field:   field,
// 		}
// 	}
// 	if !isNameValid(*value) {
// 		return ErrorDetails{
// 			Message: "invalid " + field,
// 			Field:   field,
// 		}
// 	}
// 	return ErrorDetails{}
// }

// func ValidateAlphanuemericName(value, field string) ErrorDetails {
// 	if !isNameValid(value) {
// 		return ErrorDetails{
// 			Message: "invalid " + field,
// 			Field:   field,
// 		}
// 	}
// 	return ErrorDetails{}
// }

// func isNameValid(name string) bool {
// 	pattern := "^[\\p{L}0-9\\s.-]+$"
// 	re := regexp.MustCompile(pattern)
// 	return re.MatchString(name) && len(name) >= 3 && len(name) <= 50
// }

// func ValidateUrl(value *string, field string) ErrorDetails {
// 	if value == nil {
// 		return ErrorDetails{}
// 	}
// 	url, err := url.Parse(*value)
// 	if err != nil || url.Host == "" || url.Scheme == "" {
// 		return ErrorDetails{
// 			Message: "invalid image url",
// 			Field:   "image",
// 		}
// 	}
// 	return ErrorDetails{}
// }

// func ValidateSliceSize[T any](data []T, field string, min int, max int) ErrorDetails {
// 	if len(data) < min || len(data) > max {
// 		return ErrorDetails{
// 			Message: field + " must have between " + strconv.Itoa(min) + " and " + strconv.Itoa(max) + " elements",
// 			Field:   field,
// 		}
// 	}
// 	return ErrorDetails{}
// }

// func ValidateAmountPositive[T float64 | int](amount T, field string) ErrorDetails {
// 	if amount <= 0 {
// 		return ErrorDetails{
// 			Message: field + " must be greater than 0",
// 			Field:   field,
// 		}
// 	}
// 	return ErrorDetails{}
// }
// func ValidateAmount[T float64 | int](amount T, field string, min, max T) ErrorDetails {
// 	if amount < min || amount > max {
// 		return ErrorDetails{
// 			Message: field + " must be greater than " + strconv.Itoa(int(min)) + " and less than " + strconv.Itoa(int(max)),
// 			Field:   field,
// 		}
// 	}
// 	return ErrorDetails{}
// }
