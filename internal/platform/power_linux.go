//go:build linux
// +build linux

package platform

// Linux power-event hook: subscribe to systemd-logind's PrepareForSleep signal
// on the system DBus. The signal fires twice per sleep cycle — payload `true`
// right before sleep, `false` right after wake. We dispatch sleep- and
// resume-callbacks off the matching edge.
//
// References:
//   https://www.freedesktop.org/wiki/Software/systemd/inhibit/
//   https://www.freedesktop.org/wiki/Software/systemd/logind/
//
// Notes on robustness:
//   - If the user has no systemd (e.g. WSL without logind, or a minimal
//     distro), NewSystemBus fails silently and the callbacks never fire.
//     That matches the stub behavior — no regression.
//   - Dispatch runs in a dedicated goroutine; user callbacks should be quick
//     or fan out to their own goroutine. Matching the Windows implementation.
//   - sleepingFlag is kept in sync so IsSleeping() returns something useful
//     for other subsystems (e.g. auto_update.go skips its work while true).

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/godbus/dbus/v5"

	"singbox-launcher/internal/debuglog"
)

var (
	powerCallbacksMu sync.Mutex
	sleepCallbacks   []func()
	resumeCallbacks  []func()
	listenerStarted  bool

	sleepingFlag atomic.Bool

	powerCtx       context.Context = context.Background()
	powerCtxCancel context.CancelFunc
)

// IsSleeping reports whether systemd-logind has signalled sleep but not yet
// resume. True in the brief window between PrepareForSleep(true) and
// PrepareForSleep(false).
func IsSleeping() bool {
	return sleepingFlag.Load()
}

// PowerContext returns a context that callers can use for outgoing work;
// cancelled at each suspend transition and replaced on resume so stale
// requests that started before sleep don't linger on stale sockets.
func PowerContext() context.Context {
	powerCallbacksMu.Lock()
	defer powerCallbacksMu.Unlock()
	return powerCtx
}

// RegisterSleepCallback registers fn to run when the system is about to sleep.
func RegisterSleepCallback(fn func()) {
	if fn == nil {
		return
	}
	powerCallbacksMu.Lock()
	sleepCallbacks = append(sleepCallbacks, fn)
	startListenerLocked()
	powerCallbacksMu.Unlock()
}

// RegisterPowerResumeCallback registers fn to run when the system resumes.
func RegisterPowerResumeCallback(fn func()) {
	if fn == nil {
		return
	}
	powerCallbacksMu.Lock()
	resumeCallbacks = append(resumeCallbacks, fn)
	startListenerLocked()
	powerCallbacksMu.Unlock()
}

// StopPowerResumeListener tears the listener down — optional, idempotent.
func StopPowerResumeListener() {
	powerCallbacksMu.Lock()
	defer powerCallbacksMu.Unlock()
	if powerCtxCancel != nil {
		powerCtxCancel()
	}
	listenerStarted = false
	powerCtx = context.Background()
	powerCtxCancel = nil
}

// startListenerLocked fires up the DBus match+dispatch goroutine exactly once.
// Must be called with powerCallbacksMu held.
func startListenerLocked() {
	if listenerStarted {
		return
	}

	conn, err := dbus.SystemBus()
	if err != nil {
		debuglog.DebugLog("platform/power_linux: no system DBus (%v) — power hooks inactive", err)
		return
	}
	rule := "type='signal',interface='org.freedesktop.login1.Manager',member='PrepareForSleep',path='/org/freedesktop/login1'"
	call := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, rule)
	if call.Err != nil {
		debuglog.WarnLog("platform/power_linux: AddMatch failed: %v", call.Err)
		return
	}

	ch := make(chan *dbus.Signal, 8)
	conn.Signal(ch)

	ctx, cancel := context.WithCancel(context.Background())
	powerCtx = ctx
	powerCtxCancel = cancel
	listenerStarted = true

	go func() {
		for sig := range ch {
			if sig == nil || sig.Name != "org.freedesktop.login1.Manager.PrepareForSleep" {
				continue
			}
			if len(sig.Body) == 0 {
				continue
			}
			going, ok := sig.Body[0].(bool)
			if !ok {
				continue
			}
			if going {
				sleepingFlag.Store(true)
				powerCallbacksMu.Lock()
				if powerCtxCancel != nil {
					powerCtxCancel()
				}
				powerCtx = context.Background()
				powerCtxCancel = nil
				cbs := append([]func(){}, sleepCallbacks...)
				powerCallbacksMu.Unlock()
				for _, cb := range cbs {
					cb()
				}
			} else {
				sleepingFlag.Store(false)
				powerCallbacksMu.Lock()
				ctx, cancel := context.WithCancel(context.Background())
				powerCtx = ctx
				powerCtxCancel = cancel
				cbs := append([]func(){}, resumeCallbacks...)
				powerCallbacksMu.Unlock()
				for _, cb := range cbs {
					cb()
				}
			}
		}
	}()
	debuglog.InfoLog("platform/power_linux: subscribed to systemd-logind PrepareForSleep")
}
