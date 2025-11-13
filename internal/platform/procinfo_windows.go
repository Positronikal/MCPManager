//go:build windows

package platform

import (
	"fmt"
	"syscall"
	"unsafe"
)

// DefaultProcessInfo implements ProcessInfo for Windows systems
type DefaultProcessInfo struct{}

// NewProcessInfo creates a new ProcessInfo for the current platform
func NewProcessInfo() ProcessInfo {
	return &DefaultProcessInfo{}
}

// Windows API structures and constants
var (
	kernel32                   = syscall.NewLazyDLL("kernel32.dll")
	psapi                      = syscall.NewLazyDLL("psapi.dll")
	procOpenProcess            = kernel32.NewProc("OpenProcess")
	procCloseHandle            = kernel32.NewProc("CloseHandle")
	procGetProcessMemoryInfo   = psapi.NewProc("GetProcessMemoryInfo")
)

const (
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
)

// PROCESS_MEMORY_COUNTERS structure for Windows API
type PROCESS_MEMORY_COUNTERS struct {
	CB                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

// GetMemoryUsage returns the memory usage of a process in bytes
// Uses Windows API (psapi.dll) to query process memory information
func (pi *DefaultProcessInfo) GetMemoryUsage(pid int) (uint64, error) {
	// Open process handle
	handle, _, err := procOpenProcess.Call(
		uintptr(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ),
		uintptr(0),
		uintptr(pid),
	)
	if handle == 0 {
		return 0, fmt.Errorf("failed to open process: %v", err)
	}
	defer procCloseHandle.Call(handle)

	// Get memory info
	var memCounters PROCESS_MEMORY_COUNTERS
	memCounters.CB = uint32(unsafe.Sizeof(memCounters))

	ret, _, err := procGetProcessMemoryInfo.Call(
		handle,
		uintptr(unsafe.Pointer(&memCounters)),
		uintptr(memCounters.CB),
	)
	if ret == 0 {
		return 0, fmt.Errorf("failed to get process memory info: %v", err)
	}

	// Return working set size (equivalent to RSS on Unix)
	return uint64(memCounters.WorkingSetSize), nil
}
