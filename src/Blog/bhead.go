//head最大4G，其实大于1G map就已经爆了
package blog

import (
	"fmt"
	"os"
)

type BaseHead struct {
	fd   *os.File
	strs map[string]int64
}

var baseHead *BaseHead

func HeadInit(file string) {
	baseHead = new(BaseHead)
	baseHead.fd = NewHeadFile(file)
	baseHead.strs = make(map[string]int64)
}

func HeadClose() {
	baseHead.fd.Close()
}

//创建文件并填充8位head
//若文件存在，则truncate
func NewHeadFile(file string) *os.File {
	if fd, err := os.Create(file); err != nil {
		fmt.Printf("Create dateBaseHead file failed! %v", err)
	} else {
		return fd
	}
	return nil
}

func (b *BaseHead) WriteString(s string) uint32 {
	if v, ok := b.strs[s]; ok {
		return uint32(v)
	}
	offset, _ := b.fd.Seek(0, 1)
	b.fd.WriteString(s + "\x00")
	b.strs[s] = offset
	return uint32(offset)
}

func (b *BaseHead) ReadString(s []byte) string {
	var i int
	for s[i] != 0 {
		i++
	}
	return string(s[:i])
}

func (b *BaseHead) WriteHead(cmd *Cmd) *BodySubject {
	var s BodySubject

	s.Cmd = b.WriteString(cmd.cmd[0])
	s.Cwd = b.WriteString(cmd.cwd)
	for _, v := range cmd.env {
		s.Env = append(s.Env, b.WriteString(v))
	}

	s.Status = uint16(cmd.status)
	s.Signal = uint16(cmd.signal)
	s.Pid = cmd.uid
	s.Time = uint64(cmd.time)
	s.Addition = cmd.addition
	s.Args = cmd.cmd[1:]

	return &s
}
