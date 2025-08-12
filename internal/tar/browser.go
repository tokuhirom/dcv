package tar

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/tokuhirom/dcv/internal/models"
)

// Browser provides tar archive browsing functionality
type Browser struct {
	tarData []byte
	cache   map[string]*FileEntry
}

// FileEntry represents a file or directory in the tar archive
type FileEntry struct {
	Header   *tar.Header
	Content  []byte
	Children map[string]*FileEntry
}

// NewBrowser creates a new tar browser from exported container data
func NewBrowser(tarData []byte) (*Browser, error) {
	b := &Browser{
		tarData: tarData,
		cache:   make(map[string]*FileEntry),
	}

	if err := b.buildCache(); err != nil {
		return nil, fmt.Errorf("failed to build tar cache: %w", err)
	}

	return b, nil
}

// buildCache reads the tar archive and builds an in-memory cache
func (b *Browser) buildCache() error {
	reader := tar.NewReader(bytes.NewReader(b.tarData))

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Read content for regular files
		var content []byte
		if header.Typeflag == tar.TypeReg {
			content, err = io.ReadAll(reader)
			if err != nil {
				return fmt.Errorf("failed to read file content: %w", err)
			}
		}

		entry := &FileEntry{
			Header:   copyHeader(header),
			Content:  content,
			Children: make(map[string]*FileEntry),
		}

		// Store in cache
		b.cache[header.Name] = entry

		// Build directory structure
		b.buildDirectoryStructure(header.Name, entry)
	}

	return nil
}

// copyHeader creates a copy of tar.Header to avoid reference issues
func copyHeader(h *tar.Header) *tar.Header {
	copied := *h
	return &copied
}

// buildDirectoryStructure creates parent directory entries if needed
func (b *Browser) buildDirectoryStructure(path string, entry *FileEntry) {
	dir := filepath.Dir(path)
	if dir == "." || dir == "/" {
		return
	}

	// Ensure parent directory exists
	if _, exists := b.cache[dir]; !exists {
		// Create synthetic directory entry
		b.cache[dir] = &FileEntry{
			Header: &tar.Header{
				Name:     dir,
				Typeflag: tar.TypeDir,
				Mode:     0755,
				ModTime:  time.Now(),
			},
			Children: make(map[string]*FileEntry),
		}

		// Recursively ensure parent directories exist
		b.buildDirectoryStructure(dir, b.cache[dir])
	}

	// Add this entry to parent's children
	parentDir := filepath.Dir(path)
	if parent, exists := b.cache[parentDir]; exists {
		baseName := filepath.Base(path)
		parent.Children[baseName] = entry
	}
}

