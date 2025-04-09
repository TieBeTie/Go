package gitfame

func isHash(str string) bool {
	if len(str) != 40 {
		return false
	}
	for _, c := range str {
		if ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') {
			continue
		}
		return false
	}
	return true
}
