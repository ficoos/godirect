package godirect

/*
#cgo LDFLAGS: -lblkid
#include <stdlib.h>
#include <unistd.h>
#include <blkid/blkid.h>
*/
import "C"

import (
	"fmt"
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

type TopologyData struct {
	AlignmentOffset    uint64 // Meomory alignment required direct IO
	MinimumIOSize      uint64 // Minimal IO allowed
	OptimalIOSize      uint64 // Optimal IO for device
	LogicalSectorSize  uint64 // Logical sector size
	PhysicalSectorSize uint64 // Physical sector size
}

// Gets the topology data of a device. In case of an error getting any of the
// topology properties the field will be set to 0
func GetTopologyData(file File) TopologyData {
	topo, err := GetDeviceTopologyData(file)
	if err == nil {
		return topo
	}

	return GetFileSystemTopologyData(file)
}

// Get topology data using blkid APIs. In case of an error getting any of the
// topology properties the field will be set to 0
func GetDeviceTopologyData(file File) (TopologyData, error) {
	var res TopologyData

	probe := C.blkid_new_probe()
	if probe == nil {
		return res, fmt.Errorf("Could not probe device")
	}
	C.blkid_reset_probe(probe)

	fd, err := syscall.Dup(int(file.Fd()))
	if err != nil {
		return res, fmt.Errorf("Could not dup FD as part of device probing: %s", err)
	}

	rv := C.blkid_probe_set_device(probe, C.int(fd), 0, 0)
	if rv != 0 {
		syscall.Close(fd)
	}

	defer C.blkid_free_probe(probe)

	topology := C.blkid_probe_get_topology(probe)
	if topology == nil {
		return res, fmt.Errorf("Could not get topology for device")
	}

	res.AlignmentOffset = uint64(C.blkid_topology_get_alignment_offset(topology))
	res.MinimumIOSize = uint64(C.blkid_topology_get_minimum_io_size(topology))
	res.OptimalIOSize = uint64(C.blkid_topology_get_optimal_io_size(topology))
	res.LogicalSectorSize = uint64(C.blkid_topology_get_logical_sector_size(topology))
	res.PhysicalSectorSize = uint64(C.blkid_topology_get_physical_sector_size(topology))
	return res, nil
}

// Get topology data using file system APIs. In case of an error getting any of
// the topology properties the field will be set to 0
func GetFileSystemTopologyData(file File) TopologyData {
	xfer := getMinimumTransferSize(file)
	align := getRecommendedTransferAlignment(file)
	if xfer < 0 {
		xfer = 0
	}

	if align < 0 {
		xfer = 0
	}

	ualigne := uint64(align)
	uxfer := uint64(xfer)

	return TopologyData{ualigne, uxfer, uxfer, uxfer, uxfer}
}

func getRecommendedTransferAlignment(file File) int64 {
	rv := C.fpathconf((C.int)(file.Fd()), C._PC_REC_XFER_ALIGN)
	return int64(rv)
}

func getMinimumTransferSize(file File) int64 {
	rv := C.fpathconf((C.int)(file.Fd()), C._PC_REC_MIN_XFER_SIZE)
	return int64(rv)
}
