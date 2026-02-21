# Test Improvement Tracking

## Phase 1 - Custom rendering views (highest risk)

- [x] `view_top.go` - Long string unit tests, small height edge cases
- [x] `view_stats.go` - Long NET I/O / BLOCK I/O values, narrow terminal
- [x] `view_inspect.go` - Long JSON lines, small height rendering
- [x] `view_file_content.go` - Long content lines, small viewport

## Phase 2 - List views using RenderTable (medium risk)

- [ ] `view_docker_container_list.go` - Long Image/Ports/Names fields
- [ ] `view_compose_process_list.go` - Long field values
- [ ] `view_image_list.go` - Long repository names
- [ ] `view_network_list.go` - Long network names
- [ ] `view_volume_list.go` - Long volume names (UI rendering tests)

## Phase 3 - Special views

- [ ] `view_file_browser.go` - Long filenames, responsive columns
- [ ] `view_dind_process_list.go` - Same as container list + DinD
- [ ] `view_compose_project_list.go` - Long ConfigFiles paths
- [ ] `view_command_execution.go` - Viewport, confirmation dialog

## Phase 4 - Docker client functions

- [ ] `internal/docker/` - ListContainers, ListImages, ListNetworks, ListComposeProjects, GetStats
