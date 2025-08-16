package docker

// HelperInjector is kept for backward compatibility
// The actual logic has been moved to view_helper_injector.go
type HelperInjector struct{}

// GetHelperPath returns the path where the helper binary will be injected
func GetHelperPath() string {
	return "/.dcv-helper"
}
