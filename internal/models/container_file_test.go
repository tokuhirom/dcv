package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLsOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []ContainerFile
	}{
		{
			name: "parse regular files and directories",
			input: `total 24
drwxr-xr-x  3 root root 4096 Dec 15 10:30 .
drwxr-xr-x  1 root root 4096 Dec 14 09:00 ..
-rw-r--r--  1 root root 1024 Dec 15 10:25 config.json
drwxr-xr-x  2 root root 4096 Dec 15 10:20 src
-rwxr-xr-x  1 root root 2048 Dec 15 10:15 app.sh`,
			expected: []ContainerFile{
				{
					Name:        ".",
					Permissions: "drwxr-xr-x",
					Mode:        "drwxr-xr-x",
					Size:        4096,
					IsDir:       true,
					Links:       "3",
					Owner:       "root",
					Group:       "root",
				},
				{
					Name:        "..",
					Permissions: "drwxr-xr-x",
					Mode:        "drwxr-xr-x",
					Size:        4096,
					IsDir:       true,
					Links:       "1",
					Owner:       "root",
					Group:       "root",
				},
				{
					Name:        "config.json",
					Permissions: "-rw-r--r--",
					Mode:        "-rw-r--r--",
					Size:        1024,
					IsDir:       false,
					Links:       "1",
					Owner:       "root",
					Group:       "root",
				},
				{
					Name:        "src",
					Permissions: "drwxr-xr-x",
					Mode:        "drwxr-xr-x",
					Size:        4096,
					IsDir:       true,
					Links:       "2",
					Owner:       "root",
					Group:       "root",
				},
				{
					Name:        "app.sh",
					Permissions: "-rwxr-xr-x",
					Mode:        "-rwxr-xr-x",
					Size:        2048,
					IsDir:       false,
					Links:       "1",
					Owner:       "root",
					Group:       "root",
				},
			},
		},
		{
			name: "handle symlinks",
			input: `total 8
lrwxrwxrwx  1 root root   10 Dec 15 10:30 link -> /etc/hosts
-rw-r--r--  1 root root 1024 Dec 15 10:25 file.txt`,
			expected: []ContainerFile{
				{
					Name:        "link",
					Permissions: "lrwxrwxrwx",
					Mode:        "lrwxrwxrwx",
					Size:        10,
					IsDir:       false,
					LinkTarget:  "/etc/hosts",
					Links:       "1",
					Owner:       "root",
					Group:       "root",
				},
				{
					Name:        "file.txt",
					Permissions: "-rw-r--r--",
					Mode:        "-rw-r--r--",
					Size:        1024,
					IsDir:       false,
					Links:       "1",
					Owner:       "root",
					Group:       "root",
				},
			},
		},
		{
			name: "handle filenames with spaces",
			input: `total 8
-rw-r--r--  1 root root 1024 Dec 15 10:25 my file.txt
drwxr-xr-x  2 root root 4096 Dec 15 10:20 my folder`,
			expected: []ContainerFile{
				{
					Name:        "my file.txt",
					Permissions: "-rw-r--r--",
					Mode:        "-rw-r--r--",
					Size:        1024,
					IsDir:       false,
					Links:       "1",
					Owner:       "root",
					Group:       "root",
				},
				{
					Name:        "my folder",
					Permissions: "drwxr-xr-x",
					Mode:        "drwxr-xr-x",
					Size:        4096,
					IsDir:       true,
					Links:       "2",
					Owner:       "root",
					Group:       "root",
				},
			},
		},
		{
			name:     "handle empty output",
			input:    "",
			expected: []ContainerFile{},
		},
		{
			name:     "skip total line",
			input:    "total 0",
			expected: []ContainerFile{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseLsOutput(tt.input)

			// Check length first
			assert.Equal(t, len(tt.expected), len(result))

			// Compare each file, ignoring ModTime
			for i := range tt.expected {
				if i < len(result) {
					assert.Equal(t, tt.expected[i].Name, result[i].Name)
					assert.Equal(t, tt.expected[i].Permissions, result[i].Permissions)
					assert.Equal(t, tt.expected[i].Mode, result[i].Mode)
					assert.Equal(t, tt.expected[i].Size, result[i].Size)
					assert.Equal(t, tt.expected[i].IsDir, result[i].IsDir)
					assert.Equal(t, tt.expected[i].LinkTarget, result[i].LinkTarget)
					assert.Equal(t, tt.expected[i].Owner, result[i].Owner)
					assert.Equal(t, tt.expected[i].Group, result[i].Group)
					assert.Equal(t, tt.expected[i].Links, result[i].Links)
					// ModTime is set to current time, so we don't compare it
				}
			}
		})
	}
}

func TestGetSizeString(t *testing.T) {
	tests := []struct {
		name     string
		file     ContainerFile
		expected string
	}{
		{
			name: "directory shows dash",
			file: ContainerFile{
				IsDir: true,
				Size:  4096,
			},
			expected: "-",
		},
		{
			name: "bytes",
			file: ContainerFile{
				IsDir: false,
				Size:  512,
			},
			expected: "512",
		},
		{
			name: "kilobytes",
			file: ContainerFile{
				IsDir: false,
				Size:  1536, // 1.5K
			},
			expected: "1.5K",
		},
		{
			name: "megabytes",
			file: ContainerFile{
				IsDir: false,
				Size:  1572864, // 1.5M
			},
			expected: "1.5M",
		},
		{
			name: "gigabytes",
			file: ContainerFile{
				IsDir: false,
				Size:  1610612736, // 1.5G
			},
			expected: "1.5G",
		},
		{
			name: "exactly 1K",
			file: ContainerFile{
				IsDir: false,
				Size:  1024,
			},
			expected: "1.0K",
		},
		{
			name: "exactly 1M",
			file: ContainerFile{
				IsDir: false,
				Size:  1048576,
			},
			expected: "1.0M",
		},
		{
			name: "zero bytes",
			file: ContainerFile{
				IsDir: false,
				Size:  0,
			},
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.file.GetSizeString()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		file     ContainerFile
		expected string
	}{
		{
			name: "regular file",
			file: ContainerFile{
				Name:  "file.txt",
				IsDir: false,
			},
			expected: "file.txt",
		},
		{
			name: "directory",
			file: ContainerFile{
				Name:  "folder",
				IsDir: true,
			},
			expected: "folder/",
		},
		{
			name: "symlink",
			file: ContainerFile{
				Name:       "link",
				IsDir:      false,
				LinkTarget: "/etc/hosts",
			},
			expected: "link -> /etc/hosts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.file.GetDisplayName()
			assert.Equal(t, tt.expected, result)
		})
	}
}
