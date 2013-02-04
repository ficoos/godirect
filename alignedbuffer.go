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

// A buffer that is aligned in memory. This only applies if it is created with
// NewAlignedBuffer(). Aligned buffers are not managed by Go's memory manager
// and thus has to be freed manually.
type AlignedBuffer struct {
	ptr  uintptr
	size int
}

// Creates a new aligned buffer.
func NewAlignedBuffer(alignment int64, size int) (*AlignedBuffer, error) {
	// Allocate memory
	var ptr unsafe.Pointer
	pbuff := &ptr
	rv := C.posix_memalign(pbuff, (C.size_t)(alignment), (C.size_t)(size))
	if rv < 0 {
		err := (syscall.Errno)(rv)
		return nil, err
	}

	// Create a slice on top of allocated memory
	return &AlignedBuffer{uintptr(ptr), size}, nil
}

// Frees all the non Go managed memory used by the buffer
func (m *AlignedBuffer) Close() {
	C.free(unsafe.Pointer(m.ptr))
}

// Creates a slice backed by the buffer. Be careful, when the buffer is freed
// all slices will point to invalid memory.
func (m *AlignedBuffer) CreateSlice() []byte {
	var slice []byte
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	sliceHeader.Cap = m.size
	sliceHeader.Len = m.size
	sliceHeader.Data = uintptr(m.ptr)
	return slice
}
