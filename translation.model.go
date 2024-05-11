package common

type Translation[T any] struct {
	LanguageCode string `json:"languageCode"`
	Data         T      `json:"data"`
}
