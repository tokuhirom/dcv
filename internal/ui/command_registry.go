package ui

import (
	"reflect"
	"runtime"
	"strings"
)

// CommandHandler represents a function that can be executed via command line
type CommandHandler struct {
	Handler     KeyHandler
	Description string
	// TODO: Following comment is completely wrong, ViewMask is not a bitwise mask.
	ViewMask ViewType // Bitwise mask of views where this command is available (0 = all views)
}

// initCommandRegistry initializes the command registry with all available commands
func (m *Model) initCommandRegistry() {
	m.commandRegistry = make(map[string]CommandHandler)
	m.handlerToCommand = make(map[uintptr]string)

	// Register all key handlers as commands
	m.registerCommands()
}

// registerCommands registers all key handler functions as commands
func (m *Model) registerCommands() {
	// Get all handlers from all views
	allHandlers := []struct {
		handlers []KeyConfig
		viewMask ViewType
	}{
		{m.composeProcessListViewHandlers, ComposeProcessListView},
		{m.logViewHandlers, LogView},
		{m.dindListViewHandlers, DindProcessListView},
		{m.topViewHandlers, TopView},
		{m.statsViewHandlers, StatsView},
		{m.composeProjectListViewHandlers, ComposeProjectListView},
		{m.dockerContainerListViewHandlers, DockerContainerListView},
		{m.imageListViewHandlers, ImageListView},
		{m.networkListViewHandlers, NetworkListView},
		{m.volumeListViewHandlers, VolumeListView},
		{m.fileBrowserHandlers, FileBrowserView},
		{m.fileContentHandlers, FileContentView},
		{m.inspectViewHandlers, InspectView},
		{m.helpViewHandlers, HelpView},
	}

	// Track which handlers we've already registered to avoid duplicates
	registered := make(map[uintptr]bool)

	for _, viewHandlers := range allHandlers {
		for _, handler := range viewHandlers.handlers {
			// Get the function pointer
			funcPtr := reflect.ValueOf(handler.KeyHandler).Pointer()

			// Skip if already registered
			if registered[funcPtr] {
				continue
			}
			registered[funcPtr] = true

			// Get the function name
			funcName := runtime.FuncForPC(funcPtr).Name()
			// Extract just the method name (e.g., "SelectUpContainer" from "github.com/tokuhirom/dcv/internal/ui.(*Model).SelectUpContainer-fm")
			parts := strings.Split(funcName, ".")
			if len(parts) > 0 {
				methodName := parts[len(parts)-1]
				// Remove the -fm suffix if present (from method value)
				methodName = strings.TrimSuffix(methodName, "-fm")
				// Remove the (*Model) part if present
				methodName = strings.TrimPrefix(methodName, "(*Model)")
				methodName = strings.TrimPrefix(methodName, ")")
				methodName = strings.TrimPrefix(methodName, "Cmd")

				// Otherwise, use kebab-case version
				cmdName := toKebabCase(methodName)
				m.commandRegistry[cmdName] = CommandHandler{
					Handler:     handler.KeyHandler,
					Description: handler.Description,
					ViewMask:    viewHandlers.viewMask,
				}
				m.handlerToCommand[funcPtr] = cmdName
			}
		}
	}

	// Add additional aliases for common commands
	aliases := map[string]string{
		"select":  "log",     // Alternative for entering log view
		"enter":   "log",     // Alternative for entering log view
		"delete":  "remove",  // Alternative for remove
		"rm":      "remove",  // Short for remove
		"logs":    "log",     // Alternative for log
		"exec":    "shell",   // Alternative for shell
		"unpause": "pause",   // Same as pause (it's a toggle)
		"q":       "quit",    // Short for quit
		"h":       "help",    // Short for help
		"r":       "refresh", // Already registered but good to have as alias
		"a":       "all",     // Short for toggle all
	}

	// Register aliases
	for alias, target := range aliases {
		if cmd, exists := m.commandRegistry[target]; exists {
			m.commandRegistry[alias] = cmd
		}
	}
}

// toKebabCase converts a CamelCase string to kebab-case
func toKebabCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('-')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// getAvailableCommands returns a list of commands available in the current view
func (m *Model) getAvailableCommands() []string {
	var commands []string
	for cmdName, cmd := range m.commandRegistry {
		if cmd.ViewMask == 0 || cmd.ViewMask == m.currentView {
			commands = append(commands, cmdName)
		}
	}
	return commands
}

// getCommandSuggestions returns command suggestions based on partial input
func (m *Model) getCommandSuggestions(partial string) []string {
	var suggestions []string
	availableCommands := m.getAvailableCommands()

	for _, cmdName := range availableCommands {
		if strings.HasPrefix(cmdName, partial) {
			suggestions = append(suggestions, cmdName)
		}
	}

	// If no prefix matches, try substring matching
	if len(suggestions) == 0 {
		for _, cmdName := range availableCommands {
			if strings.Contains(cmdName, partial) {
				suggestions = append(suggestions, cmdName)
			}
		}
	}

	return suggestions
}

// getCommandForHandler returns the command name for a given key handler
func (m *Model) getCommandForHandler(handler KeyHandler) string {
	if handler == nil {
		return ""
	}

	funcPtr := reflect.ValueOf(handler).Pointer()
	if cmdName, exists := m.handlerToCommand[funcPtr]; exists {
		return cmdName
	}

	return ""
}
