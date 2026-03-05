package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/tokuhirom/dcv/internal/docker"
)

// KeyHandler represents a function that handles a key press
type KeyHandler func(msg tea.KeyPressMsg) (tea.Model, tea.Cmd)

// KeyConfig represents a key binding configuration
type KeyConfig struct {
	Keys        []string
	Description string
	KeyHandler  KeyHandler
}

type GetContainerAware interface {
	GetContainer(model *Model) *docker.Container
}

type HandleInspectAware interface {
	HandleInspect(model *Model) tea.Cmd
}

type UpdateAware interface {
	Update(model *Model, msg tea.Msg) (tea.Model, tea.Cmd)
}
