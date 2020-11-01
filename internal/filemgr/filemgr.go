package filemgr

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func GetFileDescriptorsResourceLimit() (uint64, uint64, error) {
	var rscLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rscLimit); err != nil {
		return 0, 0, fmt.Errorf("getrlimit failed: %w", err)
	}
	return rscLimit.Cur, rscLimit.Max, nil
}

func GetFileDescriptersKernelLimit() (int, error) {
	content, err := ioutil.ReadFile("/proc/sys/fs/nr_open")
	if err != nil {
		return 0, fmt.Errorf("failed to read nr_open: %w", err)
	}
	num, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil {
		return 0, fmt.Errorf("failed to interpret nr_open file content: %w", err)
	}
	return num, nil
}

func RepeatedlyOpen(filePath string) (int, error) {
	fds := make([]*os.File, 0)
	defer func() {
		for _, fd := range fds {
			if err := fd.Close(); err != nil {
				log.Printf("failed to close FD: %v", err)
			}
		}
	}()
	for i := 0; ; i++ {
		fd, err := os.Open(filePath)
		if err != nil {
			return i, err
		}
		fds = append(fds, fd)
	}
}
