package common

// import (
// 	"context"
// 	"io"
// 	"net/http"

// 	"go.uber.org/zap"
// )

// var acceptedLang = map[string]bool{DefaultLang: true, "ar": true}

// const DefaultLang = "en"

// type languageKey struct{}

// func SetLanguageMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		lang := getLanguageParam(r)
// 		ctx := r.Context()
// 		ctx = context.WithValue(ctx, languageKey{}, lang)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

// func getLanguageParam(r *http.Request) string {
// 	lang := r.URL.Query().Get("lang")
// 	if lang == "" {
// 		lang = DefaultLang
// 	}
// 	return lang
// }

// func GetLanguageParam(ctx context.Context) string {
// 	lang := ctx.Value(languageKey{})
// 	if lang == nil || lang == "" {
// 		return DefaultLang
// 	}
// 	return lang.(string)
// }

//	func GetTranslatedBody[T any](w http.ResponseWriter, body io.ReadCloser, onSuccess SuccessCallback[Translation[T]]) {
//		ParseBody(w, body, func(translation Translation[T]) {
//			if acceptedLang[translation.LanguageCode] {
//				onSuccess(translation)
//			} else {
//				GetLogger().Warn(
//					"language code is not supported",
//					zap.String("languageCode", translation.LanguageCode),
//				)
//				WriteResponseFromError(w, NewValidationError("languageCode", ErrorDetails{
//					Message: "language code is not supported",
//					Field:   "",
//				}))
//			}
//		})
//	}
