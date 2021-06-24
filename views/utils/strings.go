package utils

func StringContains(sl []string, in string) bool {
	for _, v := range sl {
		if v == in {
			return true
		}
	}

	return false
}
