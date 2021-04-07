package blog

import (
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var baseBody *BaseBody

//buf,最小为BodySubjectHead大小
var rwBuff []byte = make([]byte, BodySubjectHead)

type BaseBody struct {
	//不能用bufio库，不能用缓存
	fd *os.File
}

func BodyInit(file string) {
	baseBody = new(BaseBody)
	baseBody.fd = NewBodyFile(file)
}

func BodyClose() {
	baseBody.fd.Close()
}

const BodySubjectHead = 4 + 2 + 2 + 4 + 4*2 + 8 + 4

//head设定小于4G故使用uint32
type BodySubject struct {
	//Size uint32	BodySubjectHead包含size大小，但不导出
	Status, Signal uint16
	Pid            uint32
	Cwd, Cmd       uint32
	Time           uint64
	//Count uint32 BodySubjectHead包含该大小，但不导出
	Env      []uint32
	Addition string
	Args     []string
}

func NewBodyFile(file string) *os.File {
	if f, err := os.Create(file); err != nil {
		fmt.Printf("Create dateBaseBody file failed! %v", err)
	} else {
		return f
	}
	return nil
}

func (s *BodySubject) Size() (size int) {
	lenth := len(s.Args)
	size = BodySubjectHead + len(s.Env)*4 + lenth
	for i := 0; i < lenth; i++ {
		size += len(s.Args[i])
	}
	size += 4 + len(s.Addition)
	return
}

func (b *BaseBody) WriteSubject(s *BodySubject) (offset int64) {
	offset, _ = b.fd.Seek(0, 1)

	size := uint32(s.Size())
	if len(rwBuff) < int(size) {
		rwBuff = make([]byte, 2*size)
	}

	binary.BigEndian.PutUint16(rwBuff[4:], s.Status)
	binary.BigEndian.PutUint16(rwBuff[6:], s.Signal)
	binary.BigEndian.PutUint32(rwBuff[8:], s.Pid)
	binary.BigEndian.PutUint32(rwBuff[12:], s.Cwd)
	binary.BigEndian.PutUint32(rwBuff[16:], s.Cmd)
	binary.BigEndian.PutUint64(rwBuff[20:], s.Time)

	binary.BigEndian.PutUint32(rwBuff[28:], uint32(len(s.Env)))
	for i, v := range s.Env {
		binary.BigEndian.PutUint32(rwBuff[BodySubjectHead+i*4:], v)
	}

	off := BodySubjectHead + len(s.Env)*4
	binary.BigEndian.PutUint32(rwBuff[off:], uint32(len(s.Addition)))
	copy(rwBuff[off+4:], s.Addition)
	off += len(s.Addition) + 4

	if len(s.Args) > 0 {
		size--
		copy(rwBuff[off:], []byte(strings.Join(s.Args, "\x00")))
	}

	binary.BigEndian.PutUint32(rwBuff[0:], size)

	b.fd.Write(rwBuff[0:size])
	return
}

//从传入切片的中提取单独一个数据，只需传入切片头即可
func (b *BaseBody) ReadSubject(src []byte) *BodySubject {
	var s BodySubject

	size := binary.BigEndian.Uint32(src[0:4])

	s.Status = binary.BigEndian.Uint16(src[4:])
	s.Signal = binary.BigEndian.Uint16(src[6:])
	s.Pid = binary.BigEndian.Uint32(src[8:])
	s.Cwd = binary.BigEndian.Uint32(src[12:])
	s.Cmd = binary.BigEndian.Uint32(src[16:])
	s.Time = binary.BigEndian.Uint64(src[20:])

	count := binary.BigEndian.Uint32(src[28:])
	s.Env = make([]uint32, count)
	for i := uint32(0); i < count; i++ {
		s.Env[i] = binary.BigEndian.Uint32(src[BodySubjectHead+i*4:])
	}

	off := BodySubjectHead + count*4
	offend := binary.BigEndian.Uint32(src[off:]) + off + 4
	s.Addition = string(src[off+4 : offend])

	s.Args = strings.Split(string(src[offend:size]), "\x00")
	if len(s.Args) == 1 && s.Args[0] == "" {
		s.Args = nil
	}
	return &s
}

var getNum = regexp.MustCompile(`[+-]?\d+\.?\d+`)

func (b *BaseBody) ReadAdd(src []byte) float64 {
	count := binary.BigEndian.Uint32(src[28:])
	off := BodySubjectHead + count*4
	offend := binary.BigEndian.Uint32(src[off:]) + off + 4
	if s, err := strconv.ParseFloat(string(getNum.Find(src[off+4:offend])), 32); err == nil {
		return s
	}
	return 0
}
func (b *BaseBody) ReadTime(src []byte) uint64 {
	return binary.BigEndian.Uint64(src[20:])
}

func (b *BaseBody) WriteBody(s *BodySubject) *TailSubject {
	var t TailSubject
	t.Index = b.WriteSubject(s)
	t.Pid = s.Pid
	return &t
}
