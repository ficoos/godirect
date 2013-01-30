package godirect

type File interface {
	Fd() uintptr
	Seek(offset int64, whence int) (ret int64, err error)
}
