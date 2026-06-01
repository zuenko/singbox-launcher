package core

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"singbox-launcher/core/xray"
	"singbox-launcher/internal/debuglog"
)

// XraySidecarService coordinates the Xray sidecar process for XHTTP outbounds.
type XraySidecarService struct {
	ac *AppController

	mu            sync.Mutex
	process       *xray.SidecarProcess
	registry      *xray.SidecarRegistry
	lastActiveTag string
}

// NewXraySidecarService creates a new sidecar service bound to the controller.
func NewXraySidecarService(ac *AppController) *XraySidecarService {
	return &XraySidecarService{ac: ac}
}

// SetRegistry updates the sidecar registry after a config rebuild.
func (svc *XraySidecarService) SetRegistry(r *xray.SidecarRegistry) {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	svc.registry = r
}

// Registry returns the current registry (may be nil).
func (svc *XraySidecarService) Registry() *xray.SidecarRegistry {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	return svc.registry
}

// StartForOutbound starts the Xray sidecar for the given outbound tag.
// If the tag is not registered in the sidecar registry, this is a no-op.
func (svc *XraySidecarService) StartForOutbound(tag string) error {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	if svc.registry == nil {
		debuglog.WarnLog("XraySidecarService: StartForOutbound called with nil registry — was RebuildConfigIfDirty skipped?")
		return nil
	}
	entry, ok := svc.registry.Get(tag)
	if !ok {
		return nil
	}

	xrayPath := svc.resolveXrayPath()
	if xrayPath == "" {
		return fmt.Errorf("xray core not found")
	}

	if svc.process == nil {
		svc.process = &xray.SidecarProcess{}
	}

	// Write temp config
	tempDir := filepath.Join(svc.ac.FileService.ExecDir, "temp")
	_ = os.MkdirAll(tempDir, 0755)
	tempConfig := filepath.Join(tempDir, fmt.Sprintf("xray-%s.json", tag))

	cfgJSON, err := xray.BuildSidecarConfig(entry.OriginalOutbound, entry.Port)
	if err != nil {
		return fmt.Errorf("build xray config: %w", err)
	}
	if err := os.WriteFile(tempConfig, cfgJSON, 0644); err != nil {
		return fmt.Errorf("write xray config: %w", err)
	}

	if err := svc.process.Start(xrayPath, tempConfig); err != nil {
		return fmt.Errorf("start xray sidecar: %w", err)
	}
	svc.lastActiveTag = tag
	debuglog.InfoLog("XraySidecarService: started sidecar for outbound %s on port %d", tag, entry.Port)
	return nil
}

// Stop stops the Xray sidecar process.
func (svc *XraySidecarService) Stop() {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	if svc.process != nil {
		svc.process.Stop()
		svc.process = nil
	}
}

// IsRunning reports whether the sidecar is currently running.
func (svc *XraySidecarService) IsRunning() bool {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	if svc.process == nil {
		return false
	}
	return svc.process.IsRunning()
}

// GetVersion returns the installed Xray core version, or empty string.
func (svc *XraySidecarService) GetVersion() string {
	xrayPath := svc.resolveXrayPath()
	if xrayPath == "" {
		return ""
	}
	ver, err := xray.GetVersion(xrayPath)
	if err != nil {
		debuglog.DebugLog("XraySidecarService: version check failed: %v", err)
		return ""
	}
	return ver
}

// resolveXrayPath returns the path to the xray executable, or empty if not found.
func (svc *XraySidecarService) resolveXrayPath() string {
	if svc.ac == nil || svc.ac.FileService == nil {
		return ""
	}
	execDir := svc.ac.FileService.ExecDir
	binDir := filepath.Join(execDir, "bin")

	var name string
	if runtime.GOOS == "windows" {
		name = "xray.exe"
	} else {
		name = "xray"
	}

	// Try bin/ first
	p := filepath.Join(binDir, name)
	if _, err := os.Stat(p); err == nil {
		return p
	}
	// Try execDir root (legacy)
	p = filepath.Join(execDir, name)
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return ""
}

// XrayPath returns the resolved xray path for external use (UI, etc.).
func (svc *XraySidecarService) XrayPath() string {
	return svc.resolveXrayPath()
}

// LastActiveTag returns the last outbound tag for which the sidecar was started.
func (svc *XraySidecarService) LastActiveTag() string {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	return svc.lastActiveTag
}

// Enabled returns whether the sidecar feature is enabled.
// For v1: always enabled if xray binary exists.
func (svc *XraySidecarService) Enabled() bool {
	return svc.resolveXrayPath() != ""
}
