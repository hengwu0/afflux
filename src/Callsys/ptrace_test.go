package callsys

import (
	"fmt"
	"os"
	"testing"
	"time"
)

var mp map[int][]PtraceRet = make(map[int][]PtraceRet)

func check1(res *CmdEnv, t *testing.T) {
	mp[res.Getpid()] = append(mp[res.Getpid()], res)
	if n := len(res.cmd); n != 2 {
		t.Errorf("cmd len(%#v) want %#v, but get %#v!", res.cmd, 2, n)
	}
	if n := len(res.env); n < 11 {
		t.Errorf("env len(%#v) need more than %#v, but get %#v!", res.env, 10, n)
	}
}

func check2(res *RetVal, t *testing.T) {
	if _, ok := mp[res.Getpid()]; !ok {
		return
	}
	mp[res.Getpid()] = append(mp[res.Getpid()], res)
	if (res.status != 0 && res.status != 4) || res.signal != -1 {
		t.Errorf("resRetVal want (0,-1), but get %#v!", res)
	}
}

func check3(res *RetUsage, t *testing.T) {
	if _, ok := mp[res.Getpid()]; !ok {
		return
	}
	mp[res.Getpid()] = append(mp[res.Getpid()], res)
	if GetSysTimeUsage && res.data == nil {
		t.Errorf("resRetUsage returned nil, pid=%#v!", res.Getpid())
	}
}

func checkRes(res PtraceRet, t *testing.T) (pid int) {
	pid = res.Getpid()
	switch r := res.(type) {
	case *CmdEnv:
		check1(r, t)
	case *RetVal:
		check2(r, t)
	case *RetUsage:
		check3(r, t)
	case *Ship:
	default:
		t.Errorf("unknow type of res: %#v", res)
	}
	return
}

var in, out1, out2, key int

func checkRes2(res PtraceRet, t *testing.T) (pid int) {
	pid = res.Getpid()
	switch res.(type) {
	case *CmdEnv:
		key++
	case *RetVal:
		out2++
	case *RetUsage:
		out1++
	case *Ship:
		in++
	default:
		t.Errorf("unknow type of res: %#v", res)
	}
	return
}

func checkMp(t *testing.T) {
	for _, v := range mp {
		if lenth := len(v); lenth != 3 {
			t.Errorf("check mp count(%d)!=3!", lenth)
			return
		}
		if _, ok := v[0].(*CmdEnv); !ok {
			t.Errorf("check CmdEnv in mp failed!")
		}
		if _, ok := v[1].(*RetUsage); !ok {
			t.Errorf("check RetUsage in mp failed!")
		}
		if _, ok := v[2].(*RetVal); !ok {
			t.Errorf("check RetVal in mp failed!")
		}
	}

	// for _,v := range mp {
	// a :=
	// t.Errorf("CmdEnv! %#v", a)
	// b :=
	// t.Errorf("RetUsage! %#v", b)
	// c :=
	// t.Errorf("RetVal! %#v", c)
	// }
}

func ATestGetFromStatus(t *testing.T) {
	pid := os.Getpid()
	tgid, ppid, name := getFromStatus(fmt.Sprint("/proc/", pid, "/status"))
	if tgid <= 1 || ppid <= 1 || name == "" {
		t.Errorf("getFromStatus returned tgid=%#v, ppid=%#v, name=%#v!", tgid, ppid, name)
	}
}

func ATestGetUsage(t *testing.T) {
	pid := os.Getpid()
	r := getUsage(pid, fmt.Sprint("/proc/", pid, "/stat"))
	if r == nil || r.data == nil {
		t.Errorf("getUsage returned %#v, pid=%#v!", r, pid)
	}
}

func ATestGetCwd(t *testing.T) {
	pid := os.Getpid()
	cwd := getCwd(fmt.Sprint("/proc/", pid, "/cwd"))
	if cwd == "" {
		t.Errorf("getCwd returned name=%#v!", cwd)
	}
}

func ATestPTrace(t *testing.T) {
	if res := StartTrace("./testdata/cmd -flag"); res == nil {
		t.Errorf("test StartTrace start failed!")
		return
	}
	if ptracePid == 0 {
		t.Errorf("test StartTrace failed!")
		return
	}

	var count int
	for {
		count++

		if res := ProcessTrace(); res != nil {
			checkRes(res, t)
		} else {
			break
		}
	}
	if count < 11 {
		t.Errorf("ProcessTrace need to run more than %#v times, but ran %#v times!", 10, count)
	}

	if len(mp) == 0 {
		t.Errorf("ProcessTrace ptraced none child!")
	}
	checkMp(t)
}

func PTrace(t *testing.T) {
	if res := StartTrace("GOCACHE=off CGO_ENABLED=1 go tool dist test"); res == nil {
		t.Errorf("test StartTrace start failed!")
		return
	}
	if ptracePid == 0 {
		t.Errorf("test StartTrace failed!")
		return
	}

	in = 0
	out1 = 0
	out2 = 0
	key = 0
	for {
		if res := ProcessTrace(); res != nil {
			checkRes2(res, t)
		} else {
			break
		}
	}
	if in+1 < out1 {
		fmt.Fprintln(ww, "ERROR: in, key, out1, out2====", in, key, out1, out2)
	} else {
		fmt.Fprintln(ww, "       in, key, out1, out2====", in, key, out1, out2)
	}
	ww.Sync()
}

var ww *os.File

func TestPTrace(t *testing.T) {
	ww, _ = os.Create("./outlog.log")
	defer ww.Close()
	fmt.Fprintln(ww, "start test")
	for {
		PTrace(t)
		time.Sleep(120 * time.Second)
	}
}
