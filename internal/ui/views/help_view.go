package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpView represents the help screen
type HelpView struct {
	// View state
	width      int
	height     int
	scrollY    int
	helpText   []string
	sourceView string // Remember which view we came from

	// Dependencies
	rootScreen   tea.Model
	previousView tea.Model // Store the view to return to
}

// NewHelpView creates a new help view
func NewHelpView(sourceView string, previousView tea.Model) *HelpView {
	return &HelpView{
		sourceView:   sourceView,
		previousView: previousView,
		helpText:     getHelpText(sourceView),
	}
}

// SetRootScreen sets the root screen reference
func (v *HelpView) SetRootScreen(root tea.Model) {
	v.rootScreen = root
}

// Init initializes the view
func (v *HelpView) Init() tea.Cmd {
	return nil
}

// Update handles messages for this view
func (v *HelpView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if v.scrollY > 0 {
				v.scrollY--
			}
			return v, nil

		case "down", "j":
			maxScroll := len(v.helpText) - (v.height - 4)
			if v.scrollY < maxScroll && maxScroll > 0 {
				v.scrollY++
			}
			return v, nil

		case "esc", "q", "?":
			// Go back to previous view
			if v.rootScreen != nil && v.previousView != nil {
				if switcher, ok := v.rootScreen.(interface {
					SwitchScreen(tea.Model) (tea.Model, tea.Cmd)
				}); ok {
					return switcher.SwitchScreen(v.previousView)
				}
			}
			return v, nil
		}
	}

	return v, nil
}

// View renders the help screen
func (v *HelpView) View() string {
	var s strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("4")).
		Width(v.width).
		Padding(0, 1).
		Render("Help - " + v.sourceView)
	s.WriteString(header + "\n\n")

	// Calculate visible lines
	visibleHeight := v.height - 4
	start := v.scrollY
	end := start + visibleHeight
	if end > len(v.helpText) {
		end = len(v.helpText)
	}

	// Render help text
	for i := start; i < end; i++ {
		if i < len(v.helpText) {
			s.WriteString(v.helpText[i] + "\n")
		}
	}

	// Pad to fill screen
	lines := strings.Split(s.String(), "\n")
	for len(lines) < v.height-1 {
		lines = append(lines, "")
	}

	// Footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Width(v.width).
		Align(lipgloss.Center).
		Render("↑/↓: Scroll • ESC/q/?: Close Help")

	return strings.Join(lines[:v.height-1], "\n") + "\n" + footer
}

func getHelpText(viewName string) []string {
	baseHelp := []string{
		"NAVIGATION KEYS",
		"===============",
		"↑/k     Move up",
		"↓/j     Move down",
		"Enter   Select/Open",
		"ESC/q   Go back/Quit",
		"",
		"VIEW SWITCHING",
		"==============",
		"1       Docker containers",
		"2       Docker Compose projects",
		"3       Docker images",
		"4       Docker networks",
		"5       Docker volumes",
		"",
		"COMMON ACTIONS",
		"==============",
		"r       Refresh current view",
		"?       Show this help",
		":       Enter command mode",
		"",
	}

	switch viewName {
	case "Docker Containers":
		return append(baseHelp, []string{
			"DOCKER CONTAINER ACTIONS",
			"========================",
			"a       Toggle show all (including stopped)",
			"f       Browse container files",
			"!       Execute shell in container",
			"I       Inspect container",
			"K       Kill container",
			"S       Stop container",
			"U       Start container",
			"R       Restart container",
			"P       Pause/Unpause container",
			"D       Delete stopped container",
		}...)

	case "Docker Compose":
		return append(baseHelp, []string{
			"DOCKER COMPOSE ACTIONS",
			"======================",
			"a       Toggle show all containers",
			"d       View dind containers",
			"f       Browse container files",
			"!       Execute shell in container",
			"I       Inspect container",
			"t       Show process info (top)",
			"s       Show container stats",
			"K       Kill service",
			"S       Stop service",
			"U       Start service",
			"R       Restart service",
			"P       Pause/Unpause container",
			"D       Delete stopped container",
			"u       Deploy all services (up -d)",
			"x       Stop and remove all (down)",
		}...)

	case "Docker Images":
		return append(baseHelp, []string{
			"DOCKER IMAGE ACTIONS",
			"====================",
			"a       Toggle show all images",
			"I       Inspect image",
			"D       Remove image",
			"F       Force remove image",
		}...)

	default:
		return baseHelp
	}
}
