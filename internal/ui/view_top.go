package ui

import (
	"strings"
)

func (m *Model) renderTopView(availableHeight int) string {
	var s strings.Builder

	if m.topOutput == "" {
		s.WriteString("No process information available.\n")
	} else {
		// Display the raw top output
		lines := strings.Split(m.topOutput, "\n")
		visibleHeight := availableHeight

		for i, line := range lines {
			if i >= visibleHeight {
				break
			}
			s.WriteString(line + "\n")
		}
	}

	return s.String()
}
