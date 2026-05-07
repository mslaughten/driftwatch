package ratelimit

import (
	"testing"
	"time"
)

func TestAllow_ConsumesTokens(t *testing.T) {
	l := New(3, 1)
	for i := 0; i < 3; i++ {
		if !l.Allow() {
			t.Fatalf("expected Allow() to return true on call %d", i+1)
		}
	}
	if l.Allow() {
		t.Fatal("expected Allow() to return false when tokens exhausted")
	}
}

func TestAllow_RefillsOverTime(t *testing.T) {
	now := time.Now()
	l := New(2, 2) // 2 tokens/sec
	l.clock = func() time.Time { return now }

	// drain
	l.Allow()
	l.Allow()
	if l.Allow() {
		t.Fatal("should be empty")
	}

	// advance clock by 1 second — should add 2 tokens
	now = now.Add(time.Second)
	if !l.Allow() {
		t.Fatal("expected token after refill")
	}
	if !l.Allow() {
		t.Fatal("expected second token after refill")
	}
	if l.Allow() {
		t.Fatal("should be empty again")
	}
}

func TestAllow_CapAtMax(t *testing.T) {
	now := time.Now()
	l := New(3, 10) // fast refill
	l.clock = func() time.Time { return now }

	// drain all
	l.Allow()
	l.Allow()
	l.Allow()

	// advance far into the future — tokens should cap at max
	now = now.Add(10 * time.Second)
	l.Allow() // trigger refill

	if l.Tokens() > 3 {
		t.Fatalf("tokens %v exceeded max 3", l.Tokens())
	}
}

func TestNew_FullBucket(t *testing.T) {
	l := New(5, 1)
	if l.Tokens() != 5 {
		t.Fatalf("expected 5 initial tokens, got %v", l.Tokens())
	}
}
