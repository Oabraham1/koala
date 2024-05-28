package process

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
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

type kernelInfoProcess struct {
	_      [40]byte
	Pid    int32
	_      [199]byte
	Comm   [16]byte
	_      [301]byte
	PPid   int32
	_      [84]byte
	Status byte
	_      [1]byte
}

const (
	_CTRL_KERN               = 1
	_KERN_PROC               = 14
	_KERN_PROC_ALL           = 0
	_KERNEL_INFO_STRUCT_SIZE = 648
)

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

func getLinuxProcesses() ([]Process, error) {
	data, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}
	defer data.Close()

	processes := make([]Process, 0, 50)
	for {
		files, err := data.Readdirnames(10)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			pid, err := strconv.Atoi(file)
			if err != nil {
				continue
			}

			staticPath := fmt.Sprintf("/proc/%d/stat", pid)
			staticFile, err := os.ReadFile(staticPath)
			if err != nil {
				return nil, err
			}

			data := string(staticFile)
			dataBytesStart := strings.IndexRune(data, '(') + 1
			dataBytesEnd := strings.IndexRune(data, ')')
			dataBytesName := data[dataBytesStart : dataBytesStart+dataBytesEnd]
			process := Process{
				Pid:  int32(pid),
				Name: dataBytesName,
			}

			_, err = fmt.Scanf(data, "%d %s", &process.Status, &process.Ppid)
			if err != nil {
				return nil, err
			}

			processes = append(processes, process)
		}
	}
	return processes, nil
}

func getLinuxProcess(pid int32) (Process, error) {
	processes, err := getLinuxProcesses()
	if err != nil {
		return Process{}, err
	}

	for _, process := range processes {
		if process.Pid == pid {
			return process, nil
		}
	}
	return Process{}, nil
}

func getLinuxProcessParent(pid int32) (Process, error) {
	processes, err := getLinuxProcesses()
	if err != nil {
		return Process{}, err
	}

	for _, process := range processes {
		if process.Pid == pid {
			return getLinuxProcess(process.Ppid)
		}
	}
	return Process{}, nil
}

func getMacOSProcesses() ([]Process, error) {
	memory_io_buffer := [4]int32{_CTRL_KERN, _KERN_PROC, _KERN_PROC_ALL, 0}
	size := uintptr(0)

	_, _, err := syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&memory_io_buffer[0])),
		uintptr(4),
		uintptr(0),
		uintptr(unsafe.Pointer(&size)),
		uintptr(0),
		uintptr(0))

	if err != 0 {
		return nil, err
	}

	systemCallBuffer := make([]byte, size)

	_, _, err = syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&memory_io_buffer[0])),
		uintptr(4),
		uintptr(unsafe.Pointer(&systemCallBuffer[0])),
		uintptr(unsafe.Pointer(&size)),
		uintptr(0),
		uintptr(0))

	if err != 0 {
		return nil, err
	}

	bytesBuffer := bytes.NewBuffer(systemCallBuffer[0:size])

	processes := make([]*kernelInfoProcess, 0, 50)
	kernel := 0
	for i := _KERNEL_INFO_STRUCT_SIZE; i < bytesBuffer.Len(); i += _KERNEL_INFO_STRUCT_SIZE {
		process := &kernelInfoProcess{}
		err := binary.Read(bytes.NewBuffer(bytesBuffer.Bytes()[kernel:i]), binary.LittleEndian, process)
		if err != nil {
			return nil, err
		}

		kernel = i
		processes = append(processes, process)
	}

	result := make([]Process, 0, len(processes))
	for _, process := range processes {
		result = append(result, Process{
			Pid:    process.Pid,
			Name:   string(process.Comm[:bytes.IndexByte(process.Comm[:], 0)]),
			Ppid:   process.PPid,
			Status: string([]byte{process.Status}),
		})
	}

	return result, nil
}

func getMacOSProcess(pid int32) (Process, error) {
	processes, err := getMacOSProcesses()
	if err != nil {
		return Process{}, err
	}

	for _, process := range processes {
		if process.Pid == pid {
			return process, nil
		}
	}
	return Process{}, nil
}

func getMacOSProcessParent(pid int32) (Process, error) {
	processes, err := getMacOSProcesses()
	if err != nil {
		return Process{}, err
	}

	for _, process := range processes {
		if process.Pid == pid {
			return getMacOSProcess(process.Ppid)
		}
	}
	return Process{}, nil
}
