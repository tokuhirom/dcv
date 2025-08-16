package docker

import (
	_ "embed"
	"fmt"
	"runtime"
)

// Embed the helper binaries at compile time
// These files must exist before 'go build' is run (use 'make build-helpers' first)

//go:embed static-binaries/dcv-helper-amd64
var helperBinaryAMD64 []byte

//go:embed static-binaries/dcv-helper-arm64
var helperBinaryARM64 []byte

//go:embed static-binaries/dcv-helper-arm
var helperBinaryARM []byte

// GetHelperBinary returns the appropriate helper binary for the current or specified architecture
func GetHelperBinary(arch string) ([]byte, error) {
	// If arch not specified, use runtime arch
	if arch == "" {
		arch = runtime.GOARCH
	}

	switch arch {
	case "amd64", "x86_64":
		if len(helperBinaryAMD64) == 0 {
			return nil, fmt.Errorf("helper binary for amd64 not embedded (run 'make build-helpers' first)")
		}
		return helperBinaryAMD64, nil
	case "arm64", "aarch64":
		if len(helperBinaryARM64) == 0 {
			return nil, fmt.Errorf("helper binary for arm64 not embedded (run 'make build-helpers' first)")
		}
		return helperBinaryARM64, nil
	case "arm", "armv7", "armv7l", "armhf":
		if len(helperBinaryARM) == 0 {
			return nil, fmt.Errorf("helper binary for arm not embedded (run 'make build-helpers' first)")
		}
		return helperBinaryARM, nil
	default:
		return nil, fmt.Errorf("unsupported architecture: %s", arch)
	}
}
