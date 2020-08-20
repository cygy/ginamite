package timezone

func hasValue(slice []string, value string) bool {
	for _, v := range slice {
		if value == v {
			return true
		}
	}

	return false
}
