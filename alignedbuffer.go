package godirect

/*
#include <stdlib.h>
*/
import "C"

import (
	"reflect"
	"syscall"
	"unsafe"
)

type AlignedBuffer interface {
	Close()
	CreateSlice() []byte
}

// A buffer that is aligned in memory. This only applies if it is created with
// NewAlignedBuffer(). Aligned buffers are not managed by Go's memory manager
// and thus has to be freed manually.
type _AlignedBuffer struct {
	ptr  uintptr
	size int
}

// The function allocates size bytes. The address of the allocated memory will
// be a multiple of alignment, which must be a power of two and a multiple of
// sizeof(void *). Alignment or size must not be 0 or the method will
// intentionally panic.
func NewAlignedBuffer(alignment int64, size int) (AlignedBuffer, error) {
	if alignment == 0 || size == 0 {
		panic(error(syscall.EINVAL))
	}

	// Allocate memory
	var ptr unsafe.Pointer
	pbuff := &ptr
	rv := C.posix_memalign(pbuff, (C.size_t)(alignment), (C.size_t)(size))
	if rv < 0 {
		err := (syscall.Errno)(rv)
		return nil, err
	}

	// Create a slice on top of allocated memory
	return &_AlignedBuffer{uintptr(ptr), size}, nil
}

// Frees all the non Go managed memory used by the buffer
func (m *_AlignedBuffer) Close() {
	C.free(unsafe.Pointer(m.ptr))
}

// Creates a slice backed by the buffer. Be careful, when the buffer is freed
// all slices will point to invalid memory.
func (m *_AlignedBuffer) CreateSlice() []byte {
	var slice []byte
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	sliceHeader.Cap = m.size
	sliceHeader.Len = m.size
	sliceHeader.Data = uintptr(m.ptr)
	return slice
}
