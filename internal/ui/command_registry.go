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
// Multiple handlers can be registered for the same command name (e.g., "up" for different views)
var commandRegistry map[string][]CommandHandler

// handlerToCommand maps handler function pointers to command names
var handlerToCommand map[uintptr]string

// initCommandRegistry initializes the command registry with all available commands
func (m *Model) initCommandRegistry() {
	commandRegistry = make(map[string][]CommandHandler)
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

				// Get simplified command name
				cmdName := getSimplifiedCommandName(methodName)
				if cmdName == "" {
					// Fall back to kebab-case if no simplified name exists
					cmdName = toKebabCase(methodName)
				}

				// Append to the list of handlers for this command
				commandRegistry[cmdName] = append(commandRegistry[cmdName], CommandHandler{
					Handler:     handler.KeyHandler,
					Description: handler.Description,
					ViewMask:    viewHandlers.viewMask,
				})

				// Also map the handler to command name for help display
				handlerToCommand[funcPtr] = cmdName
			}
		}
	}

	// Add additional aliases where the simplified name might conflict or be ambiguous
	// Most commands now have their simplified names registered directly
	aliases := map[string]string{
		// Alternate names for common actions
		"quit":       "q",
		"quit!":      "q!",
		"list":       "ps",
		"containers": "ps",
		"remove":     "rm",
		"delete":     "rm",
		"unpause":    "pause", // Toggle action

		// Long form alternatives
		"docker-ps":    "ps",
		"compose-logs": "logs",
		"file-browser": "files",
	}

	// Register aliases
	for alias, target := range aliases {
		if handlers, exists := commandRegistry[target]; exists {
			commandRegistry[alias] = handlers
		}
	}
}

