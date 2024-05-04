package process

import (
	"bytes"
	"encoding/binary"
	"syscall"
	"unsafe"
)

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
