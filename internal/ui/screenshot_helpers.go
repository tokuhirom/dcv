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
}

func (m *Model) GetImageListViewModel() *ImageListViewModel {
	return &m.imageListViewModel
}

func (vm *ImageListViewModel) SetImages(images []models.DockerImage) {
	vm.dockerImages = images
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