// getSimplifiedCommandName returns a simplified command name for common method names
func getSimplifiedCommandName(methodName string) string {
	// Direct mapping of method names to simplified command names
	simplifiedNames := map[string]string{
		// Navigation commands
		"SelectUpContainer":         "up",
		"SelectDownContainer":       "down",
		"SelectUpDindContainer":     "up",
		"SelectDownDindContainer":   "down",
		"SelectUpProject":           "up",
		"SelectDownProject":         "down",
		"SelectUpDockerContainer":   "up",
		"SelectDownDockerContainer": "down",
		"SelectUpImage":             "up",
		"SelectDownImage":           "down",
		"SelectUpNetwork":           "up",
		"SelectDownNetwork":         "down",
		"SelectUpVolume":            "up",
		"SelectDownVolume":          "down",
		"SelectUpFile":              "up",
		"SelectDownFile":            "down",

		// View switching commands
		"ShowDockerContainerList": "ps",
		"ShowComposeLog":          "logs",
		"ShowDindProcessList":     "dind",
		"ShowProjectList":         "projects",
		"ShowImageList":           "images",
		"ShowNetworkList":         "networks",
		"ShowVolumeList":          "volumes",
		"ShowFileBrowser":         "files",
		"ShowTopView":             "top",
		"ShowStatsView":           "stats",
		"ShowInspect":             "inspect",
		"ShowHelp":                "help",
		"ShowDindLog":             "logs",
		"ShowDockerLog":           "logs",
		"ShowDockerFileBrowser":   "files",
		"ShowDockerInspect":       "inspect",
		"ShowImageInspect":        "inspect",
		"ShowNetworkInspect":      "inspect",
		"ShowVolumeInspect":       "inspect",

		// Action commands
		"ExecuteShell":           "exec",
		"ExecuteDockerShell":     "exec",
		"KillContainer":          "kill",
		"KillDockerContainer":    "kill",
		"StopContainer":          "stop",
		"StopDockerContainer":    "stop",
		"StartDockerContainer":   "start",
		"RestartContainer":       "restart",
		"RestartDockerContainer": "restart",
		"DeleteContainer":        "rm",
		"DeleteDockerContainer":  "rm",
		"DeleteImage":            "rmi",
		"ForceDeleteImage":       "rmi-force",
		"DeleteNetwork":          "rm-network",
		"DeleteVolume":           "rm-volume",
		"ForceDeleteVolume":      "rm-volume-force",
		"PauseContainer":         "pause",
		"PauseDockerContainer":   "pause",

		// Refresh commands
		"RefreshProcessList": "refresh",
		"RefreshDindList":    "refresh",
		"RefreshTop":         "refresh",
		"RefreshStats":       "refresh",
		"RefreshProjects":    "refresh",
		"RefreshDockerList":  "refresh",
		"RefreshImageList":   "refresh",
		"RefreshNetworkList": "refresh",
		"RefreshVolumeList":  "refresh",
		"RefreshFiles":       "refresh",

		// Toggle commands
		"ToggleAllContainers":       "all",
		"ToggleAllDockerContainers": "all",
		"ToggleAllImages":           "all",

		// Back/navigation commands
		"BackFromLogView":     "back",
		"BackToDindList":      "back",
		"BackToProcessList":   "back",
		"BackFromDockerList":  "back",
		"BackFromImageList":   "back",
		"BackFromNetworkList": "back",
		"BackFromVolumeList":  "back",
		"BackFromFileBrowser": "back",
		"BackFromFileContent": "back",
		"BackFromInspect":     "back",
		"BackFromHelp":        "back",
		"BackFromCommandExec": "back",

		// Service/project commands
		"UpService":     "up-service",
		"DeployProject": "deploy",
		"DownProject":   "down",

		// Scroll commands
		"ScrollLogUp":           "scroll-up",
		"ScrollLogDown":         "scroll-down",
		"ScrollFileUp":          "scroll-up",
		"ScrollFileDown":        "scroll-down",
		"ScrollInspectUp":       "scroll-up",
		"ScrollInspectDown":     "scroll-down",
		"ScrollHelpUp":          "scroll-up",
		"ScrollHelpDown":        "scroll-down",
		"ScrollCommandExecUp":   "scroll-up",
		"ScrollCommandExecDown": "scroll-down",

		// Jump commands
		"GoToLogEnd":           "end",
		"GoToLogStart":         "start",
		"GoToFileEnd":          "end",
		"GoToFileStart":        "start",
		"GoToInspectEnd":       "end",
		"GoToInspectStart":     "start",
		"GoToCommandExecEnd":   "end",
		"GoToCommandExecStart": "start",

		// Search commands
		"StartSearch":             "search",
		"StartFilter":             "filter",
		"StartInspectSearch":      "search",
		"NextSearchResult":        "next",
		"PrevSearchResult":        "prev",
		"NextInspectSearchResult": "next",
		"PrevInspectSearchResult": "prev",

		// File browser commands
		"OpenFileOrDirectory": "open",
		"GoToParentDirectory": "parent",

		// Other commands
		"SelectProject":     "select",
		"CancelCommandExec": "cancel",
	}

	return simplifiedNames[methodName]
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
	// For simplified commands like "up", "down", etc., find the appropriate handler for current view
	if handler := m.findBestHandlerForCommand(cmdName); handler != nil {
		return handler.Handler(tea.KeyMsg{})
	}

	m.err = fmt.Errorf("unknown command: %s", cmdName)
	return m, nil
}

// findBestHandlerForCommand finds the best handler for a command in the current view
func (m *Model) findBestHandlerForCommand(cmdName string) *CommandHandler {
	handlers, exists := commandRegistry[cmdName]
	if !exists || len(handlers) == 0 {
		return nil
	}

	// If only one handler, use it
	if len(handlers) == 1 {
		return &handlers[0]
	}

	// Multiple handlers - prefer one matching current view
	for i := range handlers {
		if handlers[i].ViewMask == m.currentView {
			return &handlers[i]
		}
	}

	// No exact view match - use first one with no view mask (generic)
	for i := range handlers {
		if handlers[i].ViewMask == 0 {
			return &handlers[i]
		}
	}

	// Just return the first one
	return &handlers[0]
}

// getAvailableCommands returns a list of commands available in the current view
func (m *Model) getAvailableCommands() []string {
	commandSet := make(map[string]bool)

	for cmdName, handlers := range commandRegistry {
		for _, handler := range handlers {
			if handler.ViewMask == 0 || handler.ViewMask == m.currentView {
				commandSet[cmdName] = true
				break // No need to check other handlers for this command
			}
		}
	}

	var commands []string
	for cmdName := range commandSet {
		commands = append(commands, cmdName)
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
