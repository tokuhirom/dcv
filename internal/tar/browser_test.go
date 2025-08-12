package tar

import (
	"archive/tar"
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestTar(t *testing.T) []byte {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// Create test files and directories
	files := []struct {
		name     string
		content  string
		typeflag byte
		mode     int64
	}{
		{name: "etc/", typeflag: tar.TypeDir, mode: 0755},
		{name: "etc/passwd", content: "root:x:0:0:root:/root:/bin/bash\n", typeflag: tar.TypeReg, mode: 0644},
		{name: "etc/hosts", content: "127.0.0.1 localhost\n", typeflag: tar.TypeReg, mode: 0644},
		{name: "usr/", typeflag: tar.TypeDir, mode: 0755},
		{name: "usr/bin/", typeflag: tar.TypeDir, mode: 0755},
		{name: "usr/bin/ls", content: "#!/bin/sh\n", typeflag: tar.TypeReg, mode: 0755},
		{name: "home/", typeflag: tar.TypeDir, mode: 0755},
		{name: "home/user/", typeflag: tar.TypeDir, mode: 0755},
		{name: "home/user/.bashrc", content: "export PS1='$ '\n", typeflag: tar.TypeReg, mode: 0644},
		{name: "tmp/", typeflag: tar.TypeDir, mode: 0777},
		{name: "README.md", content: "# Test Archive\n", typeflag: tar.TypeReg, mode: 0644},
	}

	for _, file := range files {
		hdr := &tar.Header{
			Name:     file.name,
			Mode:     file.mode,
			Typeflag: file.typeflag,
			Size:     int64(len(file.content)),
			ModTime:  time.Now(),
		}

		err := tw.WriteHeader(hdr)
		require.NoError(t, err)

		if file.typeflag == tar.TypeReg && file.content != "" {
			_, err = tw.Write([]byte(file.content))
			require.NoError(t, err)
		}
	}

	err := tw.Close()
	require.NoError(t, err)

	return buf.Bytes()
}

func TestNewBrowser(t *testing.T) {
	tarData := createTestTar(t)
	browser, err := NewBrowser(tarData)

	assert.NoError(t, err)
	assert.NotNil(t, browser)
	assert.NotEmpty(t, browser.cache)
}

func TestListDirectoryRoot(t *testing.T) {
	tarData := createTestTar(t)
	browser, err := NewBrowser(tarData)
	require.NoError(t, err)

	// Test listing root directory
	files, err := browser.ListDirectory("/")
	assert.NoError(t, err)
	assert.NotEmpty(t, files)

	// Check if expected directories are present
	foundDirs := make(map[string]bool)
	for _, file := range files {
		foundDirs[file.Name] = true
	}

	assert.True(t, foundDirs["etc"])
	assert.True(t, foundDirs["usr"])
	assert.True(t, foundDirs["home"])
	assert.True(t, foundDirs["tmp"])
	assert.True(t, foundDirs["README.md"])
}

func TestListDirectorySubdir(t *testing.T) {
	tarData := createTestTar(t)
	browser, err := NewBrowser(tarData)
	require.NoError(t, err)

	// Test listing /etc directory
	files, err := browser.ListDirectory("/etc")
	assert.NoError(t, err)
	assert.NotEmpty(t, files)

	// Should have . and .. entries
	assert.Equal(t, ".", files[0].Name)
	assert.Equal(t, "..", files[1].Name)

	// Check if expected files are present
	foundFiles := make(map[string]bool)
	for _, file := range files {
		foundFiles[file.Name] = true
	}

	assert.True(t, foundFiles["passwd"])
	assert.True(t, foundFiles["hosts"])
}

func TestListDirectoryNestedDir(t *testing.T) {
	tarData := createTestTar(t)
	browser, err := NewBrowser(tarData)
	require.NoError(t, err)

	// Test listing /usr/bin directory
	files, err := browser.ListDirectory("/usr/bin")
	assert.NoError(t, err)
	assert.NotEmpty(t, files)

	// Check if ls file is present
	foundFiles := make(map[string]bool)
	for _, file := range files {
		foundFiles[file.Name] = true
	}

	assert.True(t, foundFiles["ls"])
}

func TestGetFileContent(t *testing.T) {
	tarData := createTestTar(t)
	browser, err := NewBrowser(tarData)
	require.NoError(t, err)

	// Test reading /etc/hosts
	content, err := browser.GetFileContent("/etc/hosts")
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1 localhost\n", string(content))

	// Test reading /home/user/.bashrc
	content, err = browser.GetFileContent("/home/user/.bashrc")
	assert.NoError(t, err)
	assert.Equal(t, "export PS1='$ '\n", string(content))

	// Test reading README.md from root
	content, err = browser.GetFileContent("README.md")
	assert.NoError(t, err)
	assert.Equal(t, "# Test Archive\n", string(content))
}

func TestGetFileContentErrors(t *testing.T) {
	tarData := createTestTar(t)
	browser, err := NewBrowser(tarData)
	require.NoError(t, err)

	// Test reading non-existent file
	_, err = browser.GetFileContent("/nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")

	// Test reading directory as file
	_, err = browser.GetFileContent("/etc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a regular file")
}

func TestFormatPermissions(t *testing.T) {
	tarData := createTestTar(t)
	browser, err := NewBrowser(tarData)
	require.NoError(t, err)

	// Test file permissions
	files, err := browser.ListDirectory("/etc")
	require.NoError(t, err)

	for _, file := range files {
		switch file.Name {
		case "passwd":
			assert.Equal(t, "-rw-r--r--", file.Permissions)
		case ".", "..":
			assert.Equal(t, "drwxr-xr-x", file.Permissions)
		}
	}

	// Test executable file permissions
	files, err = browser.ListDirectory("/usr/bin")
	require.NoError(t, err)

	for _, file := range files {
		if file.Name == "ls" {
			assert.Equal(t, "-rwxr-xr-x", file.Permissions)
		}
	}
}

func TestEmptyTar(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	err := tw.Close()
	require.NoError(t, err)

	browser, err := NewBrowser(buf.Bytes())
	assert.NoError(t, err)
	assert.NotNil(t, browser)

	// Should return empty list for root
	files, err := browser.ListDirectory("/")
	assert.NoError(t, err)
	assert.Empty(t, files)
}
