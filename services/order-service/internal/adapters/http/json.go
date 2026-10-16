package http

import (
	"encoding/json"
	"net/http"

	"github.com/jcmexdev/flashcart-engine/services/order/internal/apperror"
)

// UnifiedErrorResponse represents a standard unified error response across all endpoints.
type UnifiedErrorResponse struct {
	Code    apperror.ErrorCode `json:"code"`
	Message string             `json:"message"`
	Details map[string]string  `json:"details,omitempty"`
}

// WriteJSON serializes data to JSON and sets the Content-Type header and HTTP status code.
func WriteJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

// WriteUnifiedError serializes application errors following the Unified Application Error pattern.
func WriteUnifiedError(w http.ResponseWriter, err error) {
	appErr, ok := apperror.AsAppError(err)
	if !ok {
		WriteJSON(w, http.StatusInternalServerError, UnifiedErrorResponse{
			Code:    apperror.CodeInternal,
			Message: "an unexpected error occurred",
		})
		return
	}

	statusCode := http.StatusInternalServerError
	switch appErr.Code {
	case apperror.CodeInvalidInput:
		statusCode = http.StatusBadRequest
	case apperror.CodeNotFound:
		statusCode = http.StatusNotFound
	case apperror.CodeConflict:
		statusCode = http.StatusConflict
	case apperror.CodeInternal:
		statusCode = http.StatusInternalServerError
	}

	WriteJSON(w, statusCode, UnifiedErrorResponse{
		Code:    appErr.Code,
		Message: appErr.Message,
		Details: appErr.Details,
	})
}
