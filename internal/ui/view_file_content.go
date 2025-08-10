package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FileContentViewModel struct {
	containerName string
	content       string
	contentPath   string
	scrollY       int
}

// render renders the file content view
func (m *FileContentViewModel) render(model *Model) string {
	if model.err != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		return errorStyle.Render(fmt.Sprintf("Error: %v", model.err))
	}

	v := viewport.New(model.width, model.Height-4)
	v.SetContent(m.content)
	v.ScrollDown(m.scrollY)
	return v.View()
}

func (m *FileContentViewModel) Load(model *Model, containerID, containerName, path string) tea.Cmd {
	model.SwitchView(FileContentView)
	model.loading = true
	m.scrollY = 0
	m.containerName = containerName

	return func() tea.Msg {
		content, err := model.dockerClient.ReadContainerFile(containerID, path)
		return fileContentLoadedMsg{
			content: content,
			path:    path,
			err:     err,
		}
	}
}

func (m *FileContentViewModel) LoadDind(model *Model, hostContainerID, containerID, containerName, path string) tea.Cmd {
	model.SwitchView(FileContentView)
	model.loading = true
	m.scrollY = 0
	m.containerName = containerName

	return func() tea.Msg {
		dindClient := model.dockerClient.Dind(hostContainerID)
		content, err := dindClient.ReadContainerFile(containerID, path)
		return fileContentLoadedMsg{
			content: content,
			path:    path,
			err:     err,
		}
	}
}

func (m *FileContentViewModel) HandleUp() tea.Cmd {
	if m.scrollY > 0 {
		m.scrollY--
	}
	return nil
}

func (m *FileContentViewModel) HandleDown(height int) tea.Cmd {
	lines := strings.Split(m.content, "\n")
	maxScroll := len(lines) - (height - 5)
	if m.scrollY < maxScroll && maxScroll > 0 {
		m.scrollY++
	}
	return nil
}

func (m *FileContentViewModel) HandleGoToStart() tea.Cmd {
	m.scrollY = 0
	return nil
}

func (m *FileContentViewModel) HandleGoToEnd(height int) tea.Cmd {
	lines := strings.Split(m.content, "\n")
	maxScroll := len(lines) - (height - 5)
	if maxScroll > 0 {
		m.scrollY = maxScroll
	}
	return nil
}

func (m *FileContentViewModel) HandleBack(model *Model) tea.Cmd {
	model.SwitchToPreviousView()
	m.content = ""
	m.contentPath = ""
	m.contentPath = ""
	m.scrollY = 0
	return nil
}

func (m *FileContentViewModel) Title() string {
	return fmt.Sprintf("File: [%d/%d] %s [%s] ",
		m.scrollY, len(strings.Split(m.content, "\n")),
		m.contentPath,
		m.containerName,
	)
}

func (m *FileContentViewModel) Loaded(content string, path string) {
	m.content = content
	m.contentPath = path
	m.scrollY = 0
}
