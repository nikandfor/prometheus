package prometheus

import "unsafe"

//go:noescape
//go:linkname memhash runtime.memhash
func memhash(p *byte, h uintptr, s int) uintptr

func strhash(s string, h uintptr) uintptr {
	p := unsafe.StringData(s)

	return memhash(p, h, len(s))
}
