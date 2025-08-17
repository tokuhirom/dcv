# tview Testing Documentation

## Overview

This document describes the testing strategy and implementation for the tview-based UI in DCV.

## Test Structure

### Test Files
- `test_helpers.go` - Helper functions and test utilities
- `state_test.go` - State management unit tests  
- `app_test.go` - Application-level integration tests
- `views/docker_container_list_test.go` - View component tests
- `mocks.go` - Mock implementations for testing

## Testing Approach

### 1. Unit Testing
Tests individual components in isolation:
- State management functions
- View data transformations
- Business logic separated from UI

### 2. Component Testing
Tests view components without full application context:
- Table rendering
- Data updates
- Key handler setup

### 3. Integration Testing
Tests interactions between components:
- View switching
- Navigation history
- Global keyboard shortcuts

## Key Testing Patterns

### Mock Implementations
```go
// Mock Docker client for testing
type MockDockerClient struct {
    Containers []models.DockerContainer
    ListContainersCalled int
}
```

### Test Helpers
```go
// Create test application with simulation screen
func NewTestApplication() (*tview.Application, tcell.SimulationScreen)

// Create test data fixtures
func CreateTestContainers() []models.DockerContainer
```

### Table-Driven Tests
```go
tests := []struct {
    name     string
    input    ui.ViewType
    expected string
}{
    // test cases
}
```

## Running Tests

```bash
# Run all tview tests
go test ./internal/tui/...

# Run with verbose output
go test ./internal/tui/... -v

# Run specific test
go test ./internal/tui/views -run TestDockerContainerListView

# Run with coverage
go test ./internal/tui/... -cover
```

## Testing Challenges & Solutions

### Challenge: tview Run() blocks
**Solution**: Test components directly without running the full application event loop

### Challenge: SimulationScreen limitations
**Solution**: Focus on testing business logic and component behavior rather than full UI rendering

### Challenge: Async operations
**Solution**: Use channels and goroutines carefully in tests, with timeouts for safety

## Test Coverage Areas

### ✅ Covered
- State management (push/pop views, error handling)
- View initialization and configuration
- Table data updates
- Mock Docker operations
- View switching logic
- Navigation history

### 🔄 Partially Covered
- Keyboard event handling (structure verified, not full simulation)
- Color and styling (text content verified, not visual appearance)

### ❌ Not Covered
- Full application run cycle (requires terminal)
- Complex user interactions
- Real Docker operations

## Best Practices

1. **Isolate UI from Business Logic**: Keep Docker operations and data processing testable independently
2. **Use Interfaces**: Define interfaces for views and operations to enable mocking
3. **Test Data First**: Verify data transformations before testing UI rendering
4. **Keep Tests Fast**: Avoid real Docker calls and terminal operations in unit tests
5. **Document Limitations**: Be clear about what aspects of UI behavior can't be tested automatically

## Manual Testing Checklist

For aspects that can't be automatically tested:

- [ ] Application starts without errors
- [ ] View navigation with number keys (1-6)
- [ ] Container operations (start/stop/kill/delete)
- [ ] Search functionality (/)
- [ ] Refresh functionality (r)
- [ ] Help view display (?)
- [ ] Quit confirmation (q)
- [ ] ESC for back navigation
- [ ] Visual appearance and colors
- [ ] Performance with large datasets

## Future Improvements

1. **Screenshot Testing**: Capture and compare terminal output for regression testing
2. **E2E Testing Framework**: Build a framework for automated terminal interaction testing
3. **Performance Benchmarks**: Add benchmarks for table rendering with large datasets
4. **Mock Server**: Create a mock Docker daemon for more realistic integration tests