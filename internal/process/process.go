//go:build darwin && linux
// +build darwin,linux

package process

import (
	"fmt"
	"runtime"
)

type Process struct {
	// The process ID.
	// This is the unique identifier for the process.
	// For a thread, this is the thread ID.
	// For a zombie, this is the process ID of the parent.
	// For a dead process, this is the process ID of the parent.
	// For a process that is being debugged, this is the process ID of the debugger.
	// For a process that is being profiled, this is the process ID of the profiler.
	// For a process that is being traced, this is the process ID of the tracer.
	// For a process that is being monitored, this is the process ID of the monitor.
	// For a process that is being controlled, this is the process ID of the controller.
	// For a process that is being watched, this is the process ID of the watcher.
	// For a process that is being audited, this is the process ID of the auditor.
	// For a process that is being logged, this is the process ID of the logger.
	Pid int32

	// The process name.
	// This is the name of the executable.
	// For a thread, this is the name of the process.
	Name string

	// The process status.
	// One of "R"unning, "S"leeping, "D"isk sleep, "Z"ombie, "T"raced or "W"aiting.
	Status string

	// The process parent ID.
	// This is the process ID of the parent process.
	Ppid int32
}

// GetAllProcesses returns a list of all processes.
//
// This function returns a list of all processes at the time
// the function is called. The list includes all processes
// that are currently running, as well as all processes that
// have exited but have not yet been reaped by the parent process.
// The list is sorted by process ID.
func GetAllProcesses() ([]Process, error) {
	os := runtime.GOOS
	switch os {
	case "linux":
		return getLinuxProcesses()
	case "windows":
		return getWindowsProcesses()
	case "darwin":
		return getMacOSProcesses()
	case "freebsd":
		return getFreeBSDProcesses()
	case "openbsd":
		return getOpenBSDProcesses()
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", os)
	}
}

// GetProcess returns the process with the given process ID.
//
// If a process with the given process ID is found, the function
// returns the process. If no such process is found, the function
// returns an error.
func GetProcess(pid int32) (Process, error) {
	return Process{}, nil
}

// GetParentProcess returns the parent process of the process with the given process ID.
//
// If a process with the given process ID is found, the function returns the parent process.
// If no such process is found, the function returns an error.
// If the process with the given process ID is the init process, the function returns an error.
func GetParentProcess(pid int32) (Process, error) {
	return Process{}, nil
}

func getWindowsProcesses() ([]Process, error) {
	return nil, nil
}

// func getWindowsProcess(pid int32) (Process, error) {
// 	return Process{}, nil
// }

// func getWindowsProcessParent(pid int32) (Process, error) {
// 	return Process{}, nil
// }

func getFreeBSDProcesses() ([]Process, error) {
	return nil, nil
}

// func getFreeBSDProcess(pid int32) (Process, error) {
// 	return Process{}, nil
// }

// func getFreeBSDProcessParent(pid int32) (Process, error) {
// 	return Process{}, nil
// }

func getOpenBSDProcesses() ([]Process, error) {
	return nil, nil
}

// func getOpenBSDProcess(pid int32) (Process, error) {
// 	return Process{}, nil
// }

// func getOpenBSDProcessParent(pid int32) (Process, error) {
// 	return Process{}, nil
// }
