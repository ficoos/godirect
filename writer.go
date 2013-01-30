package godirect

// This is a high level implementation that wraps direct IO. It will buffer
// data until enough data has been written to the buffer to allow a clean
// direct write. To force a write the user can use Flush() but the writer will
// fill will complete the buffer to a full chunk with '\0'
type BufferedWriter struct {
	file     File  // managed file
	xfer     int64 // recommended transfer size
	align    int64 // recommended alignment
	fpointer int64 // current location in the file
}

type Writer struct {
}
