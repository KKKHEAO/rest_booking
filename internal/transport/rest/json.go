package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/KKKHEAO/rest_booking/internal/domain"
)

// decodeJSON читает тело запроса в dst, запрещая неизвестные поля.
func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

// writeJSON пишет статус и JSON-тело ответа.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

// writeError маппит доменную ошибку в HTTP-код и пишет JSON с описанием.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrValidation):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrTableTaken):
		writeJSON(w, http.StatusConflict, errorResponse{Error: err.Error()})
	default:
		// неизвестную ошибку клиенту НЕ показываем — только логируем
		slog.ErrorContext(r.Context(), "unhandled error", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}
