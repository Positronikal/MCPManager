// +build windows

package discovery

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows API constants
const (
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
	MAX_PATH                  = 260
	TH32CS_SNAPPROCESS        = 0x00000002
)

// ProcessBasicInformation structure for NtQueryInformationProcess
type PROCESS_BASIC_INFORMATION struct {
	ExitStatus                   uintptr
	PebBaseAddress               uintptr
	AffinityMask                 uintptr
	BasePriority                 uintptr
	UniqueProcessId              uintptr
	InheritedFromUniqueProcessId uintptr
}

// PEB (Process Environment Block) structure - partial definition
type PEB struct {
	Reserved1              [2]byte
	BeingDebugged          byte
	Reserved2              [1]byte
	Reserved3              [2]uintptr
	Ldr                    uintptr
	ProcessParameters      uintptr
	Reserved4              [3]uintptr
	AtlThunkSListPtr       uintptr
	Reserved5              uintptr
	Reserved6              uint32
	Reserved7              uintptr
	Reserved8              uint32
	AtlThunkSListPtr32     uint32
	Reserved9              [45]uintptr
	Reserved10             [96]byte
	PostProcessInitRoutine uintptr
	Reserved11             [128]byte
	Reserved12             [1]uintptr
	SessionId              uint32
}

// RTL_USER_PROCESS_PARAMETERS structure - partial definition
type RTL_USER_PROCESS_PARAMETERS struct {
	Reserved1     [16]byte
	Reserved2     [10]uintptr
	ImagePathName UNICODE_STRING
	CommandLine   UNICODE_STRING
}

// UNICODE_STRING structure
type UNICODE_STRING struct {
	Length        uint16
	MaximumLength uint16
	Buffer        uintptr
}

// PROCESSENTRY32 structure for toolhelp snapshot
type PROCESSENTRY32 struct {
	Size              uint32
	CntUsage          uint32
	ProcessID         uint32
	DefaultHeapID     uintptr
	ModuleID          uint32
	CntThreads        uint32
	ParentProcessID   uint32
	PriorityClassBase int32
	Flags             uint32
	ExeFile           [MAX_PATH]uint16
}

var (
	modNtdll    = windows.NewLazySystemDLL("ntdll.dll")
	modKernel32 = windows.NewLazySystemDLL("kernel32.dll")
	modPsapi    = windows.NewLazySystemDLL("psapi.dll")

	procNtQueryInformationProcess = modNtdll.NewProc("NtQueryInformationProcess")
	procReadProcessMemory         = modKernel32.NewProc("ReadProcessMemory")
	procCreateToolhelp32Snapshot  = modKernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First            = modKernel32.NewProc("Process32FirstW")
	procProcess32Next             = modKernel32.NewProc("Process32NextW")
	procEnumProcesses             = modPsapi.NewProc("EnumProcesses")
	procGetModuleFileNameEx       = modPsapi.NewProc("GetModuleFileNameExW")
)

// listProcessesWindows enumerates all processes using native Win32 API
func (pd *ProcessDiscovery) listProcessesWindows() ([]ProcessInfo, error) {
	// First, get snapshot of all processes to build parent-child relationships
	snapshot, err := createProcessSnapshot()
	if err != nil {
		return nil, fmt.Errorf("failed to create process snapshot: %w", err)
	}
	defer windows.CloseHandle(windows.Handle(snapshot))

	var processes []ProcessInfo

	// Get first process
	var pe PROCESSENTRY32
	pe.Size = uint32(unsafe.Sizeof(pe))

	ret, _, _ := procProcess32First.Call(uintptr(snapshot), uintptr(unsafe.Pointer(&pe)))
	if ret == 0 {
		return nil, fmt.Errorf("Process32First failed")
	}

	// Iterate through all processes
	for {
		pid := pe.ProcessID
		parentPID := pe.ParentProcessID
		exeName := windows.UTF16ToString(pe.ExeFile[:])

		// Skip system idle process
		if pid == 0 {
			goto next
		}

		// Filter to relevant processes for MCP servers
		// We're interested in: node.exe, python.exe, uv.exe, go.exe, claude.exe
		if !isRelevantProcess(exeName) {
			goto next
		}

		// Try to get more detailed information
		{
			procInfo, err := getProcessDetails(pid, parentPID, exeName)
			if err != nil {
				// If we can't get details, at least add basic info
				processes = append(processes, ProcessInfo{
					PID:         int(pid),
					ParentPID:   int(parentPID),
					Name:        exeName,
					CommandLine: exeName,
				})
			} else {
				processes = append(processes, procInfo)
			}
		}

	next:
		// Get next process
		ret, _, _ = procProcess32Next.Call(uintptr(snapshot), uintptr(unsafe.Pointer(&pe)))
		if ret == 0 {
			break
		}
	}

	return processes, nil
}

