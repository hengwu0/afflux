package callsys

import (
	"encoding/binary"
	"syscall"
)

type myorder struct {
	order binary.ByteOrder
}

type CmdEnv struct {
	pid      int
	cwd      string
	cmd, env []string
}

type RetVal struct {
	pid, status, signal int
}

type Ship struct {
	pid, cpid int
}

type RetUsage struct {
	pid      int
	data     []byte
	addition string
}

type PtraceRet interface {
	Getpid() int
}

type paused struct {
	pid     int
	status  syscall.WaitStatus
	on      bool
	channel chan int
}
