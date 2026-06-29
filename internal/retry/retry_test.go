package retry

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsRetriable(t *testing.T) {
	if IsRetriable(nil) {
		t.Fatal("nil не должен быть retriable")
	}
	// Обычная (не pg) ошибка считается retriable.
	if !IsRetriable(errors.New("boom")) {
		t.Fatal("обычная ошибка должна быть retriable")
	}
	// Ошибка соединения PostgreSQL (класс 08) — retriable.
	connErr := &pgconn.PgError{Code: "08006"}
	if !IsRetriable(connErr) {
		t.Fatal("ошибка соединения должна быть retriable")
	}
	// Прочая pg-ошибка — не retriable.
	otherErr := &pgconn.PgError{Code: "23505"} // unique_violation
	if IsRetriable(otherErr) {
		t.Fatal("unique_violation не должна быть retriable")
	}
}

func TestDo_SuccessFirstTry(t *testing.T) {
	calls := 0
	err := Do(context.Background(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("ожидался nil, получено %v", err)
	}
	if calls != 1 {
		t.Fatalf("функция должна быть вызвана один раз, вызвана %d", calls)
	}
}

func TestDo_NonRetriableStops(t *testing.T) {
	calls := 0
	sentinel := &pgconn.PgError{Code: "23505"}
	err := Do(context.Background(), func() error {
		calls++
		return sentinel
	})
	if !errors.Is(err, error(sentinel)) {
		t.Fatalf("ожидалась исходная ошибка, получено %v", err)
	}
	if calls != 1 {
		t.Fatalf("не-retriable ошибка не должна повторяться, вызвана %d раз", calls)
	}
}

func TestDo_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := Do(ctx, func() error {
		return errors.New("transient")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ожидался context.Canceled, получено %v", err)
	}
}
