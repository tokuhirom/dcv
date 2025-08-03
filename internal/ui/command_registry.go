package ui

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// CommandHandler represents a function that can be executed via command line
type CommandHandler struct {
	Handler     KeyHandler
	Description string
	ViewMask    ViewType // Bitwise mask of views where this command is available (0 = all views)
}

// commandRegistry maps command names to their handlers
var commandRegistry map[string]CommandHandler

// handlerToCommand maps handler function pointers to command names
var handlerToCommand map[uintptr]string

// initCommandRegistry initializes the command registry with all available commands
func (m *Model) initCommandRegistry() {
	commandRegistry = make(map[string]CommandHandler)
	handlerToCommand = make(map[uintptr]string)

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
		{m.processListViewHandlers, ComposeProcessListView},
		{m.logViewHandlers, LogView},
		{m.dindListViewHandlers, DindComposeProcessListView},
		{m.topViewHandlers, TopView},
		{m.statsViewHandlers, StatsView},
		{m.projectListViewHandlers, ProjectListView},
		{m.dockerListViewHandlers, DockerContainerListView},
		{m.imageListViewHandlers, ImageListView},
		{m.networkListViewHandlers, NetworkListView},
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
				
				// Convert to kebab-case command name (e.g., "SelectUpContainer" -> "select-up-container")
				cmdName := toKebabCase(methodName)
				
				commandRegistry[cmdName] = CommandHandler{
					Handler:     handler.KeyHandler,
					Description: handler.Description,
					ViewMask:    viewHandlers.viewMask,
				}
				
				// Also map the handler to command name for help display
				handlerToCommand[funcPtr] = cmdName
			}
		}
	}

	// Add some view-agnostic aliases for common commands
	aliases := map[string]string{
		"up":     "select-up-container",
		"down":   "select-down-container",
		"select": "show-compose-log",
		"enter":  "show-compose-log",
		"back":   "back-to-process-list",
		"refresh": "refresh-process-list",
		"kill":   "kill-container",
		"stop":   "stop-container",
		"start":  "up-service",
		"restart": "restart-container",
		"delete": "delete-container",
		"rm":     "delete-container",
		"logs":   "show-compose-log",
		"top":    "show-top-view",
		"stats":  "show-stats-view",
		"images": "show-image-list",
		"networks": "show-network-list",
		"projects": "show-project-list",
		"ps":     "show-docker-container-list",
		"inspect": "show-inspect",
		"exec":   "execute-shell",
		"files":  "show-file-browser",
		"pause":  "pause-container",
		"unpause": "pause-container", // Toggle
	}

	// Register aliases
	for alias, target := range aliases {
		if cmd, exists := commandRegistry[target]; exists {
			commandRegistry[alias] = cmd
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

// executeKeyHandlerCommand executes a command by name
func (m *Model) executeKeyHandlerCommand(cmdName string) (tea.Model, tea.Cmd) {
	cmd, exists := commandRegistry[cmdName]
	if !exists {
		m.err = fmt.Errorf("unknown command: %s", cmdName)
		return m, nil
	}

	// Check if command is available in current view
	if cmd.ViewMask != 0 && cmd.ViewMask != m.currentView {
		// Try to find a similar command for the current view
		currentViewCmd := m.findCommandForCurrentView(cmdName)
		if currentViewCmd != nil {
			return currentViewCmd.Handler(tea.KeyMsg{})
		}
		m.err = fmt.Errorf("command '%s' is not available in current view", cmdName)
		return m, nil
	}

	// Execute the command
	return cmd.Handler(tea.KeyMsg{})
}

// findCommandForCurrentView tries to find a similar command for the current view
func (m *Model) findCommandForCurrentView(baseCmdName string) *CommandHandler {
	// Common command patterns that might have view-specific variants
	patterns := []string{
		"select-up-",
		"select-down-",
		"refresh-",
		"back-from-",
		"show-",
	}

	for _, pattern := range patterns {
		if strings.HasPrefix(baseCmdName, pattern) {
			// Try to find a view-specific version
			for cmdName, cmd := range commandRegistry {
				if strings.HasPrefix(cmdName, pattern) && 
				   (cmd.ViewMask == 0 || cmd.ViewMask == m.currentView) {
					return &cmd
				}
			}
		}
	}

	return nil
}

// getAvailableCommands returns a list of commands available in the current view
func (m *Model) getAvailableCommands() []string {
	var commands []string
	for cmdName, cmd := range commandRegistry {
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
func getCommandForHandler(handler KeyHandler) string {
	if handler == nil {
		return ""
	}
	
	funcPtr := reflect.ValueOf(handler).Pointer()
	if cmdName, exists := handlerToCommand[funcPtr]; exists {
		return cmdName
	}
	
	return ""
}