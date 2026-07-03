package rest

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/KKKHEAO/rest_booking/internal/domain"
)

type createBookingRequest struct {
	TableID   string    `json:"table_id"`
	GuestName string    `json:"guest_name"`
	Guests    int       `json:"guests"`
	From      time.Time `json:"from"`
	To        time.Time `json:"to"`
}

type bookingResponse struct {
	ID        string    `json:"id"`
	TableID   string    `json:"table_id"`
	GuestName string    `json:"guest_name"`
	Guests    int       `json:"guests"`
	From      time.Time `json:"from"`
	To        time.Time `json:"to"`
	Status    string    `json:"status"`
}

func toBookingResponse(b domain.Booking) bookingResponse {
	return bookingResponse{
		ID:        b.ID,
		TableID:   b.TableID,
		GuestName: b.GuestName,
		Guests:    b.Guests,
		From:      b.From,
		To:        b.To,
		Status:    string(b.Status),
	}
}

func (a *API) createBooking(w http.ResponseWriter, r *http.Request) {
	var req createBookingRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, fmt.Errorf("%w: %w", domain.ErrValidation, err))
		return
	}

	b := domain.Booking{
		TableID:   req.TableID,
		GuestName: req.GuestName,
		Guests:    req.Guests,
		From:      req.From,
		To:        req.To,
	}

	created, err := a.svc.CreateBooking(r.Context(), b)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, toBookingResponse(created))
}

func (a *API) getBooking(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	b, err := a.svc.GetBooking(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toBookingResponse(b))
}

func (a *API) listBookings(w http.ResponseWriter, r *http.Request) {
	bookings, err := a.svc.ListBookings(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}

	resp := make([]bookingResponse, 0, len(bookings))
	for _, b := range bookings {
		resp = append(resp, toBookingResponse(b))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (a *API) cancelBooking(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := a.svc.CancelBooking(r.Context(), id); err != nil {
		writeError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
