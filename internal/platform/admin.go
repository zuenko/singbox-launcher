//go:build !windows

package platform

// IsAdmin is a no-op on non-Windows platforms.
func IsAdmin() bool {
	return true
}
