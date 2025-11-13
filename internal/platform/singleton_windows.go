//go:build windows

package platform

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procCreateMutex         = kernel32.NewProc("CreateMutexW")
	procReleaseMutex        = kernel32.NewProc("ReleaseMutex")
	procFindWindow          = user32.NewProc("FindWindowW")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procShowWindow          = user32.NewProc("ShowWindow")
)

const (
	ERROR_ALREADY_EXISTS = 183
	SW_RESTORE           = 9
)

// WindowsSingleInstance implements SingleInstance for Windows using named mutex
type WindowsSingleInstance struct {
	mutexHandle syscall.Handle
	mutexName   string
	windowTitle string
}

// NewSingleInstance creates a new Windows single instance enforcer
func NewSingleInstance(appName string, windowTitle string) SingleInstance {
	return &WindowsSingleInstance{
		mutexName:   fmt.Sprintf("Global\\%s", appName),
		windowTitle: windowTitle,
	}
}

// Acquire attempts to acquire the single instance lock
func (w *WindowsSingleInstance) Acquire() (bool, error) {
	mutexNamePtr, err := syscall.UTF16PtrFromString(w.mutexName)
	if err != nil {
		return false, fmt.Errorf("failed to convert mutex name to UTF16: %w", err)
	}

	// Try to create the mutex
	ret, _, err := procCreateMutex.Call(
		0,                          // lpMutexAttributes (NULL)
		0,                          // bInitialOwner (FALSE)
		uintptr(unsafe.Pointer(mutexNamePtr)), // lpName
	)

	if ret == 0 {
		return false, fmt.Errorf("failed to create mutex: %w", err)
	}

	w.mutexHandle = syscall.Handle(ret)

	// Check if mutex already existed
	if err == syscall.Errno(ERROR_ALREADY_EXISTS) {
		// Another instance is running - bring it to foreground
		w.bringExistingWindowToFront()
		// Close the mutex handle since we won't be using it
		procCloseHandle.Call(uintptr(w.mutexHandle))
		w.mutexHandle = 0
		return false, nil
	}

	// We successfully acquired the mutex (first instance)
	return true, nil
}

// Release releases the single instance lock
func (w *WindowsSingleInstance) Release() error {
	if w.mutexHandle == 0 {
		return nil // Already released or never acquired
	}

	// Release the mutex
	ret, _, err := procReleaseMutex.Call(uintptr(w.mutexHandle))
	if ret == 0 {
		return fmt.Errorf("failed to release mutex: %w", err)
	}

	// Close the mutex handle
	ret, _, err = procCloseHandle.Call(uintptr(w.mutexHandle))
	if ret == 0 {
		return fmt.Errorf("failed to close mutex handle: %w", err)
	}

	w.mutexHandle = 0
	return nil
}

// bringExistingWindowToFront finds the existing application window and brings it to foreground
func (w *WindowsSingleInstance) bringExistingWindowToFront() {
	if w.windowTitle == "" {
		return
	}

	windowTitlePtr, err := syscall.UTF16PtrFromString(w.windowTitle)
	if err != nil {
		return
	}

	// Find the window by title
	hwnd, _, _ := procFindWindow.Call(
		0, // lpClassName (NULL - search by title only)
		uintptr(unsafe.Pointer(windowTitlePtr)),
	)

	if hwnd == 0 {
		// Window not found - might have different title or not created yet
		return
	}

	// Restore the window if it's minimized
	procShowWindow.Call(hwnd, SW_RESTORE)

	// Bring the window to foreground
	procSetForegroundWindow.Call(hwnd)
}
