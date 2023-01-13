package response

import (
	"encoding/json"
	"net/http"

	"github.com/nayefradwi/nayef_go_common/baseError"
)

func WriteErrorResponse(w http.ResponseWriter, err *baseError.BaseError) {
	response := err.GenerateResponse()
	w.WriteHeader(err.Status)
	w.Write(response)
}

func WriteEmptyCreatedResponse(w http.ResponseWriter, m string) {
	w.WriteHeader(http.StatusCreated)
	body := make(map[string]interface{})
	body["status"] = http.StatusCreated
	body["message"] = m
	json.NewEncoder(w).Encode(body)
}

func WriteEmptySuccessResponse(w http.ResponseWriter, m string) {
	body := make(map[string]interface{})
	body["status"] = http.StatusOK
	body["message"] = m
	json.NewEncoder(w).Encode(body)
}
