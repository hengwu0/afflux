package callsys

import (
	"syscall"
)

func ptraceGetSPreg(pid int) (rsp uint64, err error) {
	var regs syscall.PtraceRegs
	err = syscall.PtraceGetRegs(ptracePid, &regs)
	return regs.Rsp, err
}
