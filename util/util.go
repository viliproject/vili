package util

// Contains checks if the given element `e` exists in given slice `s`
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
