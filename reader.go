package godirect

import (
	"os"
	"syscall"
)

// A high level wrapper that allows reading direct IO in arbitrary locations
// and chunk sizes. This implementation prefers simplicity over performance.
// It is meant to be easily integrated with other IO operations in Go that
// might not expect the constraints that direct IO might demand.
type Reader struct {
	file  File  // managed file
	xfer  int64 // recommended transfer size
	align int64 // recommended alignment
}

// Creates a new Reader for that specified file
// It is assume that the file has been opened with the O_DIRECT flag
func NewReader(file File) *Reader {
	var align, xfer int64 = 4096, 4096
	topo := GetTopologyData(file)
	if topo.AlignmentOffset > 0 {
		align = int64(topo.AlignmentOffset)
	}
	if topo.OptimalIOSize > 0 {
		xfer = int64(topo.OptimalIOSize)
	} else if topo.MinimumIOSize > 0 {
		xfer = int64(topo.MinimumIOSize)
	}

	return &Reader{file, xfer, align}
}

// Reads bytes from the file.
// Note that the movement of the offset pointer is not atomic.
func (r *Reader) Read(p []byte) (n int, err error) {
	fpointer, e := r.tell()
	if e != nil {
		return 0, nil
	}
	n, err = r.ReadAt(p, fpointer)
	// TBD: Should I do error handling here?
	r.file.Seek(int64(len(p)), os.SEEK_CUR)
	return n, err

}

// Gets the current offset of the managed file
func (r *Reader) tell() (offset int64, err error) {
	return r.file.Seek(0, os.SEEK_CUR)
}

func (r *Reader) ReadAt(p []byte, off int64) (int, error) {
	//TODO: check if buffer is already conformant in case user wants to
	//      optimize
	l := len(p)
	start := off

	// offset from start of block
	offset := int(start % r.xfer)
	start -= int64(offset)
	l += offset

	// Make sure you read complete blocks
	remainder := int64(l) % r.xfer
	l += int(r.xfer - remainder)

	buff, err := NewAlignedBuffer(r.align, l)
	if err != nil {
		return -1, err
	}
	defer buff.Close()

	slice := buff.CreateSlice()

	bread, err := syscall.Pread(int(r.file.Fd()), slice, start)
	if err != nil {
		return -1, err
	}

	bread -= offset
	if bread > len(p) {
		bread = len(p)
	}

	copy(p, slice[offset:offset+bread])
	return bread, nil
}

func (r *Reader) ReadByte() (c byte, err error) {
	buff := make([]byte, 1)
	_, err = r.Read(buff)
	return buff[0], err
}
