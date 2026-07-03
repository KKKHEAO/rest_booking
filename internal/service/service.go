package service

import (
	"context"
	"fmt"
	"time"

	"github.com/KKKHEAO/rest_booking/internal/domain"
	"github.com/KKKHEAO/rest_booking/internal/storage"
)

// Service содержит с бизнес-логикой
type Service struct {
	store storage.Storage
}

// NewService создаёт Service поверх переданного хранилища.
func NewService(store storage.Storage) *Service {
	return &Service{store: store}
}

// CreateBooking валидирует корректность и сохраняет бронь.
func (s *Service) CreateBooking(ctx context.Context, b domain.Booking) (domain.Booking, error) {
	// Валидация бизнес-правил
	if b.GuestName == "" {
		return domain.Booking{}, fmt.Errorf("%w: guest name required", domain.ErrValidation)
	}
	if b.Guests <= 0 {
		return domain.Booking{}, fmt.Errorf("%w: at least one guest required", domain.ErrValidation)
	}
	if !b.From.Before(b.To) {
		return domain.Booking{}, fmt.Errorf("%w: from must be before to", domain.ErrValidation)
	}
	if b.From.Before(time.Now()) {
		return domain.Booking{}, fmt.Errorf("%w: cannot book in the past", domain.ErrValidation)
	}

	// Проверяем вместимость стола.
	table, err := s.store.GetTable(ctx, b.TableID)
	if err != nil {
		return domain.Booking{}, err
	}
	if b.Guests > table.Seats {
		return domain.Booking{}, fmt.Errorf("%w: table %d has only %d seats", domain.ErrValidation, table.Number, table.Seats)
	}

	// Проверка пересечений. Пока так, потом через constraint.
	existing, err := s.store.ListBookings(ctx)
	if err != nil {
		return domain.Booking{}, err
	}
	for _, e := range existing {
		if e.TableID != b.TableID || e.Status != domain.StatusConfirmed {
			continue
		}
		if b.Intersect(e) {
			return domain.Booking{}, domain.ErrTableTaken
		}
	}

	b.Status = domain.StatusConfirmed
	return s.store.CreateBooking(ctx, b)
}

// CreateTable валидирует и создаёт стол.
func (s *Service) CreateTable(ctx context.Context, t domain.Table) (domain.Table, error) {
	if t.Number <= 0 {
		return domain.Table{}, fmt.Errorf("%w: table number must be positive", domain.ErrValidation)
	}
	if t.Seats <= 0 {
		return domain.Table{}, fmt.Errorf("%w: seats must be positive", domain.ErrValidation)
	}

	return s.store.CreateTable(ctx, t)
}

// GetTable возвращает стол по ID.
func (s *Service) GetTable(ctx context.Context, id string) (domain.Table, error) {
	return s.store.GetTable(ctx, id)
}

// ListTables возвращает все столы.
func (s *Service) ListTables(ctx context.Context) ([]domain.Table, error) {
	return s.store.ListTables(ctx)
}

// GetBooking возвращает бронь по ID.
func (s *Service) GetBooking(ctx context.Context, id string) (domain.Booking, error) {
	return s.store.GetBooking(ctx, id)
}

// ListBookings возвращает все брони.
func (s *Service) ListBookings(ctx context.Context) ([]domain.Booking, error) {
	return s.store.ListBookings(ctx)
}

// CancelBooking изменяет статус брони на cancelled.
func (s *Service) CancelBooking(ctx context.Context, id string) error {
	return s.store.UpdateBookingStatus(ctx, id, domain.StatusCancelled)
}
