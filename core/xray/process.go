package xray

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"singbox-launcher/internal/debuglog"
	"singbox-launcher/internal/platform"
)

// SidecarProcess manages a single Xray-core sidecar process.
type SidecarProcess struct {
	cmd      *exec.Cmd
	mu       sync.Mutex
	running  bool
	configPath string
}

// Start launches the Xray process with the given config file.
func (sp *SidecarProcess) Start(xrayPath, configPath string) error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if sp.running && sp.cmd != nil && sp.cmd.Process != nil {
		debuglog.DebugLog("XraySidecar: already running (PID=%d), stopping first", sp.cmd.Process.Pid)
		sp.stopUnlocked()
	}

	if _, err := os.Stat(xrayPath); err != nil {
		return fmt.Errorf("xray binary not found at %s: %w", xrayPath, err)
	}
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("xray config not found at %s: %w", configPath, err)
	}

	sp.configPath = configPath
	sp.cmd = exec.Command(xrayPath, "run", "-c", configPath)
	platform.PrepareCommand(sp.cmd)

	// Log to a dedicated sidecar log file
	logPath := filepath.Join(filepath.Dir(configPath), "xray-sidecar.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, platform.DefaultFileMode)
	if err == nil {
		sp.cmd.Stdout = logFile
		sp.cmd.Stderr = logFile
	} else {
		debuglog.WarnLog("XraySidecar: cannot open log file %s: %v", logPath, err)
	}

	if err := sp.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start xray: %w", err)
	}
	sp.running = true
	debuglog.InfoLog("XraySidecar: started xray (PID=%d) with config %s", sp.cmd.Process.Pid, configPath)

	go sp.monitor()
	return nil
}

// Stop terminates the Xray process.
func (sp *SidecarProcess) Stop() {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.stopUnlocked()
}

func (sp *SidecarProcess) stopUnlocked() {
	if !sp.running || sp.cmd == nil || sp.cmd.Process == nil {
		sp.running = false
		return
	}

	debuglog.InfoLog("XraySidecar: stopping xray (PID=%d)...", sp.cmd.Process.Pid)
	if runtime.GOOS == "windows" {
		_ = platform.KillProcessByPID(sp.cmd.Process.Pid)
	} else {
		_ = sp.cmd.Process.Signal(os.Interrupt)
		// Give it a moment, then force kill
		go func(pid int) {
			<-time.After(2 * time.Second)
			if p, found, err := findProcess(pid); err == nil && found {
				_ = p.Kill()
				_ = p // suppress unused
			}
		}(sp.cmd.Process.Pid)
	}
	sp.running = false
	sp.cmd = nil
}

// IsRunning reports whether the sidecar process is currently running.
func (sp *SidecarProcess) IsRunning() bool {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	if !sp.running || sp.cmd == nil || sp.cmd.Process == nil {
		return false
	}
	// Verify process still exists
	if runtime.GOOS == "windows" {
		_, found, _ := findProcess(sp.cmd.Process.Pid)
		return found
	}
	return sp.cmd.ProcessState == nil || !sp.cmd.ProcessState.Exited()
}

// GetVersion runs `xray version` and returns the version string.
func GetVersion(xrayPath string) (string, error) {
	if _, err := os.Stat(xrayPath); err != nil {
		return "", fmt.Errorf("xray binary not found: %w", err)
	}
	cmd := exec.Command(xrayPath, "version")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("xray version failed: %w", err)
	}
	lines := strings.SplitN(string(out), "\n", 2)
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}
	return "", fmt.Errorf("empty version output")
}

// monitor waits for the Xray process to exit and logs the result.
func (sp *SidecarProcess) monitor() {
	if sp.cmd == nil {
		return
	}
	err := sp.cmd.Wait()
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.running = false
	if err != nil {
		debuglog.WarnLog("XraySidecar: xray exited with error: %v", err)
	} else {
		debuglog.InfoLog("XraySidecar: xray exited cleanly")
	}
}

// findProcess is a small helper to check if a process exists.
func findProcess(pid int) (*os.Process, bool, error) {
	p, err := os.FindProcess(pid)
	if err != nil {
		return nil, false, err
	}
	// On Unix, FindProcess always succeeds; need to send signal 0 to check.
	if runtime.GOOS != "windows" {
		if err := p.Signal(os.Signal(nil)); err != nil {
			return nil, false, err
		}
	}
	return p, true, nil
}
