package repository

var bundleService BundleService

// BundleService is a bundle service instance that fetches images from a repository
type BundleService interface {
	GetRepository(repo string, branches []string) ([]*Image, error)
	FullName(repo, tag string) (string, error)
}

// GetBundleRepository returns the images in the given repository for the provided branch names
func GetBundleRepository(repo string, branches []string) ([]*Image, error) {
	return bundleService.GetRepository(repo, branches)
}

// BundleFullName returns the complete bundle image name
func BundleFullName(repo, tag string) (string, error) {
	return bundleService.FullName(repo, tag)
}
