package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

// go run main.go run <cmd> <args>
func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("help")
	}
}

func run() {
	fmt.Printf("Running %v \n", os.Args[2:])

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	checkError(cmd.Run())
}

func child() {
	fmt.Printf("Running %v \n", os.Args[2:])

	cg()

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	checkError(syscall.Sethostname([]byte("container")))
	checkError(syscall.Chroot("/home/mauro/ubuntufs"))
	checkError(os.Chdir("/"))
	checkError(syscall.Mount("proc", "proc", "proc", 0, ""))
	checkError(syscall.Mount("thing", "mytemp", "tmpfs", 0, ""))

	checkError(cmd.Run())

	checkError(syscall.Unmount("proc", 0))
	checkError(syscall.Unmount("thing", 0))
}

func cg() {
	cgroups := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")
	os.Mkdir(filepath.Join(pids, "mauro"), 0755)
	checkError(ioutil.WriteFile(filepath.Join(pids, "mauro/pids.max"), []byte("20"), 0700))
	// Removes the new cgroup in place after the container exits
	checkError(ioutil.WriteFile(filepath.Join(pids, "mauro/notify_on_release"), []byte("1"), 0700))
	checkError(ioutil.WriteFile(filepath.Join(pids, "mauro/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
