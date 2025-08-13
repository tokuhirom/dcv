//go:build screenshots
// +build screenshots

package ui

import (
	"strings"

	"github.com/tokuhirom/dcv/internal/docker"
	"github.com/tokuhirom/dcv/internal/models"
)

// Helper methods for screenshot generation
// These are only available when building with the "screenshots" build tag

func (m *Model) SetLoading(loading bool) {
	m.loading = loading
}

func (m *Model) GetComposeProcessListViewModel() *ComposeProcessListViewModel {
	return &m.composeProcessListViewModel
}

func (vm *ComposeProcessListViewModel) SetComposeContainers(containers []models.ComposeContainer) {
	vm.composeContainers = containers
}

func (vm *ComposeProcessListViewModel) SetProjectName(name string) {
	vm.projectName = name
}

func (m *Model) GetDockerContainerListViewModel() *DockerContainerListViewModel {
	return &m.dockerContainerListViewModel
}

func (vm *DockerContainerListViewModel) SetDockerContainers(containers []models.DockerContainer) {
	vm.dockerContainers = containers
	// Build rows for table rendering
	vm.Rows = vm.buildRows()
	// Set End to show all rows initially
	if len(vm.Rows) > 0 {
		vm.End = len(vm.Rows)
	}
}

func (m *Model) GetImageListViewModel() *ImageListViewModel {
	return &m.imageListViewModel
}

func (vm *ImageListViewModel) SetImages(images []models.DockerImage) {
	vm.dockerImages = images
	// Build rows for table rendering
	vm.Rows = vm.buildRows()
	// Set End to show all rows initially
	if len(vm.Rows) > 0 {
		vm.End = len(vm.Rows)
	}
}

func (m *Model) GetNetworkListViewModel() *NetworkListViewModel {
	return &m.networkListViewModel
}

func (vm *NetworkListViewModel) SetNetworks(networks []models.DockerNetwork) {
	vm.dockerNetworks = networks
}

func (m *Model) GetVolumeListViewModel() *VolumeListViewModel {
	return &m.volumeListViewModel
}

func (vm *VolumeListViewModel) SetVolumes(volumes []models.DockerVolume) {
	vm.dockerVolumes = volumes
}

func (m *Model) GetComposeProjectListViewModel() *ComposeProjectListViewModel {
	return &m.composeProjectListViewModel
}

func (vm *ComposeProjectListViewModel) SetProjects(projects []models.ComposeProject) {
	vm.projects = projects
	// Build rows for table rendering
	vm.Rows = vm.buildRows()
	// Set End to show all rows initially
	if len(vm.Rows) > 0 {
		vm.End = len(vm.Rows)
	}
}

func (m *Model) GetLogViewModel() *LogViewModel {
	return &m.logViewModel
}

func (vm *LogViewModel) SetLogContent(content string) {
	vm.logs = strings.Split(content, "\n")
}

func (vm *LogViewModel) SetContainer(container *docker.Container) {
	vm.container = container
}

func (m *Model) GetStatsViewModel() *StatsViewModel {
	return &m.statsViewModel
}

func (vm *StatsViewModel) SetStats(stats []models.ContainerStats) {
	vm.stats = stats
}

func (m *Model) GetTopViewModel() *TopViewModel {
	return &m.topViewModel
}

func (vm *TopViewModel) SetProcesses(processes []models.Process) {
	vm.processes = processes
}

func (vm *TopViewModel) SetContainer(container *docker.Container) {
	vm.container = container
}

func (m *Model) GetDindProcessListViewModel() *DindProcessListViewModel {
	return &m.dindProcessListViewModel
}

func (vm *DindProcessListViewModel) SetDindContainers(containers []models.DockerContainer) {
	vm.dindContainers = containers
}

func (vm *DindProcessListViewModel) SetHostContainer(container *docker.Container) {
	vm.hostContainer = container
}

func (m *Model) GetFileBrowserViewModel() *FileBrowserViewModel {
	return &m.fileBrowserViewModel
}

func (vm *FileBrowserViewModel) SetContainerFiles(files []models.ContainerFile) {
	vm.containerFiles = files
}

func (vm *FileBrowserViewModel) SetBrowsingContainer(container *docker.Container) {
	vm.browsingContainer = container
}

func (vm *FileBrowserViewModel) SetCurrentPath(path string) {
	vm.currentPath = path
}

func (m *Model) GetFileContentViewModel() *FileContentViewModel {
	return &m.fileContentViewModel
}

func (vm *FileContentViewModel) SetContent(content string) {
	vm.content = content
}

func (vm *FileContentViewModel) SetContainer(container *docker.Container) {
	vm.container = container
}

func (vm *FileContentViewModel) SetContentPath(path string) {
	vm.contentPath = path
}

func (m *Model) GetInspectViewModel() *InspectViewModel {
	return &m.inspectViewModel
}

func (vm *InspectViewModel) SetInspectContent(content string) {
	vm.inspectContent = content
}

func (vm *InspectViewModel) SetInspectTargetName(name string) {
	vm.inspectTargetName = name
}
