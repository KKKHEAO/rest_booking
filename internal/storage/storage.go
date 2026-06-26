package storage

import (
	"context"

	"github.com/KKKHEAO/rest_booking/internal/domain"
)

// Storage — описание методов для хранилища.
type Storage interface {
	// Tables
	CreateTable(ctx context.Context, t domain.Table) (domain.Table, error)
	GetTable(ctx context.Context, id string) (domain.Table, error)
	ListTables(ctx context.Context) ([]domain.Table, error)

	// Bookings
	CreateBooking(ctx context.Context, b domain.Booking) (domain.Booking, error)
	GetBooking(ctx context.Context, id string) (domain.Booking, error)
	ListBookings(ctx context.Context) ([]domain.Booking, error)
	UpdateBookingStatus(ctx context.Context, id string, status domain.Status) error
}
