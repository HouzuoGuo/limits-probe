package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/HouzuoGuo/limits-probe/internal/filemgr"
	"github.com/HouzuoGuo/limits-probe/internal/tcpserver"
)

type ExperimentNames []string

func (names *ExperimentNames) String() string {
	return "not-used"
}

func (names *ExperimentNames) Set(name string) error {
	*names = append(*names, name)
	return nil
}

func main() {
	var experimentNames ExperimentNames
	flag.Var(&experimentNames, "ex", `Name of the experiment ("file" - try to open max number of FDs, "sock" - try to establish max number of TCP connections, "mem" - try to allocate plenty of memory, "exec" - try to start plenty of external processes). This flag may be used more than once.`)
	flag.Parse()

	if len(experimentNames) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	for i, name := range experimentNames {
		switch name {
		case "file":
			OpenFileExperiment()
		case "sock":
			SocketConnectionExperiment()
		case "mem":
			MemoryAllocationExperiment()
		case "exec":
			ExecExperiment()
		default:
			flag.Usage()
			os.Exit(1)
		}
		if i != len(experimentNames)-1 {
			log.Print("Sleeping 10 seconds in between experiments")
			time.Sleep(10 * time.Second)
		}
	}
}

func OpenFileExperiment() {
	nrFileSoftLimit, nrFileHardLimit, err := filemgr.GetFileDescriptorsResourceLimit()
	if err != nil {
		log.Fatal(err)
	}
	nrFileKernelLimit, err := filemgr.GetFileDescriptersKernelLimit()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Max number of open files: soft limit %d, hard limit %d, kernel limit %d", nrFileSoftLimit, nrFileHardLimit, nrFileKernelLimit)

	file, err := ioutil.TempFile("", "github-HouzuoGuo-limits-probe")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	defer os.Remove(file.Name())
	nrSuccess, ultimateErr := filemgr.RepeatedlyOpen(file.Name())
	log.Printf("Successfully opened %d FDs and then encountered failure: %v", nrSuccess, ultimateErr)
}

func SocketConnectionExperiment() {
	server := &tcpserver.Server{Addr: "localhost", Port: 0}
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
	go func() {
		if err := server.Serve(); err != nil {
			log.Printf("server is quitting: %v", err)
		}
	}()
	defer server.Shutdown()
	nrSuccess, ultimateErr := tcpserver.RepeatedlyConnect(server.Addr, server.GetListenerPort())
	log.Printf("Successfully made %d TCP connections and then encountered failure: %v", nrSuccess, ultimateErr)
}

func MemoryAllocationExperiment() {
	bufs := make([][]byte, 0)
	for i := 0; ; i++ {
		// Allocate 100 MB of memory
		buf := make([]byte, 1024*1024*100)
		// Touch the memory to ensure they truly are allocated
		for i := range buf {
			buf[i] = 1
		}
		bufs = append(bufs, buf)
		log.Printf("Allocated %d MB of memory", len(bufs)*100)
	}
}

func ExecExperiment() {
	cmds := make([]*exec.Cmd, 0)
	defer func() {
		for _, cmd := range cmds {
			_ = cmd.Process.Kill()
		}
	}()
	for i := 0; ; i++ {
		cmd := exec.Command("/usr/bin/sleep", "100")
		if err := cmd.Start(); err != nil {
			log.Printf("Successfully made %d external processes and then encountered failure: %v", i, err)
			return
		}
		cmds = append(cmds, cmd)
		if i%10 == 0 {
			log.Printf("Spawned %d external processes", len(cmds))
		}
	}
}
