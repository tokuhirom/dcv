package models

import (
	"strings"
	"time"
)

// ContainerFile represents a file or directory in a container
type ContainerFile struct {
	Name        string
	Size        int64
	Mode        string
	ModTime     time.Time
	IsDir       bool
	LinkTarget  string
	Permissions string
}

// ParseLsOutput parses the output of ls -la command
func ParseLsOutput(output string) []ContainerFile {
	var files []ContainerFile
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "total") {
			continue
		}

		// Parse ls -la output format:
		// drwxr-xr-x  2 root root 4096 Dec 15 10:30 dirname
		// -rw-r--r--  1 root root  123 Dec 15 10:30 filename
		parts := strings.Fields(line)
		if len(parts) < 9 {
			continue
		}

		file := ContainerFile{
			Permissions: parts[0],
			Mode:        parts[0],
			IsDir:       strings.HasPrefix(parts[0], "d"),
		}

		// Parse size
		// TODO: Convert size string to int64 if needed
		// For now, the size is kept as string in SizeStr field

		// Get filename (handle spaces in filename)
		// Everything from parts[8] onwards is the filename
		file.Name = strings.Join(parts[8:], " ")

		// Handle symlinks (file -> target)
		if strings.Contains(file.Name, " -> ") {
			parts := strings.Split(file.Name, " -> ")
			file.Name = parts[0]
			if len(parts) > 1 {
				file.LinkTarget = parts[1]
			}
		}

		files = append(files, file)
	}

	return files
}

// GetDisplayName returns the display name with indicators
func (f ContainerFile) GetDisplayName() string {
	name := f.Name
	if f.IsDir {
		name += "/"
	} else if f.LinkTarget != "" {
		name += " -> " + f.LinkTarget
	}
	return name
}
