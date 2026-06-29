package scraper

import (
	"net/http"
	"testing"
	"time"
)

func TestAntiBanTransport_RoundTrip(t *testing.T) {
	cfg := DefaultAntiBanConfig()
	cfg.MinDelay = 10 * time.Millisecond
	cfg.MaxDelay = 50 * time.Millisecond
	cfg.MaxRetries = 1
	cfg.RetryDelay = 100 * time.Millisecond

	transport := NewAntiBanTransport(cfg)

	// Use example.com which is stable and won't trigger anti-bot
	req, err := http.NewRequest("GET", "https://example.com", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify headers were set
	ua := req.Header.Get("User-Agent")
	if ua == "" {
		t.Error("User-Agent should be set by transport")
	}

	referer := req.Header.Get("Referer")
	if referer == "" {
		t.Error("Referer should be set by transport")
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	lim := NewRateLimiter(1.0, 5) // 1 token/sec, max burst 5

	start := time.Now()
	for i := 0; i < 5; i++ {
		lim.Wait()
	}
	elapsed := time.Since(start)

	// 5 tokens should consume the burst allowance instantly
	if elapsed > 500*time.Millisecond {
		t.Errorf("burst wait took too long: %v", elapsed)
	}

	// Next token should take ~1 second
	start = time.Now()
	lim.Wait()
	elapsed = time.Since(start)

	if elapsed < 800*time.Millisecond || elapsed > 2000*time.Millisecond {
		t.Errorf("expected ~1s wait, got %v", elapsed)
	}
}

func TestRateLimiter_TryWait(t *testing.T) {
	lim := NewRateLimiter(0.5, 2) // 2 tokens max

	// Should succeed twice (burst)
	if !lim.TryWait() {
		t.Error("TryWait should succeed with available tokens")
	}
	if !lim.TryWait() {
		t.Error("TryWait should succeed with available tokens")
	}

	// Should fail (no more tokens)
	if lim.TryWait() {
		t.Error("TryWait should fail when no tokens available")
	}
}

func TestCloneRequest(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	req.Header.Set("X-Custom", "test-value")

	cloned := cloneRequest(req)

	// Modify cloned headers should not affect original
	cloned.Header.Set("X-Custom", "modified")

	if req.Header.Get("X-Custom") != "test-value" {
		t.Error("original request header should not be modified")
	}
}
