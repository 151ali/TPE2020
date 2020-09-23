package main

/* notes :
$ go build cg.go
$ sudo chown root:root cg
$ sudo chmod u+s cg
*/

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) > 1 {
		p, err := strconv.Atoi(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		pid := p
		//fmt.Println(pid)
		cgroups(pid)
		//
		//
		//
	} else {
		fmt.Println("you should pass a PID as an argument !")
	}
}

// ===================================================================================
// cgroups :
func cgroups(pid int) {
	for _, c := range []string{"cpu", "memory", "pids"} { // TODO : add the rest of cgroups
		cpath := fmt.Sprintf("/sys/fs/cgroup/%s/container0/", c)
		if err := os.MkdirAll(cpath, 0755); err != nil {
			fmt.Println("failed to create cpu cgroup for my container: ", err)
			os.Exit(1)
		}
		addProcessToCgroup(cpath+"cgroup.procs", pid)
	}

	// TODO : setup the rest of configuration here
	pidspath := "/sys/fs/cgroup/pids/container0/"
	add(22, "pids.max", pidspath)
	add(1, "notify_on_release", pidspath)

	cpupath := "/sys/fs/cgroup/cpu/container0/"
	// limit the CPU hard to 0.5 cores
	add(10000, "cpu.cfs_period_us", cpupath)
	add(1000, "cpu.cfs_quota_us", cpupath)
	add(1, "notify_on_release", cpupath)

	mempath := "/sys/fs/cgroup/pids/container0/"
	// see more : https://www.kernel.org/doc/Documentation/cgroup-v1/memory.txt
	add(1024, "memory.limit_in_bytes", mempath) // set/show limit of memory usage
	add(1, "notify_on_release", mempath)

}

func addProcessToCgroup(filepath string, pid int) {
	file, err := os.OpenFile(filepath, os.O_WRONLY, 0755)
	if err != nil {
		//
		fmt.Println(":{")
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	if _, err := file.WriteString(fmt.Sprintf("%d", pid)); err != nil {
		fmt.Println("failed to setup cgroup for the container: ", err)
		os.Exit(1)
	}
}

func add(arg int, fileName string, fileDir string) {
	file, _ := os.OpenFile(fileDir+fileName, os.O_WRONLY, 0755)
	file.WriteString(fmt.Sprintf("%d", arg))
	file.Close()
}
