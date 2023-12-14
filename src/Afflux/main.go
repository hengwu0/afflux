package main

import (
	"Blog"
	"Callsys"
	"Olog"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

var KillAll bool

func main() {
	var wg sync.WaitGroup
	if runtime.GOMAXPROCS(0) < 4 {
		runtime.GOMAXPROCS(4)
	}

	mode := flagParse()
	//mode:
	//1:Strace mode
	//2:Query mode, output log
	//3:Query mode, query ids
	//4:Query mode, web-server
	switch mode {
	case 1:
		pBuffer := make(chan callsys.PtraceRet, 1024)
		callsys.SignalInit(pBuffer, &KillAll)
		//启动数据缓冲模块
		go TransPtraceToBlog(pBuffer)
		//启动监控模块
		go PTraceModule(pBuffer)
		//启动BuildLog模块
		blog.BuildLogMain(VdatOut)
		//生成数据库文件
		wg.Add(1)
		go func(wg *sync.WaitGroup) { blog.Tar(VdatOut); wg.Done() }(&wg)
		//终端打印结果
		olog.BuildLogMain(VdatOut, VlogOut, nil, Vomits, Vsort)
	case 2:
		olog.BuildLogMain(VdatIn, VlogOut, nil, Vomits, Vsort)
	case 3:
		olog.BuildLogMain(VdatIn, VlogOut, Vids, "", "")
	case 4:
		olog.BuildLogServe(VdatIn, Vserve)
	}
	wg.Wait()

	if KillAll {
		callsys.KillAllExit()
	}
}

// 监控模块
func PTraceModule(pBuffer chan<- callsys.PtraceRet) {
	defer func() {
		if err := recover(); err != nil {
			s := fmt.Sprint(err)
			if strings.HasPrefix(s, "send on closed channel") {
				time.Sleep(time.Second * 2)
			} else {
				panic(err)
			}
		}
	}()

	if !callsys.StartTrace(Vcmd, Vomits, Vadditon) {
		fmt.Printf("afflux start failed!\n")
		callsys.Exit(VdatOut)
		callsys.KillAllExit()
	}
	var ptraceRet callsys.PtraceRet
	for {
		ptraceRet = callsys.ProcessTrace()
		if ptraceRet == nil {
			close(pBuffer)
			break
		} else {
			pBuffer <- ptraceRet
		}
	}
}

// 缓冲模块，用于缓冲并整合数据
func TransPtraceToBlog(pBuffer chan callsys.PtraceRet) {
	recv := <-pBuffer
	if !PtreeInit(recv) {
		callsys.Exit(VdatOut)
		callsys.KillAllExit()
	}

	for recv = range pBuffer {
		switch r := recv.(type) {
		case *callsys.Ship:
			Insert(r)
		case *callsys.CmdEnv:
			SetKey(r)
		case *callsys.RetUsage:
			SetUsage(r)
		case *callsys.RetVal:
			Delete(r)
			if Verr {
				if st, _ := r.ReadRetVal(); st != 0 {
					KillAll = true
					close(pBuffer)
				}
			}
		}
	}

	EndBlog()
	close(blog.BaseChan)
}

func PtreeInit(recv callsys.PtraceRet) bool {
	pusage = make(map[int]*callsys.RetUsage)
	dolists = make(map[int]*list_t)
	root.alive = true
	root.cmd = &blog.Cmd{}
	root.childs = make(map[int]*Node)
	if r, ok := recv.(*callsys.CmdEnv); !ok {
		if r, ok := recv.(*callsys.Ship); ok {
			new := newNode(0, r.Getpid())
			root.insert(new)
			new = newNode(r.Getpid(), r.ReadShip())
			root.insert(new)
		} else {
			fmt.Printf("Sorry, unsupport cmd: %s!\n", Vcmd)
			return false
		}
	} else {
		new := newNode(0, r.Getpid())
		new.cmdAdd(r)
		root.insert(new)
	}
	return true
}

func EndBlog() {
	if len(root.childs) != 0 {
		makeCleanAll(&root)
	}
	cleanlistsAll()
}
