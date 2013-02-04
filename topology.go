package godirect

/*
#cgo LDFLAGS: -lblkid
#include <unistd.h>
#include <blkid/blkid.h>
*/
import "C"

import (
	"fmt"
	"syscall"
)

type DeviceTopology struct {
	AlignmentOffset    uint64 // Meomory alignment required direct IO
	MinimumIOSize      uint64 // Minimal IO allowed
	OptimalIOSize      uint64 // Optimal IO for device
	LogicalSectorSize  uint64 // Logical sector size
	PhysicalSectorSize uint64 // Physical sector size
}

// Detects the topology data of a device. In case of an error getting any of
// the topology properties the field will be set to 0
func DetectDeviceTopology(file File) DeviceTopology {
	topo, err := getBlockDeviceTopology(file)
	if err == nil {
		return topo
	}

	return getFileSystemTopology(file)
}

// Get topology data using blkid APIs. In case of an error getting any of the
// topology properties the field will be set to 0
func getBlockDeviceTopology(file File) (DeviceTopology, error) {
	var res DeviceTopology

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
func getFileSystemTopology(file File) DeviceTopology {
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

	return DeviceTopology{ualigne, uxfer, uxfer, uxfer, uxfer}
}

func getRecommendedTransferAlignment(file File) int64 {
	rv := C.fpathconf((C.int)(file.Fd()), C._PC_REC_XFER_ALIGN)
	return int64(rv)
}

func getMinimumTransferSize(file File) int64 {
	rv := C.fpathconf((C.int)(file.Fd()), C._PC_REC_MIN_XFER_SIZE)
	return int64(rv)
}
