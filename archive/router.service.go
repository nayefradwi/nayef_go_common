package common

// func NewRouter() *chi.Mux {
// 	r := chi.NewRouter()
// 	r.Use(middleware.Logger)
// 	// TODO: find a way to modify cors options from main.go
// 	r.Use(cors.Handler(cors.Options{
// 		// TODO change when in production
// 		AllowedOrigins: []string{"https://*", "http://*"},
// 		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
// 		AllowedHeaders: []string{
// 			"Accept",
// 			"Authorization",
// 			"Content-Type",
// 			"X-CSRF-Token",
// 			"X-Warehouse-Id",
// 		},
// 		ExposedHeaders: []string{"Link"},
// 	}))
// 	r.Use(Recover)
// 	r.Use(JsonResponseMiddleware)
// 	r.Use(SetLanguageMiddleware)
// 	r.Use(SetPaginatedDataMiddleware)
// 	r.Get("/health-check", healthCheck)
// 	return r
// }

// func healthCheck(w http.ResponseWriter, r *http.Request) {
// 	ok, _ := json.Marshal(map[string]interface{}{"status": "ok"})
// 	w.Header().Set("Content-Type", "application/json")
// 	w.Write(ok)
// }

// func GetIntURLParam(r *http.Request, key string) int {
// 	val := chi.URLParam(r, key)
// 	intVal, err := strconv.Atoi(val)
// 	if err != nil {
// 		GetLogger().Warn("Failed to parse int url param", zap.String("error", err.Error()))
// 	}
// 	return intVal
// }

// func GetBoolQueryParam(r *http.Request, key string) bool {
// 	val := r.URL.Query().Get(key)
// 	boolVal, err := strconv.ParseBool(val)
// 	if err != nil {
// 		GetLogger().Warn("Failed to parse bool url param", zap.String("error", err.Error()))
// 	}
// 	return boolVal
// }
