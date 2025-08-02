package ui

import (
	"fmt"
	"strings"
)

func (m *Model) renderTopView() string {
	var s strings.Builder

	title := titleStyle.Render(fmt.Sprintf("Process Info: %s", m.topService))
	s.WriteString(title + "\n\n")

	if m.loading {
		s.WriteString("Loading...\n")
		return s.String()
	}

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		return s.String()
	}

	if m.topOutput == "" {
		s.WriteString("No process information available.\n")
	} else {
		// Display the raw top output
		lines := strings.Split(m.topOutput, "\n")
		visibleHeight := m.height - 5 // Account for title and help

		for i, line := range lines {
			if i >= visibleHeight {
				break
			}
			s.WriteString(line + "\n")
		}
	}

	s.WriteString("\n")

	// Show help hint
	s.WriteString(helpStyle.Render("Press ? for help"))

	return s.String()
}