// createProcessSnapshot creates a snapshot of all running processes
func createProcessSnapshot() (syscall.Handle, error) {
	ret, _, err := procCreateToolhelp32Snapshot.Call(
		uintptr(TH32CS_SNAPPROCESS),
		0,
	)
	if ret == uintptr(syscall.InvalidHandle) {
		return syscall.InvalidHandle, fmt.Errorf("CreateToolhelp32Snapshot failed: %w", err)
	}
	return syscall.Handle(ret), nil
}

// isRelevantProcess checks if a process is relevant for MCP server detection
func isRelevantProcess(name string) bool {
	relevantNames := []string{
		"node.exe",
		"python.exe",
		"python3.exe",
		"uv.exe",
		"go.exe",
		"claude.exe",
	}

	nameLower := toLower(name)
	for _, relevant := range relevantNames {
		if nameLower == toLower(relevant) {
			return true
		}
	}
	return false
}

// getProcessDetails gets detailed information about a process including command line
func getProcessDetails(pid, parentPID uint32, exeName string) (ProcessInfo, error) {
	// Open process with query and read permissions
	handle, err := windows.OpenProcess(
		PROCESS_QUERY_INFORMATION|PROCESS_VM_READ,
		false,
		pid,
	)
	if err != nil {
		return ProcessInfo{}, fmt.Errorf("OpenProcess failed: %w", err)
	}
	defer windows.CloseHandle(handle)

	// Get executable path
	var exePath [MAX_PATH]uint16
	size := uint32(MAX_PATH)
	ret, _, _ := procGetModuleFileNameEx.Call(
		uintptr(handle),
		0,
		uintptr(unsafe.Pointer(&exePath[0])),
		uintptr(size),
	)

	executablePath := exeName
	if ret > 0 {
		executablePath = windows.UTF16ToString(exePath[:])
	}

	// Get command line from PEB
	commandLine, err := getProcessCommandLine(handle)
	if err != nil {
		// If we can't get command line, use executable path as fallback
		commandLine = executablePath
	}

	return ProcessInfo{
		PID:         int(pid),
		ParentPID:   int(parentPID),
		Name:        exeName,
		CommandLine: commandLine,
	}, nil
}

