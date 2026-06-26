package domain

import "time"

// Status - статус брони.
type Status string

// Варианты статуса брони.
const (
	StatusConfirmed Status = "confirmed"
	StatusCancelled Status = "cancelled"
)

// Table - столик в ресторане.
type Table struct {
	ID     string
	Number int
	Seats  int
}

// Table - бронирование столика на интервал времени [From, To].
type Booking struct {
	ID        string
	TableID   string
	GuestName string
	Guests    int
	From      time.Time
	To        time.Time
	Status    Status
}

// Intersect - проверяет пересечение двух броней столика
func (b Booking) Intersect(other Booking) bool {
	return b.From.Before(other.To) && other.From.Before(b.To)
}
