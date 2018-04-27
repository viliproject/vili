package repository

var dockerService DockerService

// DockerService is a docker service instance that fetches images from a repository
type DockerService interface {
	GetRepository(repo string, branches []string) ([]*Image, error)
	GetTag(repo, tag string) (string, error)
	FullName(repo, tag string) (string, error)
}

// GetDockerRepository returns the images in the given repository for the provided branch names
func GetDockerRepository(repo string, branches []string) ([]*Image, error) {
	return dockerService.GetRepository(repo, branches)
}

// GetDockerTag returns an image digest for the given tag
func GetDockerTag(repo, tag string) (string, error) {
	return dockerService.GetTag(repo, tag)
}

// DockerFullName returns the complete docker image name
func DockerFullName(repo, tag string) (string, error) {
	return dockerService.FullName(repo, tag)
}
