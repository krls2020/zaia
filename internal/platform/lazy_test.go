package platform

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestLazyClient_InitOnFirstCall(t *testing.T) {
	var callCount int32
	client := NewLazyClient(func() (string, string, error) {
		atomic.AddInt32(&callCount, 1)
		return "tok", "api.zerops.io", nil
	})

	// First call should trigger init.
	_, err := client.GetUserInfo(t.Context())
	// We expect an API error since we're hitting a real endpoint with fake token.
	// The point is that init() was called.
	_ = err
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected 1 init call, got %d", callCount)
	}

	// Second call should reuse cached client.
	_, _ = client.GetUserInfo(t.Context())
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected still 1 init call after second use, got %d", callCount)
	}
}

func TestLazyClient_InitError(t *testing.T) {
	client := NewLazyClient(func() (string, string, error) {
		return "", "", errors.New("no credentials")
	})

	_, err := client.ListServices(t.Context(), "proj-1")
	if err == nil {
		t.Fatal("expected error from resolver")
	}
	if err.Error() != "no credentials" {
		t.Errorf("error = %q, want 'no credentials'", err.Error())
	}
}

func TestLazyClient_ConcurrentInit(t *testing.T) {
	var callCount int32
	client := NewLazyClient(func() (string, string, error) {
		atomic.AddInt32(&callCount, 1)
		return "tok", "api.zerops.io", nil
	})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = client.GetUserInfo(t.Context())
		}()
	}
	wg.Wait()

	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected exactly 1 init call with concurrent access, got %d", callCount)
	}
}
