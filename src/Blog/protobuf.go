package blog

import (
	"bytes"
	"fmt"
	"strconv"
	// "strings"
	"time"
)

type Cmd struct {
	status, signal int16
	uid, puid      uint32 //unique id, parent unique id;
	time           int64
	now            time.Time
	cwd            string
	addition       string
	cmd, env       []string
}

var BaseChan = make(chan *Cmd, 1024)

func setPartition(strs []string) []string {
	return strs
}

//创建并返回一个Cmd条目
func Create(cwd string, cmd, env []string) *Cmd {
	cmd = setPartition(cmd)
	return &Cmd{puid: 0,
		cwd: cwd, cmd: cmd, env: env,
		status: -1, signal: -1, time: -1, now: time.Now()}
}

func add(a, b []string) []string {
	var end int
	if a[0][0] != '{' {
		a[0] = "{" + a[0]
		end = len(a) - 1
		a[end] = a[end] + "}"
	}

	b[0] = "{" + b[0]
	end = len(b) - 1
	b[end] = b[end] + "}"

	return append(a, b...)
}
func add2(a, b []string) []string {
	tmp := []string{"{"}
	if a[0] != "{" {
		a = append(tmp, a...)
		a = append(a, "}")
	}

	b = append(tmp, b...)
	b = append(b, "}")

	return append(a, b...)
}
func (c *Cmd) Add(n *Cmd) {
	if c.cwd != n.cwd {
		c.cwd = "{" + c.cwd + "};{" + n.cwd + "}"
	}

	c.cmd = add(c.cmd, n.cmd)
	c.env = add2(c.env, n.env)
}

func (c *Cmd) Setpuid(pp int) {
	c.puid = uint32(pp)
}
func (c *Cmd) Getpuid() int {
	return int(c.puid)
}

func (c *Cmd) Setuid(pp int) {
	c.uid = uint32(pp)
}
func (c *Cmd) Getuid() int {
	return int(c.uid)
}

func (c *Cmd) SetExitCode(status, signal int) {
	c.status = int16(status)
	c.signal = int16(signal)
}

func (c *Cmd) SetUsage(data []byte, addition string, clktck int64) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("SetUsage(get time info) recoverd: %#v\n", r)
			c.time = 0
		}
	}()

	c.addition = addition
	if data == nil {
		c.time = round(int64(time.Since(c.now)))
		return
	}
	//去除括号内的文件名(防止含空格)
	tmp := data[bytes.IndexByte(data, '(')+1 : bytes.LastIndexByte(data, ')')]
	data = bytes.Replace(data, tmp, nil, 1)
	infos := bytes.Fields(data)
	//数组从0开始
	t1, _ := strconv.Atoi(string(infos[13])) //utime
	t2, _ := strconv.Atoi(string(infos[14])) //stime
	t3, _ := strconv.Atoi(string(infos[15])) //cutime
	t4, _ := strconv.Atoi(string(infos[16])) //cstime
	t5, _ := strconv.Atoi(string(infos[42])) //guest_time
	t6, _ := strconv.Atoi(string(infos[43])) //cguest_time
	c.time = int64(t1+t2+t3+t4+t5+t6) * clktck
}

func round(num int64) int64 {
	var i, a, b, c int64
	i, a = 1, num
	for a > 10 {
		b, c = a, b
		b = a % 10
		a, i = a/10, i*10
	}
	return a*i + b*i/10 + c*i/100
}

//TODO: //setsid进程又卡住的
func (c *Cmd) SetEExit() {

}