// getProcessCommandLine extracts the full command line from a process's PEB
func getProcessCommandLine(handle windows.Handle) (string, error) {
	// Query process basic information to get PEB address
	var pbi PROCESS_BASIC_INFORMATION
	var returnLength uint32

	ret, _, _ := procNtQueryInformationProcess.Call(
		uintptr(handle),
		0, // ProcessBasicInformation
		uintptr(unsafe.Pointer(&pbi)),
		uintptr(unsafe.Sizeof(pbi)),
		uintptr(unsafe.Pointer(&returnLength)),
	)

	if ret != 0 {
		return "", fmt.Errorf("NtQueryInformationProcess failed with status: 0x%x", ret)
	}

	// Read PEB structure from process memory
	var peb PEB
	err := readProcessMemory(handle, pbi.PebBaseAddress, unsafe.Pointer(&peb), unsafe.Sizeof(peb))
	if err != nil {
		return "", fmt.Errorf("failed to read PEB: %w", err)
	}

	// Read RTL_USER_PROCESS_PARAMETERS structure
	var params RTL_USER_PROCESS_PARAMETERS
	err = readProcessMemory(handle, peb.ProcessParameters, unsafe.Pointer(&params), unsafe.Sizeof(params))
	if err != nil {
		return "", fmt.Errorf("failed to read process parameters: %w", err)
	}

	// Read the actual command line string
	if params.CommandLine.Length == 0 || params.CommandLine.Buffer == 0 {
		return "", fmt.Errorf("command line is empty")
	}

	// Allocate buffer for command line
	cmdLineLength := params.CommandLine.Length / 2 // Length is in bytes, we need characters
	if cmdLineLength > 32768 {                     // Sanity check
		cmdLineLength = 32768
	}

	cmdLineBuffer := make([]uint16, cmdLineLength+1)
	err = readProcessMemory(
		handle,
		params.CommandLine.Buffer,
		unsafe.Pointer(&cmdLineBuffer[0]),
		uintptr(params.CommandLine.Length),
	)
	if err != nil {
		return "", fmt.Errorf("failed to read command line string: %w", err)
	}

	return windows.UTF16ToString(cmdLineBuffer), nil
}

// readProcessMemory reads memory from another process
func readProcessMemory(handle windows.Handle, baseAddress uintptr, buffer unsafe.Pointer, size uintptr) error {
	var bytesRead uintptr
	ret, _, err := procReadProcessMemory.Call(
		uintptr(handle),
		baseAddress,
		uintptr(buffer),
		size,
		uintptr(unsafe.Pointer(&bytesRead)),
	)

	if ret == 0 {
		return fmt.Errorf("ReadProcessMemory failed: %w", err)
	}

	if bytesRead != size {
		return fmt.Errorf("ReadProcessMemory: expected %d bytes, read %d", size, bytesRead)
	}

	return nil
}

// findClaudeDesktopChildren finds all child processes of Claude Desktop
func findClaudeDesktopChildren() ([]uint32, error) {
	// First, find Claude.exe PID
	snapshot, err := createProcessSnapshot()
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(windows.Handle(snapshot))

	var claudePID uint32
	var pe PROCESSENTRY32
	pe.Size = uint32(unsafe.Sizeof(pe))

	ret, _, _ := procProcess32First.Call(uintptr(snapshot), uintptr(unsafe.Pointer(&pe)))
	if ret == 0 {
		return nil, fmt.Errorf("Process32First failed")
	}

	// Find Claude.exe
	for {
		exeName := windows.UTF16ToString(pe.ExeFile[:])
		if toLower(exeName) == "claude.exe" {
			claudePID = pe.ProcessID
			break
		}

		ret, _, _ = procProcess32Next.Call(uintptr(snapshot), uintptr(unsafe.Pointer(&pe)))
		if ret == 0 {
			break
		}
	}

	if claudePID == 0 {
		return nil, fmt.Errorf("Claude.exe not found")
	}

	// Now find all children of Claude.exe
	var children []uint32

	// Reset snapshot
	snapshot2, err := createProcessSnapshot()
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(windows.Handle(snapshot2))

	pe.Size = uint32(unsafe.Sizeof(pe))
	ret, _, _ = procProcess32First.Call(uintptr(snapshot2), uintptr(unsafe.Pointer(&pe)))
	if ret == 0 {
		return nil, fmt.Errorf("Process32First failed")
	}

	for {
		if pe.ParentProcessID == claudePID {
			children = append(children, pe.ProcessID)
		}

		ret, _, _ = procProcess32Next.Call(uintptr(snapshot2), uintptr(unsafe.Pointer(&pe)))
		if ret == 0 {
			break
		}
	}

	return children, nil
}

// toLower is a simple ASCII lowercase function to avoid string imports
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		b[i] = c
	}
	return string(b)
}
