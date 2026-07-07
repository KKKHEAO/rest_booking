package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/KKKHEAO/rest_booking/internal/domain"
)

// Repository - реализация Storage interface для Postgres.
type Repository struct {
	pool *pgxpool.Pool
}

// CreateBooking сохраняет бронь, генерируя ID при необходимости.
func (r *Repository) CreateBooking(ctx context.Context, b domain.Booking) (domain.Booking, error) {
	if b.ID == "" {
		b.ID = uuid.NewString()
	}

	tx, err := r.pool.Begin(ctx)

	if err != nil {
		return domain.Booking{}, fmt.Errorf("begin tx: %w", err)
	}

	if _, err := tx.Exec(ctx, lockTable, b.TableID); err != nil {
		return domain.Booking{}, fmt.Errorf("lock table: %w", err)
	}

	if _, err := tx.Exec(ctx, createBooking,
		b.ID, b.TableID, b.GuestName, b.Guests, b.From, b.To, b.Status,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ExclusionViolation {
			return domain.Booking{}, domain.ErrTableTaken
		}
		return domain.Booking{}, fmt.Errorf("insert booking: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Booking{}, fmt.Errorf("commit tx: %w", err)
	}

	return b, nil
}

// CreateTable создает столик.
func (r *Repository) CreateTable(ctx context.Context, t domain.Table) (domain.Table, error) {
	if t.ID == "" {
		t.ID = uuid.NewString()
	}

	if _, err := r.pool.Exec(ctx, createTable, t.ID, t.Number, t.Seats); err != nil {
		return domain.Table{}, fmt.Errorf("insert table: %w", err)
	}

	return t, nil
}

// UpdateBookingStatus изменяет статус брони или возвращает ErrNotFound.
func (r *Repository) UpdateBookingStatus(ctx context.Context, id string, status domain.Status) error {
	tag, err := r.pool.Exec(ctx, updateBookingStatus, id, status)
	if err != nil {
		return fmt.Errorf("update booking status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// GetBooking возвращает бронь по ID или domain.ErrNotFound.
func (r *Repository) GetBooking(ctx context.Context, id string) (domain.Booking, error) {
	var b domain.Booking
	err := r.pool.QueryRow(ctx, getBooking, id).Scan(
		&b.ID, &b.TableID, &b.GuestName, &b.Guests, &b.From, &b.To, &b.Status,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Booking{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Booking{}, fmt.Errorf("get booking: %w", err)
	}

	return b, nil
}

// GetTable возвращает стол по ID или domain.ErrNotFound.
func (r *Repository) GetTable(ctx context.Context, id string) (domain.Table, error) {
	var t domain.Table
	err := r.pool.QueryRow(ctx, getTable, id).Scan(&t.ID, &t.Number, &t.Seats)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Table{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Table{}, fmt.Errorf("get table: %w", err)
	}

	return t, nil
}

// ListBookings возвращает брони отсортированные по времени начала.
func (r *Repository) ListBookings(ctx context.Context) ([]domain.Booking, error) {
	rows, err := r.pool.Query(ctx, getListBookings)
	if err != nil {
		return nil, fmt.Errorf("query bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]domain.Booking, 0)

	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(
			&b.ID, &b.TableID, &b.GuestName, &b.Guests, &b.From, &b.To, &b.Status,
		); err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}
		bookings = append(bookings, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bookings: %w", err)
	}

	return bookings, nil
}

// ListTables возвращает столы отсортированные по номеру.
func (r *Repository) ListTables(ctx context.Context) ([]domain.Table, error) {
	rows, err := r.pool.Query(ctx, getListTables)
	if err != nil {
		return nil, fmt.Errorf("query tables: %w", err)
	}
	defer rows.Close()

	tables := make([]domain.Table, 0)

	for rows.Next() {
		var t domain.Table
		if err := rows.Scan(&t.ID, &t.Number, &t.Seats); err != nil {
			return nil, fmt.Errorf("scan table: %w", err)
		}
		tables = append(tables, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tables: %w", err)
	}

	return tables, nil
}

// NewRepository создает репозиторий для Postgres.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}
