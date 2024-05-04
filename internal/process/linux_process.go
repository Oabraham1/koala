package process

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

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
