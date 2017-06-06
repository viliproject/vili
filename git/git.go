package git

var service Service

// Service is a git service that allows for querying and retrieving content
// from a repository
type Service interface {
	Branches() ([]string, error)
	Contents(branch, path string) (string, error)
	List(branch, path string) ([]string, error)
}

// Branches returns a list of branches for the repository
func Branches() ([]string, error) {
	return service.Branches()
}

// Contents returns the contents of the file at the given path
func Contents(branch, path string) (string, error) {
	return service.Contents(branch, path)
}

// List returns a list of subpaths of the given directory path
func List(branch, path string) ([]string, error) {
	return service.List(branch, path)
}