// ListDirectory returns files in the specified directory
func (b *Browser) ListDirectory(path string) ([]models.ContainerFile, error) {
	// Normalize path
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		path = "."
	}

	var files []models.ContainerFile

	// Handle root directory specially
	if path == "/" || path == "." {
		// List top-level entries
		seen := make(map[string]bool)
		for entryPath := range b.cache {
			// Skip entries that are not at root level
			if strings.Contains(entryPath, "/") {
				parts := strings.Split(entryPath, "/")
				if len(parts) > 0 && parts[0] != "" {
					topLevel := parts[0]
					if !seen[topLevel] {
						seen[topLevel] = true
						// Check if it's a directory
						if entry, exists := b.cache[topLevel]; exists {
							files = append(files, b.entryToContainerFile(entry))
						} else {
							// Create synthetic directory entry
							files = append(files, models.ContainerFile{
								Permissions: "drwxr-xr-x",
								Size:        4096,
								Name:        topLevel,
								IsDir:       true,
							})
						}
					}
				}
			} else if entryPath != "" && entryPath != "." {
				// Direct root file
				if !seen[entryPath] {
					seen[entryPath] = true
					if entry, exists := b.cache[entryPath]; exists {
						files = append(files, b.entryToContainerFile(entry))
					}
				}
			}
		}
	} else {
		// List specific directory
		dirEntry, exists := b.cache[path]
		if !exists {
			// Try without leading slash
			path = strings.TrimPrefix(path, "/")
			dirEntry, exists = b.cache[path]
		}

		if exists && dirEntry.Header.Typeflag == tar.TypeDir {
			// List children
			for name, child := range dirEntry.Children {
				file := b.entryToContainerFile(child)
				file.Name = name
				files = append(files, file)
			}
		} else {
			// Directory doesn't exist or is not a directory
			// List entries that start with this path
			prefix := path + "/"
			seen := make(map[string]bool)

			for entryPath, entry := range b.cache {
				if strings.HasPrefix(entryPath, prefix) {
					relativePath := strings.TrimPrefix(entryPath, prefix)
					parts := strings.Split(relativePath, "/")
					if len(parts) > 0 && parts[0] != "" {
						name := parts[0]
						if !seen[name] {
							seen[name] = true
							// Check if it's a directory (has more parts)
							isDir := len(parts) > 1 || entry.Header.Typeflag == tar.TypeDir
							files = append(files, models.ContainerFile{
								Permissions: b.formatPermissions(entry.Header.Mode, isDir),
								Size:        entry.Header.Size,
								Name:        name,
								IsDir:       isDir,
							})
						}
					}
				}
			}
		}
	}

	// Add special entries if not at root
	if path != "/" && path != "." {
		files = append([]models.ContainerFile{
			{Permissions: "drwxr-xr-x", Size: 4096, Name: ".", IsDir: true},
			{Permissions: "drwxr-xr-x", Size: 4096, Name: "..", IsDir: true},
		}, files...)
	}

	return files, nil
}

// GetFileContent returns the content of a specific file
func (b *Browser) GetFileContent(path string) ([]byte, error) {
	// Normalize path
	path = strings.TrimPrefix(path, "/")

	entry, exists := b.cache[path]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", path)
	}

	if entry.Header.Typeflag != tar.TypeReg {
		return nil, fmt.Errorf("not a regular file: %s", path)
	}

	return entry.Content, nil
}

// entryToContainerFile converts a FileEntry to ContainerFile
func (b *Browser) entryToContainerFile(entry *FileEntry) models.ContainerFile {
	isDir := entry.Header.Typeflag == tar.TypeDir
	permissions := b.formatPermissions(entry.Header.Mode, isDir)

	return models.ContainerFile{
		Permissions: permissions,
		Size:        entry.Header.Size,
		Name:        filepath.Base(entry.Header.Name),
		IsDir:       isDir,
		LinkTarget:  entry.Header.Linkname,
	}
}

// formatPermissions formats file mode as Unix permission string
func (b *Browser) formatPermissions(mode int64, isDir bool) string {
	var perms strings.Builder

	// File type
	if isDir {
		perms.WriteString("d")
	} else {
		perms.WriteString("-")
	}

	// Owner permissions
	if mode&0400 != 0 {
		perms.WriteString("r")
	} else {
		perms.WriteString("-")
	}
	if mode&0200 != 0 {
		perms.WriteString("w")
	} else {
		perms.WriteString("-")
	}
	if mode&0100 != 0 {
		perms.WriteString("x")
	} else {
		perms.WriteString("-")
	}

	// Group permissions
	if mode&0040 != 0 {
		perms.WriteString("r")
	} else {
		perms.WriteString("-")
	}
	if mode&0020 != 0 {
		perms.WriteString("w")
	} else {
		perms.WriteString("-")
	}
	if mode&0010 != 0 {
		perms.WriteString("x")
	} else {
		perms.WriteString("-")
	}

	// Other permissions
	if mode&0004 != 0 {
		perms.WriteString("r")
	} else {
		perms.WriteString("-")
	}
	if mode&0002 != 0 {
		perms.WriteString("w")
	} else {
		perms.WriteString("-")
	}
	if mode&0001 != 0 {
		perms.WriteString("x")
	} else {
		perms.WriteString("-")
	}

	return perms.String()
}
