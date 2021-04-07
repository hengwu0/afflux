/*
ptrace存在缺陷，无法从根本上解决：
1、若某进程的一个线程执行fork，而另一个线程执行exit。
则可能会存在子进程创建成功，而ptrace无法监控到创建。
但是子进程退出反而能监控到。
2、存在子进程创建成功，但是被内核先调度，
从而发生先监控到子进程的EXEC事件，
再监控到父进程的创建事件。

情况1目前暂未解决，其不会造成进程信息丢失，
最多导致进程树构建异常
情况2已在Afflux/pstree.go中解决
*/
package callsys

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"syscall"
)

var ptrace_setoptions int = syscall.PTRACE_O_TRACEFORK | syscall.PTRACE_O_TRACEVFORK |
	syscall.PTRACE_O_TRACECLONE | syscall.PTRACE_O_TRACEEXEC |
	syscall.PTRACE_O_TRACEEXIT | syscall.PTRACE_O_TRACESYSGOOD

var omit_reg *regexp.Regexp
var getFrom *regexp.Regexp
var ptracePid int
var sig syscall.Signal
var str []byte = make([]byte, 1, 1024)
var GetSysTimeUsage bool
var knowpid map[int]bool

//执行buildCmd命令，并抓取该命令的CmdEnv值
//返回nil表示启动ptrace失败
//注意StartTrace必须会绑定到OS线程
func StartTrace(buildCmd []string, omit string, addition string) bool {
	runtime.LockOSThread() //ptrace need it
	cmd := exec.Command(buildCmd[0], buildCmd[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Ptrace: true,
	}
	if err := cmd.Start(); err != nil {
		fmt.Printf("Can't start command with Ptrace: /bin/sh -c %#v, %v\n", buildCmd, err)
		return false
	}

	if err := cmd.Wait(); err.Error() != "stop signal: trace/breakpoint trap" {
		fmt.Printf("trace failure(%v)!\n", err)
		return false
	}
	ptracePid = int(cmd.Process.Pid)
	if err := syscall.PtraceSetOptions(ptracePid, ptrace_setoptions); err != nil {
		fmt.Printf("trace setOptions failure(%v)!\n", err)
		return false
	}

	// var regs syscall.PtraceRegs
	// if err := syscall.PtraceGetRegs(ptracePid, &regs); err != nil {
	// fmt.Printf("trace getRegs failure(%v)!\n", err)
	// return nil
	// }
	// r, _ := getCmdEnv(ptracePid, uintptr(regs.Rsp), GetPtrSize(regs.Rsp))

	knowpid = make(map[int]bool)
	if omit != "" {
		omit_reg = regexp.MustCompile(omit)
	}
	if addition != "" {
		getFrom = regexp.MustCompile(addition)
	}

	return true
}

//抓取一次数据，返回CmdEnv、Usage、RetVal三种之一，不会返回nil
//返回nil代表退出
//若返回数据为RetVal，则再次调用pid入参必须为0(监控的子进程已不存在)
func ProcessTrace() PtraceRet {
	var err error
	var status syscall.WaitStatus
	var regs syscall.PtraceRegs

	for {
		if ptracePid != 0 { //未暂停就不需要continue
			syscall.PtraceCont(ptracePid, int(sig)) //ignore err
		}

		if ptracePid, err = syscall.Wait4(-1, &status, syscall.WALL, nil); err != nil {
			if err.Error() == "no child processes" {
				return nil
			} else {
				fmt.Printf("trace Wait4 %d failure(%v)!\n", ptracePid, err)
				ptracePid = 0
				continue
			}
		}

		sig = status.StopSignal()
		if int(sig) == -1 || sig == 0x80|syscall.SIGTRAP || sig == syscall.SIGSTOP {
			sig = 0
		}

		switch status.TrapCause() {
		case syscall.PTRACE_EVENT_VFORK, syscall.PTRACE_EVENT_FORK:
			cpid, _ := syscall.PtraceGetEventMsg(ptracePid)
			if cpid != 0 {
				p := getTGID(ptracePid)
				knowpid[p] = true
				return CreateShip(p, int(cpid))
			}
		//启动子进程
		case syscall.PTRACE_EVENT_EXEC:
			if err = syscall.PtraceGetRegs(ptracePid, &regs); err != nil {
				fmt.Printf("trace getRegs %d failure(%v)!\n", ptracePid, err)
			} else {
				if r, detach := getCmdEnv(ptracePid, uintptr(regs.Rsp), GetPtrSize(regs.Rsp)); r != nil {
					knowpid[ptracePid] = true
					if detach {
						syscall.PtraceDetach(ptracePid)
						ptracePid = 0
					}
					return r
				}
			}
		//子进程准备退出
		case syscall.PTRACE_EVENT_EXIT:
			if getTGID(ptracePid) == ptracePid {
				if r := getUsage(ptracePid); r != nil {
					knowpid[ptracePid] = true
					return r
				}
			}
		default:
			//子进程退出码
			if !status.Stopped() || status.Exited() || status.Signaled() {
				if _, ok := knowpid[ptracePid]; ok {
					r := CreateRetVal(ptracePid, status.ExitStatus(), int(status.Signal()))
					delete(knowpid, ptracePid)
					ptracePid = 0
					return r
				}
				ptracePid = 0
			}
		}

	}
}
