package rest

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/KKKHEAO/rest_booking/internal/domain"
)

type createTableRequest struct {
	Number int `json:"number"`
	Seats  int `json:"seats"`
}

type tableResponse struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	Seats  int    `json:"seats"`
}

func toTableResponse(t domain.Table) tableResponse {
	return tableResponse{
		ID:     t.ID,
		Number: t.Number,
		Seats:  t.Seats,
	}
}

func (a *API) createTable(w http.ResponseWriter, r *http.Request) {
	var req createTableRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, fmt.Errorf("%w: %w", domain.ErrValidation, err))
		return
	}

	t := domain.Table{
		Number: req.Number,
		Seats:  req.Seats,
	}

	created, err := a.svc.CreateTable(r.Context(), t)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, toTableResponse(created))
}

func (a *API) getTable(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	t, err := a.svc.GetTable(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toTableResponse(t))
}

func (a *API) listTables(w http.ResponseWriter, r *http.Request) {
	tables, err := a.svc.ListTables(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}

	resp := make([]tableResponse, 0, len(tables))
	for _, t := range tables {
		resp = append(resp, toTableResponse(t))
	}

	writeJSON(w, http.StatusOK, resp)
}
