package psqlpool

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func resetPoolState() {
	poolMu.Lock()
	PsqlPool = make(map[string]*Pool)
	poolMu.Unlock()
	connector = nil
}

// TestConcurrentAccess exercises Get/Add/Replace/Remove from many goroutines;
// run with -race to prove map access is synchronized.
func TestConcurrentAccess(t *testing.T) {
	resetPoolState()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("project-%d", i%10)

		wg.Add(4)
		go func() {
			defer wg.Done()
			Add(key, &Pool{})
		}()
		go func() {
			defer wg.Done()
			_, _ = Get(key)
		}()
		go func() {
			defer wg.Done()
			Replace(key, &Pool{})
		}()
		go func() {
			defer wg.Done()
			Remove(key)
		}()
	}
	wg.Wait()
}

func TestGetMissWithoutConnector(t *testing.T) {
	resetPoolState()

	_, err := Get("missing")
	if err == nil || !strings.Contains(err.Error(), "connection not found") {
		t.Fatalf("expected 'connection not found' error, got %v", err)
	}
}

// TestConnectorSingleflight proves concurrent misses on the same id collapse
// into one connector call and all callers get the same pool.
func TestConnectorSingleflight(t *testing.T) {
	resetPoolState()

	var calls int32
	release := make(chan struct{})
	SetConnector(func(projectId string) (*Pool, error) {
		atomic.AddInt32(&calls, 1)
		<-release
		return &Pool{}, nil
	})
	defer resetPoolState()

	const workers = 20
	var wg sync.WaitGroup
	results := make([]*Pool, workers)
	started := make(chan struct{}, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			started <- struct{}{}
			p, err := Get("tenant-a")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			results[i] = p
		}(i)
	}

	for i := 0; i < workers; i++ {
		<-started
	}
	close(release)
	wg.Wait()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected exactly 1 connector call, got %d", got)
	}
	for i := 1; i < workers; i++ {
		if results[i] != results[0] {
			t.Fatalf("worker %d got a different pool instance", i)
		}
	}
	if _, err := Get("tenant-a"); err != nil {
		t.Fatalf("pool should be registered after self-heal, got %v", err)
	}
}

// TestConnectorFailureKeepsMarker: callers across services match on the
// "connection not found" substring, so a failed self-heal must preserve it,
// and a later attempt must retry (no negative caching).
func TestConnectorFailureKeepsMarker(t *testing.T) {
	resetPoolState()

	var calls int32
	SetConnector(func(projectId string) (*Pool, error) {
		atomic.AddInt32(&calls, 1)
		return nil, errors.New("vault: no creds")
	})
	defer resetPoolState()

	for i := 0; i < 2; i++ {
		_, err := Get("tenant-broken")
		if err == nil || !strings.Contains(err.Error(), "connection not found") {
			t.Fatalf("expected wrapped 'connection not found' error, got %v", err)
		}
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected connector retried per miss, got %d calls", got)
	}
}
