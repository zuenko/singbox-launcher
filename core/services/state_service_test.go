package services

import (
	"sync"
	"testing"
	"time"
)

func TestAutoPingAfterConnectDefaultEnabled(t *testing.T) {
	s := NewStateService()
	if !s.IsAutoPingAfterConnectEnabled() {
		t.Error("auto-ping should default to enabled")
	}
	s.SetAutoPingAfterConnectEnabled(false)
	if s.IsAutoPingAfterConnectEnabled() {
		t.Error("auto-ping should be disabled after Set(false)")
	}
}

func TestTemplateDirtyRoundTrip(t *testing.T) {
	s := NewStateService()
	if s.IsTemplateDirty() {
		t.Error("template dirty should default to false")
	}
	s.SetTemplateDirty(true)
	if !s.IsTemplateDirty() {
		t.Error("template dirty should be true after Set(true)")
	}
	s.SetTemplateDirty(false)
	if s.IsTemplateDirty() {
		t.Error("template dirty should be false after Set(false)")
	}
}

func TestRecordUpdateFailureAndSuccess(t *testing.T) {
	s := NewStateService()

	msg, at := s.GetLastUpdateFailure()
	if msg != "" || !at.IsZero() {
		t.Errorf("fresh state should have no failure, got msg=%q at=%v", msg, at)
	}

	s.RecordUpdateFailure("network timeout")
	msg, at = s.GetLastUpdateFailure()
	if msg != "network timeout" {
		t.Errorf("msg = %q, want %q", msg, "network timeout")
	}
	if at.IsZero() {
		t.Error("at should be non-zero after RecordUpdateFailure")
	}

	// Success replaces the failure tombstone.
	time.Sleep(1 * time.Millisecond)
	s.RecordUpdateSuccess()
	msg, at = s.GetLastUpdateFailure()
	if msg != "" || !at.IsZero() {
		t.Errorf("after success, failure should be cleared, got msg=%q at=%v", msg, at)
	}
	s.LastUpdateMutex.RLock()
	succ := s.LastUpdateSucceededAt
	s.LastUpdateMutex.RUnlock()
	if succ.IsZero() {
		t.Error("LastUpdateSucceededAt should be non-zero after success")
	}
}

// Stress the failure/success race surface: if the mutex is wrong, -race will
// bark. Trivial test — the real defence is go test -race on CI.
func TestStateServiceConcurrency(t *testing.T) {
	s := NewStateService()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%3 == 0 {
				s.RecordUpdateFailure("err")
			} else {
				s.RecordUpdateSuccess()
			}
			_, _ = s.GetLastUpdateFailure()
			s.SetAutoPingAfterConnectEnabled(i%2 == 0)
			_ = s.IsAutoPingAfterConnectEnabled()
			s.SetTemplateDirty(i%2 == 0)
			_ = s.IsTemplateDirty()
		}(i)
	}
	wg.Wait()
}
