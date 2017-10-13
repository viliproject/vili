package docker

import (
	"sort"
	"time"
)

var service Service

// Service is a docker service instance that fetches images from a repository
type Service interface {
	GetRepository(repo string, branches []string) ([]*Image, error)
	GetTag(repo, tag string) (string, error)
	FullName(repo, tag string) (string, error)
}

// GetRepository returns the images in the given repository for the provided branch names
func GetRepository(repo string, branches []string) ([]*Image, error) {
	return service.GetRepository(repo, branches)
}

// GetTag returns an image digest for the given tag
func GetTag(repo, tag string) (string, error) {
	return service.GetTag(repo, tag)
}

// FullName returns the complete docker image name
func FullName(repo, tag string) (string, error) {
	return service.FullName(repo, tag)
}

// Image represents a docker image in a repository
type Image struct {
	Tag          string    `json:"tag"`
	Branch       string    `json:"branch"`
	Revision     string    `json:"revision"`
	LastModified time.Time `json:"lastModified"`
}

type getImagesResult struct {
	images []*Image
	err    error
}

// imageSorter joins a By function and a slice of Images to be sorted.
type imageSorter struct {
	images []*Image
	by     func(i1, i2 *Image) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *imageSorter) Len() int {
	return len(s.images)
}

// Swap is part of sort.Interface.
func (s *imageSorter) Swap(i, j int) {
	s.images[i], s.images[j] = s.images[j], s.images[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *imageSorter) Less(i, j int) bool {
	return s.by(s.images[i], s.images[j])
}

func sortByLastModified(images []*Image) {
	ps := &imageSorter{
		images: images,
		by: func(i1, i2 *Image) bool {
			return i1.LastModified.After(i2.LastModified)
		},
	}
	sort.Sort(ps)
}

// NotFoundError is raised when a given repository or image tag is not found
type NotFoundError struct {
}

func (e *NotFoundError) Error() string {
	return "Docker image not found"
}
