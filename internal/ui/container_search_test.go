package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/tokuhirom/dcv/internal/models"
)

func TestContainerSearch(t *testing.T) {
	t.Run("search in compose process list", func(t *testing.T) {
		// Create model with compose process list view
		m := createTestModel(ComposeProcessListView)
		m.initializeKeyHandlers()
		m.composeProcessListViewModel.Init(m) // Initialize the view model
		m.composeProcessListViewModel.composeContainers = []models.ComposeContainer{
			{Service: "web", Image: "nginx:latest", State: "running"},
			{Service: "db", Image: "postgres:14", State: "running"},
			{Service: "redis", Image: "redis:alpine", State: "running"},
		}
		m.composeProcessListViewModel.SetRows(
			m.composeProcessListViewModel.buildRows(),
			10,
		)

		// Start search mode
		assert.False(t, m.composeProcessListViewModel.IsSearchActive())
		newModel, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
		m = newModel.(*Model)
		assert.True(t, m.composeProcessListViewModel.IsSearchActive())

		// Type search term "db"
		newModel, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
		m = newModel.(*Model)
		newModel, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
		m = newModel.(*Model)

		// Check search text is set
		assert.Equal(t, "db", m.composeProcessListViewModel.GetSearchText())

		// Exit search mode
		newModel, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyEsc})
		m = newModel.(*Model)
		assert.False(t, m.composeProcessListViewModel.IsSearchActive())

		// Cursor should be on the matching container
		assert.Equal(t, 1, m.composeProcessListViewModel.Cursor) // db container is at index 1
	})

	t.Run("search in docker container list", func(t *testing.T) {
		// Create model with docker container list view
		m := createTestModel(DockerContainerListView)
		m.initializeKeyHandlers()
		m.dockerContainerListViewModel.Init(m) // Initialize the view model
		m.dockerContainerListViewModel.dockerContainers = []models.DockerContainer{
			{Names: "nginx", Image: "nginx:latest", State: "running"},
			{Names: "postgres", Image: "postgres:14", State: "running"},
			{Names: "redis", Image: "redis:alpine", State: "running"},
		}
		m.dockerContainerListViewModel.SetRows(
			m.dockerContainerListViewModel.buildRows(),
			10,
		)

		// Start search mode
		assert.False(t, m.dockerContainerListViewModel.IsSearchActive())
		newModel, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
		m = newModel.(*Model)
		assert.True(t, m.dockerContainerListViewModel.IsSearchActive())

		// Type search term "redis"
		for _, r := range "redis" {
			newModel, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			m = newModel.(*Model)
		}

		// Check search text is set
		assert.Equal(t, "redis", m.dockerContainerListViewModel.GetSearchText())

		// Exit search mode
		newModel, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyEnter})
		m = newModel.(*Model)
		assert.False(t, m.dockerContainerListViewModel.IsSearchActive())

		// Cursor should be on the matching container
		assert.Equal(t, 2, m.dockerContainerListViewModel.Cursor) // redis container is at index 2
	})

	t.Run("navigate search results with n and N", func(t *testing.T) {
		// Create model with compose process list view
		m := createTestModel(ComposeProcessListView)
		m.initializeKeyHandlers()
		m.composeProcessListViewModel.Init(m) // Initialize the view model
		m.composeProcessListViewModel.composeContainers = []models.ComposeContainer{
			{Service: "web1", Image: "nginx:latest", State: "running"},
			{Service: "web2", Image: "nginx:latest", State: "running"},
			{Service: "db", Image: "postgres:14", State: "running"},
			{Service: "web3", Image: "nginx:latest", State: "running"},
		}
		m.composeProcessListViewModel.SetRows(
			m.composeProcessListViewModel.buildRows(),
			10,
		)

		// Start search and search for "web"
		newModel, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
		m = newModel.(*Model)
		for _, r := range "web" {
			newModel, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			m = newModel.(*Model)
		}
		newModel, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyEnter})
		m = newModel.(*Model)

		// Should be on first match (web1 at index 0)
		assert.Equal(t, 0, m.composeProcessListViewModel.Cursor)

		// Press 'n' to go to next match (web2 at index 1)
		newModel, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
		m = newModel.(*Model)
		assert.Equal(t, 1, m.composeProcessListViewModel.Cursor)

		// Press 'n' again to go to next match (web3 at index 3)
		newModel, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
		m = newModel.(*Model)
		assert.Equal(t, 3, m.composeProcessListViewModel.Cursor)

		// Press 'N' to go to previous match (web2 at index 1)
		newModel, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("N")})
		m = newModel.(*Model)
		assert.Equal(t, 1, m.composeProcessListViewModel.Cursor)
	})
}
