package callsys

import (
	"debug/elf"
	"syscall"
	"unsafe"
)

func ptraceGetSPreg(pid int) (sp uint64, err error) {
	var regs syscall.PtraceRegs
	iov := syscall.Iovec{Base: (*byte)(unsafe.Pointer(&regs)), Len: uint64(unsafe.Sizeof(regs))}
	_, _, err = syscall.Syscall6(syscall.SYS_PTRACE, syscall.PTRACE_GETREGSET, uintptr(pid), uintptr(elf.NT_PRSTATUS), uintptr(unsafe.Pointer(&iov)), 0, 0)
	if err == syscall.Errno(0) {
		err = nil
	}
	return regs.Sp, err
}
