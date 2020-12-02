package utils

func CopyMap(dst, src map[string]string) {
	if dst == nil {
		return
	}
	for k,v := range src {
		dst[k] = v
	}
}
