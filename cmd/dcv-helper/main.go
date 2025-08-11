// dcv-helper: Minimal static binary for file operations in containers
// Build with: CGO_ENABLED=0 go build -ldflags="-s -w" -o dcv-helper main.go
package main

import (
	"fmt"
	"io"
	"os"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "ls":
		cmdLs()
	case "cat":
		cmdCat()
	case "version":
		fmt.Println("dcv-helper", version)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: dcv-helper <command> [args...]")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  ls [path]    - List directory contents")
	fmt.Fprintln(os.Stderr, "  cat <file>   - Display file contents")
	fmt.Fprintln(os.Stderr, "  version      - Show version")
}

// cmdLs implements a simple ls command
func cmdLs() {
	path := "."
	if len(os.Args) > 2 {
		path = os.Args[2]
	}

	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ls: %s: %v\n", path, err)
		os.Exit(1)
	}

	if !info.IsDir() {
		// Single file
		printFileInfo(info)
		return
	}

	// Directory listing
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ls: %s: %v\n", path, err)
		os.Exit(1)
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		printFileInfo(info)
	}
}

// printFileInfo prints file information in a simple format
func printFileInfo(info os.FileInfo) {
	mode := info.Mode()
	typeChar := '-'

	if mode.IsDir() {
		typeChar = 'd'
	} else if mode&os.ModeSymlink != 0 {
		typeChar = 'l'
	} else if mode&os.ModeDevice != 0 {
		typeChar = 'b'
	} else if mode&os.ModeCharDevice != 0 {
		typeChar = 'c'
	} else if mode&os.ModeNamedPipe != 0 {
		typeChar = 'p'
	} else if mode&os.ModeSocket != 0 {
		typeChar = 's'
	}

	// Format: type permissions size name
	// Simple format: drwxr-xr-x 4096 dirname
	fmt.Printf("%c%s %10d %s\n",
		typeChar,
		formatMode(mode),
		info.Size(),
		info.Name(),
	)
}

// formatMode formats file mode as rwxrwxrwx
func formatMode(mode os.FileMode) string {
	const str = "rwxrwxrwx"
	var buf [9]byte

	perm := mode.Perm()
	for i := 0; i < 9; i++ {
		if perm&(1<<uint(8-i)) != 0 {
			buf[i] = str[i]
		} else {
			buf[i] = '-'
		}
	}
	return string(buf[:])
}

// cmdCat implements a simple cat command
func cmdCat() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "cat: missing file operand")
		os.Exit(1)
	}

	exitCode := 0
	for i := 2; i < len(os.Args); i++ {
		if err := catFile(os.Args[i]); err != nil {
			fmt.Fprintf(os.Stderr, "cat: %s: %v\n", os.Args[i], err)
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}

// catFile displays contents of a single file
func catFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	_, err = io.Copy(os.Stdout, file)
	return err
}
