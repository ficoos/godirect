package godirect

import (
	"bytes"
	"syscall"
)

// Writer implements buffering on top of a specially allocated buffer to allow
// for direct IO writes. If an error occurs writing to a Writer, no more data
// will be accepted and all subsequent writes will return the error. The user
// can force a write by using Flush() but the writer will fill the ramainder of
// the buffer with '\0' before writing.
type Writer struct {
	file  File          // managed file
	xfer  int64         // recommended transfer size
	align int64         // recommended alignment
	abuff AlignedBuffer // internal buffer
	buff  []byte        // slice to internal buffer
	pbuff int           // current location in internall buffer
	err   error         // error that broke this Writer
}

// NewWriter returns a new Writer with an internal buffer that is suitable for
// for direct IO operations.
func NewWriter(file File) (*Writer, error) {
	var align, xfer int64 = 4096, 4096
	topo := DetectDeviceTopology(file)
	if topo.AlignmentOffset > 0 {
		align = int64(topo.AlignmentOffset)
	}
	// TBD: Always use minimal?
	if topo.OptimalIOSize > 0 {
		xfer = int64(topo.OptimalIOSize)
	} else if topo.MinimumIOSize > 0 {
		xfer = int64(topo.MinimumIOSize)
	}

	buff, err := NewAlignedBuffer(align, int(xfer))
	if err != nil {
		return nil, err
	}
	pbuff := 0

	w := &Writer{file, xfer, align, buff, buff.CreateSlice(), pbuff, nil}

	return w, nil
}

// Write writes the contents of p into the buffer. It returns the number of
// bytes written. If nn < len(p), it also returns an error explaining why the
// write is short.
func (w *Writer) Write(p []byte) (nn int, err error) {
	if w.err != nil {
		return 0, w.err
	}

	var bcopied int
	var n int
	nn = 0

	for len(p) > 0 {
		bcopied = copy(w.buff[w.pbuff:], p)
		if bcopied < len(p) {
			p = p[bcopied:]
			n, w.err = syscall.Write(int(w.file.Fd()), w.buff)
			nn += n
			if w.err != nil {
				return nn, w.err
			}

			w.pbuff = 0
		} else { // Buffer not full
			w.pbuff += bcopied
		}
	}

	nn += bcopied
	return nn, nil
}

// Available returns how many bytes are unused in the buffer.
func (w *Writer) Available() int {
	return len(w.buff) - w.pbuff
}

// Buffered returns the number of bytes that have been written into the current
// buffer.
func (w *Writer) Buffered() int {
	return w.pbuff
}

// Flushes all buffered data. If there is not enough data to write a complete
// block the function will fill the remaining space with '\0'.
func (w *Writer) Flush() error {
	if w.Buffered() == 0 {
		return nil // Nothing to do
	}

	_, err := w.Write(bytes.Repeat([]byte{0}, w.Available()))
	return err
}

// Frees the internal buffers but does not close the underlying file.
func (w *Writer) Close() {
	w.abuff.Close()
}
