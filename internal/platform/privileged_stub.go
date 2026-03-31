//go:build !darwin
// +build !darwin

package platform

import "errors"

// errPrivilegedNotSupported is returned by RunWithPrivileges on non-darwin platforms.
var errPrivilegedNotSupported = errors.New("privileged execution not supported on this platform")

// Privileged script/pid names and pkill pattern are empty on non-darwin.
const (
	PrivilegedScriptName   = ""
	PrivilegedPidFileName  = ""
	PrivilegedPkillPattern = ""
)

// RunWithPrivileges runs a command with elevated privileges (macOS only).
// On non-darwin platforms it returns (0, 0, error).
func RunWithPrivileges(toolPath string, args []string) (scriptPID, singboxPID int, err error) {
	_ = toolPath
	_ = args
	return 0, 0, errPrivilegedNotSupported
}

// WritePrivilegedStartScript is not used on non-darwin.
func WritePrivilegedStartScript(scriptPath, pidFilePath, binDir, singboxPath, configName, logPath string) error {
	_ = scriptPath
	_ = pidFilePath
	_ = binDir
	_ = singboxPath
	_ = configName
	_ = logPath
	return errPrivilegedNotSupported
}

// KillPrivilegedProcess is a no-op on non-darwin (privileged mode is macOS-only).
func KillPrivilegedProcess(scriptPID, singboxPID int, pidFile string) error {
	_ = scriptPID
	_ = singboxPID
	_ = pidFile
	return nil
}

// WaitForPrivilegedExit is a no-op on non-darwin.
func WaitForPrivilegedExit(pid int) {
	_ = pid
}

// FreePrivilegedAuthorization is a no-op on non-darwin.
func FreePrivilegedAuthorization() {}
