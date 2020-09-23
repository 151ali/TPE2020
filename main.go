package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/vishvananda/netlink"
)

/*
go build main.go
./main run <command>
*/

const ipAdrr = "10.1.1.1/24" // TODO

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("bad command")
	}
}

func run() {

	fmt.Printf("Parent : Running %v\n", os.Args[2:])

	cmd := exec.Command("/proc/self/exe",
		append([]string{"child"}, os.Args[2:]...)...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWUSER |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET,

		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
	}

	if err := cmd.Start(); err != nil {
		fmt.Println("failed to start the command: ", err)
		os.Exit(1)
	}

	// !
	fmt.Println("(PID)", cmd.Process.Pid)

	// cgroup conf :
	pid := fmt.Sprintf("%d", cmd.Process.Pid)

	cg := exec.Command("binaires/cgroup/cg", pid)
	if err := cg.Run(); err != nil {
		fmt.Printf("Error running netsetgo - %s\n", err)
		os.Exit(1)
	}
	// network conf :
	netw := exec.Command("binaires/network/netw", pid)
	if err := netw.Run(); err != nil {
		fmt.Printf("Error running netsetgo - %s\n", err)
		os.Exit(1)
	}
	//
	//
	//
	//
	if err := cmd.Wait(); err != nil {
		fmt.Printf("Error running the exec.Command - %s\n", err)

	}
}
func child() {
	fmt.Printf("Child : Running %v\n", os.Args[2:])
	// set hostname
	syscall.Sethostname([]byte("container"))
	// mount new rootfs
	syscall.Chroot("alpine") // see : $ man 2 chroot
	syscall.Chdir("/")       // see : $ man 2 chdir

	// hna
	if err := syscall.Mount("proc", "proc", "proc", 0, ""); err != nil {
		fmt.Println("mount said that : ", err)
	}
	// network configs:
	vth, err := waitforVContainer()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	err = netlink.LinkSetUp(vth)
	if err != nil {
		fmt.Println(err)
	}

	lo, err := netlink.LinkByName("lo")
	if err != nil {
		fmt.Println(err)
	}
	if err := netlink.LinkSetUp(lo); err != nil {
		fmt.Println("failed de setup lo :", err)
	}

	addr, err := netlink.ParseAddr(ipAdrr)
	if err != nil {
		fmt.Println("can't parse addr :", err)
	}

	netlink.AddrAdd(vth, addr)

	// ***

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	syscall.Unmount("proc", 1) // see : $ man 2 umount

	cmd.Run()
}

// ===========================================================================
func waitforVContainer() (netlink.Link, error) {
	s := time.Now()
	fmt.Println("waiting for interfaces ")
	// waiting loop
	for {
		fmt.Print("*")
		if time.Since(s) > 5*time.Second {
			return nil, fmt.Errorf("waiting timeout ended")
		}

		//check if a veth exist
		vlist, err := netlink.LinkList()
		if err != nil {
			return nil, err
		}

		for _, v := range vlist {
			if v.Type() == "veth" {
				fmt.Printf("Done!\n\n")
				return v, nil
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

}
