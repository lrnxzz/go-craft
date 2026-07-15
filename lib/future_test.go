package lib_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lrnxzz/go-craft/lib"
)

func TestFutureWaitReturnsCompletedValue(t *testing.T) {
	future := lib.NewFuture[int]()
	future.Complete(7, nil)

	value, err := future.Wait(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if value != 7 {
		t.Errorf("value = %d, want 7", value)
	}
}

func TestFutureCompletesOnlyOnce(t *testing.T) {
	future := lib.NewFuture[string]()
	future.Complete("first", nil)
	future.Complete("second", errors.New("late"))

	value, err := future.Wait(context.Background())
	if err != nil || value != "first" {
		t.Errorf("value = %q, err = %v, want first and nil", value, err)
	}
}

func TestFutureWaitHonorsContext(t *testing.T) {
	future := lib.NewFuture[int]()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := future.Wait(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want deadline exceeded", err)
	}
}

func TestFailedFutureIsImmediatelySettled(t *testing.T) {
	sentinel := errors.New("nope")
	future := lib.FailedFuture[int](sentinel)

	select {
	case <-future.Done():
	default:
		t.Fatal("failed future should be settled on creation")
	}

	_, err := future.Wait(context.Background())
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, want the creation error", err)
	}
}
