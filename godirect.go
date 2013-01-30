// Package to allow direct IO in Go.
// Currently only works on Linux
package godirect

type File interface {
	Fd() uintptr
	Seek(offset int64, whence int) (ret int64, err error)
}
