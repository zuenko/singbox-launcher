package services

import (
	"sync"
	"testing"
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

func TestRecordUpdateSuccess(t *testing.T) {
	s := NewStateService()

	s.LastUpdateMutex.RLock()
	succ := s.LastUpdateSucceededAt
	s.LastUpdateMutex.RUnlock()
	if !succ.IsZero() {
		t.Errorf("fresh state should have zero LastUpdateSucceededAt, got %v", succ)
	}

	s.RecordUpdateSuccess()
	s.LastUpdateMutex.RLock()
	succ = s.LastUpdateSucceededAt
	s.LastUpdateMutex.RUnlock()
	if succ.IsZero() {
		t.Error("LastUpdateSucceededAt should be non-zero after RecordUpdateSuccess")
	}
}

// Stress the success / concurrent-flag-update surface: if any mutex is wrong,
// -race will bark. The real defence is go test -race on CI.
func TestStateServiceConcurrency(t *testing.T) {
	s := NewStateService()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				s.RecordUpdateSuccess()
			}
			s.SetAutoPingAfterConnectEnabled(i%2 == 0)
			_ = s.IsAutoPingAfterConnectEnabled()
			s.SetTemplateDirty(i%2 == 0)
			_ = s.IsTemplateDirty()
		}(i)
	}
	wg.Wait()
}
