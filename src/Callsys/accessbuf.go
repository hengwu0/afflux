package callsys

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func (myorder) ToInt(b []byte, SizeofPtr uintptr) int {
	if SizeofPtr == 8 {
		return int(ByteOrder.order.Uint64(b))
	}
	return int(ByteOrder.order.Uint32(b))
}

func (myorder) ToUintptr(b []byte, SizeofPtr uintptr) uintptr {
	if SizeofPtr == 8 {
		return uintptr(ByteOrder.order.Uint64(b))
	}
	return uintptr(ByteOrder.order.Uint32(b))
}

func CreateCmdEnv(pid int, cwd string, cmd, env []string) *CmdEnv {
	return &CmdEnv{pid, cwd, cmd, env}
}

func (r *CmdEnv) ReadCmdEnv() (cwd string, cmd, env []string) {
	return r.cwd, r.cmd, r.env
}

func CreateShip(pid, cpid int) *Ship {
	return &Ship{pid, cpid}
}

func (r *Ship) ReadShip() int {
	return r.cpid
}

func CreateRetVal(pid int, status int, signal int) *RetVal {
	return &RetVal{pid: pid, status: status, signal: signal}
}

func (r *RetVal) ReadRetVal() (int, int) {
	return r.status, r.signal
}

func CreateRetUsage(pid int, data []byte, addition string) *RetUsage {
	return &RetUsage{pid: pid, data: data, addition: addition}
}

func (r *RetUsage) ReadRetUsage() []byte {
	return r.data
}

func (r *RetUsage) ReadRetUsageAddition() string {
	return r.addition
}

func (p *RetVal) Getpid() int {
	return p.pid
}

func (p *CmdEnv) Getpid() int {
	return p.pid
}

func (p *Ship) Getpid() int {
	return p.pid
}

func (p *RetUsage) Getpid() int {
	return p.pid
}

func getUsage(pid int) *RetUsage {
	var data []byte
	var addition string
	if GetSysTimeUsage {
		path := fmt.Sprint("/proc/", pid, "/stat")
		if b, err := ioutil.ReadFile(path); err == nil {
			data = b
		}
	}
	if getFrom != nil {
		addition = getFromStatus(pid)
	}
	return CreateRetUsage(pid, data, addition)
}

//超时函数，超时后关闭channel
func Timeout(done chan int) {
	time.Sleep(time.Second * 1)
	close(done)
}

//安全调用getStrings的封装，避免子函数异常并死循环导致程序挂起
//限定子函数调用不得超时2秒！
func SafeGetString(pid int, rsp uintptr, SizeofPtr uintptr) (ret []string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("SafeGetString recoverd: %#v\n", r)
			ret = nil
		}
	}()

	c := make(chan int)
	go Timeout(c)
	return getStrings(c, pid, rsp, SizeofPtr)
}

func getStrings(done chan int, pid int, rsp uintptr, SizeofPtr uintptr) (strs []string) {
	var addr uintptr
	tmp := make([]byte, SizeofPtr)
	for count := 0; ; count++ {
		syscall.PtracePeekText(pid, rsp+SizeofPtr*uintptr(count), tmp) //get argv[x]
		addr = ByteOrder.ToUintptr(tmp, SizeofPtr)
		if addr == 0 {
			break
		}
		str = str[:0] //使用一个str，避免每次都要扩张str
		str_end := -1
		for j := uintptr(0); str_end == -1; j++ {
			syscall.PtracePeekText(pid, addr+j*SizeofPtr, tmp) //get argv[x][y]
			for k, a := range tmp {
				if a == 0 {
					str_end = k
					break
				}
			}
			if str_end == -1 {
				str = append(str, tmp...)
			} else {
				str = append(str, tmp[:str_end]...)
			}

			select {
			case <-done:
				return nil
			default:
			}
		}
		strs = append(strs, string(str))
	}
	return
}

func getTGID(pid int) (tgid int) {
	var f *os.File
	var err error
	if f, err = os.Open(fmt.Sprint("/proc/", pid, "/status")); err != nil {
		return
	}
	defer f.Close()
	r := bufio.NewReader(f)
	var line []byte
	Stgid := []byte("Tgid:")
	for {
		if line, _, err = r.ReadLine(); err != nil {
			break
		}
		if bytes.HasPrefix(line, Stgid) {
			line = bytes.TrimPrefix(line, Stgid)
			line = bytes.TrimSpace(line)
			tgid, _ = strconv.Atoi(string(line))
			break
		}
	}

	return
}

func getFromStatus(pid int) (res string) {
	var f *os.File
	var err error
	var line []byte
	var tmp []string
	if f, err = os.Open(fmt.Sprint("/proc/", pid, "/status")); err != nil {
		return
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		if line, _, err = r.ReadLine(); err != nil {
			break
		}
		ans := getFrom.FindSubmatch(line)
		for i := 1; i < len(ans); i++ {
			if len(ans[i]) > 0 {
				tmp = append(tmp, string(ans[i]))
			}
		}
	}

	return strings.Join(tmp, ", ")
}

func getCwd(path string) (cwd string) {
	cwd, _ = os.Readlink(path)
	return
}

func getCmdEnv(pid int, rsp uintptr, SizeofPtr uintptr) (*CmdEnv, bool) {
	tmp := make([]byte, SizeofPtr)
	if _, err := syscall.PtracePeekText(pid, rsp, tmp); err != nil {
		return nil, false
	}
	count := ByteOrder.ToInt(tmp, SizeofPtr)
	cmd := SafeGetString(pid, rsp+SizeofPtr, SizeofPtr)
	if cmd == nil {
		return nil, false
	}
	env := SafeGetString(pid, rsp+SizeofPtr*uintptr(count+2), SizeofPtr)
	//_, ppid, _ := getFromStatus(fmt.Sprint("/proc/", pid, "/status"))
	cwd := getCwd(fmt.Sprint("/proc/", pid, "/cwd"))

	if omit_reg != nil && omit_reg.FindStringIndex(cmd[0]) != nil {
		return CreateCmdEnv(pid, cwd, cmd, env), true
	}
	return CreateCmdEnv(pid, cwd, cmd, env), false
}
