package domain_test

import (
	"testing"
	"time"

	"github.com/KKKHEAO/rest_booking/internal/domain"
)

func TestIntersect(t *testing.T) {
	slot := func(fromHour, toHour int) domain.Booking {
		day := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
		return domain.Booking{
			From: day.Add(time.Duration(fromHour) * time.Hour),
			To:   day.Add(time.Duration(toHour) * time.Hour),
		}
	}

	tests := []struct {
		name string
		a    domain.Booking
		b    domain.Booking
		want bool
	}{
		{"полное пересечение", slot(18, 20), slot(18, 20), true},
		{"частичное пересечение", slot(18, 20), slot(19, 21), true},
		{"стык встык не пересекается", slot(18, 20), slot(20, 22), false},
		{"совсем не пересекается b>a", slot(18, 20), slot(21, 23), false},
		{"совсем не пересекается b<a", slot(21, 23), slot(18, 20), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.Intersect(tt.b)
			if got != tt.want {
				t.Errorf("Intersect() = %v, want %v", got, tt.want)
			}
		})
	}
}
