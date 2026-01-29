package responses

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

func JSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

func Success(w http.ResponseWriter, data interface{}, message string) error {
	response := Response{
		Success: true,
		Data:    data,
		Message: message,
	}
	return JSON(w, http.StatusOK, response)
}

func Error(w http.ResponseWriter, statusCode int, err error, code string) error {
	response := ErrorResponse{
		Success: false,
		Error:   err.Error(),
		Code:    code,
	}
	return JSON(w, statusCode, response)
}

func BadRequest(w http.ResponseWriter, err error) error {
	return Error(w, http.StatusBadRequest, err, "BAD_REQUEST")
}

func NotFound(w http.ResponseWriter, err error) error {
	return Error(w, http.StatusNotFound, err, "NOT_FOUND")
}

func InternalError(w http.ResponseWriter, err error) error {
	return Error(w, http.StatusInternalServerError, err, "INTERNAL_ERROR")
}

func Unauthorized(w http.ResponseWriter, err error) error {
	return Error(w, http.StatusUnauthorized, err, "UNAUTHORIZED")
}

func MethodNotAllowed(w http.ResponseWriter) error {
	err := &apiError{message: "Method not allowed"}
	return Error(w, http.StatusMethodNotAllowed, err, "METHOD_NOT_ALLOWED")
}

type apiError struct {
	message string
}

func (e *apiError) Error() string {
	return e.message
}
