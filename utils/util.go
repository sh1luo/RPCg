package utils

import "unsafe"

func CopyMap(dst, src map[string]string) {
	if dst == nil {
		return
	}
	for k, v := range src {
		dst[k] = v
	}
}

// SliceByteToString reserved to use
//go:nosplit
func SliceByteToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StringToSliceByte reserved to use
//go:nosplit
func StringToSliceByte(s string) []byte {
	ss := (*[2]uintptr)(unsafe.Pointer(&s))
	b := [3]uintptr{ss[0],ss[1],ss[1]}
	return *(*[]byte)(unsafe.Pointer(&b))
}