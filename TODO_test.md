# Test Improvement Tracking

## Phase 1 - Custom rendering views (highest risk)

- [x] `view_top.go` - Long string unit tests, small height edge cases
- [x] `view_stats.go` - Long NET I/O / BLOCK I/O values, narrow terminal
- [x] `view_inspect.go` - Long JSON lines, small height rendering
- [x] `view_file_content.go` - Long content lines, small viewport

## Phase 2 - List views using RenderTable (medium risk)

- [x] `view_docker_container_list.go` - Long Image/Ports/Names fields
- [x] `view_compose_process_list.go` - Long field values
- [x] `view_image_list.go` - Long repository names
- [x] `view_network_list.go` - Long network names
- [x] `view_volume_list.go` - Long volume names (UI rendering tests)

## Phase 3 - Special views

- [x] `view_file_browser.go` - Long filenames, responsive columns
- [x] `view_dind_process_list.go` - Same as container list + DinD
- [x] `view_compose_project_list.go` - Long ConfigFiles paths
- [x] `view_command_execution.go` - Viewport, confirmation dialog

## Phase 4 - Docker client functions

- [ ] `internal/docker/` - ListContainers, ListImages, ListNetworks, ListComposeProjects, GetStats
