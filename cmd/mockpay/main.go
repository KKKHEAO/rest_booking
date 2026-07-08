// Command mockpay — фейковый платёжный сервис для ручного тестирования задания 4.
// Поведение управляется переменными окружения:
//
//	MOCK_ADDR       адрес прослушивания (по умолчанию :9090)
//	MOCK_MODE       ok | slow | flaky | decline | down
//	MOCK_LATENCY    задержка перед ответом, напр. 3s (для проверки таймаутов)
//	MOCK_FAIL_FIRST сколько первых попыток по ключу отдать 503 (режим flaky)
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

func main() {
	addr := envOr("MOCK_ADDR", ":9090")
	mode := envOr("MOCK_MODE", "ok")
	latency, _ := time.ParseDuration(envOr("MOCK_LATENCY", "0"))
	failFirst, _ := strconv.Atoi(envOr("MOCK_FAIL_FIRST", "2"))

	var mu sync.Mutex
	attempts := make(map[string]int) // Idempotency-Key -> число попыток

	http.HandleFunc("/charges", func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")

		if latency > 0 {
			time.Sleep(latency)
		}

		mu.Lock()
		attempts[key]++
		n := attempts[key]
		mu.Unlock()

		log.Printf("charge key=%s attempt=%d mode=%s", key, n, mode)

		switch mode {
		case "down":
			w.WriteHeader(http.StatusServiceUnavailable) // 503 — транзиентно, клиент ретраит
			return
		case "decline":
			w.WriteHeader(http.StatusPaymentRequired) // 402 (4xx) — бизнес-отказ, без ретрая
			return
		case "flaky":
			if n <= failFirst {
				w.WriteHeader(http.StatusServiceUnavailable) // падаем первые N попыток
				return
			}
			// далее проваливаемся в успешный ответ
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"payment_id": "pay_" + key,
			"status":     "captured",
		})
	})

	log.Printf("mock payment listening on %s (mode=%s latency=%s failFirst=%d)",
		addr, mode, latency, failFirst)
	log.Fatal(http.ListenAndServe(addr, nil)) //nolint:gosec // локальный мок, таймауты не нужны
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
