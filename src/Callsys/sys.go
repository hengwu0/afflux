package callsys

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

//初始化clock_tick
const Clktck = 1000000000 / 100 //man proc(5)说sysconf(_SC_CLK_TCK)大多为100

//初始化字节序
var ByteOrder myorder = func() (od myorder) {
	i := uint32(1)
	b := (*[4]byte)(unsafe.Pointer(&i))
	if b[0] == 1 {
		od.order = binary.LittleEndian
	} else {
		od.order = binary.BigEndian
	}
	return
}()

//需要根据rsp指针大小来确定32位还是64位程序，并用于dump数据
func GetPtrSize(rsp uint64) uintptr {
	if rsp/0xffffffff == 0 {
		return 4
	} else {
		return 8
	}
}

//KillAll杀死所有子进程并退出，通常用于程序出现超异常
func KillAllExit() {
	syscall.Kill(0, 9)
	os.Exit(255)
}

func Exit(vdatOut string) {
	os.Remove(vdatOut + ".hdat")
	os.Remove(vdatOut + ".bdat")
	os.Remove(vdatOut + ".tdat")
	os.Exit(255)
}

func SignalInit(c chan PtraceRet, k *bool) {
	var s = make(chan os.Signal)
	signal.Notify(s, syscall.SIGINT)
	go signalProcess(s, c, k)
}

func signalProcess(s <-chan os.Signal, c chan PtraceRet, k *bool) {
	switch <-s {
	case syscall.SIGINT:
		close(c)
		*k = true
	default:
		fmt.Printf("Recv signal:%#v\n", s)
	}
}
