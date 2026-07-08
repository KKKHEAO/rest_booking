package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// Доменные ошибки платежного клиента. Используем только в этом пакете.
var (
	// ErrDeclined - платеж был отклонен (тогда retry не используем).
	ErrDeclined = errors.New("payment declined")
	// ErrUnavailable - платеж не удалось провести после всех попыток. (Платежка недоступна).
	ErrUnavailable = errors.New("payment service unavailable")
)

// errTransient — внутренний маркер на retry.
var errTransient = errors.New("transient")

// Config - настройки клиента.
type Config struct {
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
}

// Client - клиент платежного сервиса.
type Client struct {
	http *http.Client
	cfg  Config
}

// NewClient - создает клиент.
func NewClient(cfg Config) *Client {
	return &Client{
		http: &http.Client{Timeout: cfg.Timeout},
		cfg:  cfg,
	}
}

// ChargeRequest — запрос на списание депозита.
type ChargeRequest struct {
	BookingID string `json:"booking_id"`
	Amount    int    `json:"amount"` // в копейках/центах
	Currency  string `json:"currency"`
}

// ChargeResult — ответ платёжки.
type ChargeResult struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
}

// Charge списывает депозит с ретраями и backoff.
func (c *Client) Charge(ctx context.Context, req ChargeRequest) (ChargeResult, error) {
	var lastErr error

	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			// backoff между попытками, прерываемый по ctx
			if err := backoff(ctx, attempt); err != nil {
				return ChargeResult{}, err
			}
		}

		res, err := c.doCharge(ctx, req)
		if err == nil {
			return res, nil
		}
		lastErr = err

		// ретраим только транзиентные ошибки; бизнес-отказ — сразу наружу
		if !errors.Is(err, errTransient) {
			return ChargeResult{}, err
		}
	}

	return ChargeResult{}, fmt.Errorf("%w: after %d attempts: %w",
		ErrUnavailable, c.cfg.MaxRetries+1, lastErr)
}

// doCharge — одна HTTP-попытка.
func (c *Client) doCharge(ctx context.Context, req ChargeRequest) (ChargeResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return ChargeResult{}, fmt.Errorf("marshal: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.cfg.BaseURL+"/charges", bytes.NewReader(body))
	if err != nil {
		return ChargeResult{}, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	// Идемпотентность: один и тот же ключ => платёжка не спишет дважды при ретрае.
	httpReq.Header.Set("Idempotency-Key", req.BookingID)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		// сетевые ошибки/таймаут — транзиентные, можно ретраить
		return ChargeResult{}, fmt.Errorf("%w: do request: %w", errTransient, err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode == http.StatusOK:
		var out ChargeResult
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return ChargeResult{}, fmt.Errorf("decode: %w", err)
		}
		return out, nil

	case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
		// перегрузка/серверная ошибка — транзиентно
		return ChargeResult{}, fmt.Errorf("%w: status %d", errTransient, resp.StatusCode)

	default:
		// 4xx (кроме 429) — бизнес-отказ, ретраить нельзя
		return ChargeResult{}, fmt.Errorf("%w: status %d", ErrDeclined, resp.StatusCode)
	}
}

// backoff ждёт экспоненциально растущую паузу с джиттером, прерываясь по ctx.
func backoff(ctx context.Context, attempt int) error {
	base := time.Duration(1<<uint(attempt-1)) * 100 * time.Millisecond // 100ms, 200ms, 400ms...
	jitter := time.Duration(rand.Int63n(int64(50 * time.Millisecond)))
	wait := base + jitter

	select {
	case <-time.After(wait):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
