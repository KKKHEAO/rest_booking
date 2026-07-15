package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/KKKHEAO/rest_booking/internal/config"
	"github.com/KKKHEAO/rest_booking/internal/observability"
	"github.com/KKKHEAO/rest_booking/internal/payment"
	"github.com/KKKHEAO/rest_booking/internal/service"
	"github.com/KKKHEAO/rest_booking/internal/storage/postgres"
	"github.com/KKKHEAO/rest_booking/internal/transport/rest"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	logger := newLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	// Контекст, который отменяется по SIGINT/SIGTERM — сигнал к остановке.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	// Инициализируем трейсер
	shutdownTracer, err := observability.InitTracer(ctx, "rest_booking", cfg.OTLPEndpoint)
	if err != nil {
		return err
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTracer(shutdownCtx); err != nil {
			logger.Error("tracer shutdown", "error", err)
		}
	}()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	repo := postgres.NewRepository(pool)

	payClient := payment.NewClient(payment.Config{
		BaseURL:    cfg.PaymentBaseURL,
		Timeout:    cfg.PaymentTimeout,
		MaxRetries: cfg.PaymentRetries,
	})

	svc := service.NewService(repo, payClient)
	handler := rest.NewRouter(svc, logger)
	// Обернем хэндлер
	handler = otelhttp.NewHandler(handler, "http", otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
		return r.Method + " " + r.URL.Path
	}))

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Сервер слушает в отдельной горутине, ошибку отдаёт в канал.
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server started", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Ждём либо ошибку сервера, либо сигнал остановки.
	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	// Даём 10 секунд на завершение текущих запросов.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	logger.Info("server stopped gracefully")
	return nil
}

// newLogger создаёт JSON-логгер slog с уровнем из конфига.
// TODO: вынести в internal
func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}
