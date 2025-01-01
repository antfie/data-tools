package utils

func IsInArray(in string, array []string) bool {
	for _, item := range array {
		if item == in {
			return true
		}
	}

	return false
}
