package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

func main()  {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("invalid cmd")
	}
}

func run()  {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())
	/*proc目录是所有进程的元数据存放地方
	  我们的二进制文件也会出现在这里
	  下面这行代码会在新创建的容器内执行child函数，
	  proc/self/exe是一个特殊的文件，包含当前可执行文件的内存映像。
	  换句话说，会让进程重新运行自己，但是传递child作为第一个参数。
	  这个可执行程序让我们能够执行另一个程序，执行一个由用户请求的程序（由‘os.Args[2:]’中定义的内容）。
	  基于这个简单的结构，我们就能够创建一个容器。*/
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	// 将操作系统标准io重定向到容器中
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 设置一些系统进程属性，下面这行代码负责创建一个新的独立进程
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// 创建进程或容器来运行我们提供的命令
		// CLONE_NEWUTS运行容器有独立的UTS
		// CLONE_NEWPID为新的命名空间进程提供pids
		// CLONE_NEWNS为mount提供新的命名空间
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		// systemd中的挂载会递归共享属性。
		//取消对新挂载命名空间的递归共享属性。
		//它阻止与主机共享新的命名空间。
		Unshareflags: syscall.CLONE_NEWNS,
	}

	// 运行命令并捕获错误
	if err := cmd.Run(); err != nil {
		log.Fatal("Error: ", err)
	}
}

func child()  {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

	cg()

	// 下面一行代码才是在我们自己创建的容器中执行用户命令
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	// 将io重定向到容器内
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//下面是一些设置容器属性的系统调用
	//为新创建命名空间设置主机名
	must(syscall.Sethostname([]byte("container")))
	// 为容器设置根目录
	must(syscall.Chroot("/home/jabin/test"))
	// 设置“/”作为默认目录
	must(syscall.Chdir("/"))
	// 挂载/proc目录查看在容器内运行的进程
	must(syscall.Mount("proc", "proc", "proc", 0, ""))



	// 运行命令并捕获错误
	if err := cmd.Run(); err != nil {
		log.Fatal("Error: ", err)
	}
	// 在命令完成后卸载proc
	syscall.Unmount("/proc", 0)
}

func cg() {
	cgroups := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")
	os.Mkdir(filepath.Join(pids, "jabin"), 0755)
	must(ioutil.WriteFile(filepath.Join(pids, "jabin/pids.max"), []byte("20"), 0700))
	// Removes the new cgroup in place after the container exits
	must(ioutil.WriteFile(filepath.Join(pids, "jabin/notify_on_release"), []byte("1"), 0700))
	must(ioutil.WriteFile(filepath.Join(pids, "jabin/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}


func must(err error) {
	if err != nil {
		panic(err)
	}
}