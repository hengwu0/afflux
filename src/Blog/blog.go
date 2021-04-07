package blog

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const BlogHead = "Afflux dataBase"

//定义dataBase文件头大小
const BlogHeadSize = len(BlogHead) + 1 + 0 + 8*3

//定义每个节的节头大小
//head最大4G，其实大于1G map就已经爆了
const baseHeadSize = 4

func Init(file string) bool {
	if HeadInit(file + ".hdat"); baseHead.fd == nil {
		return false
	}

	if BodyInit(file + ".bdat"); baseBody.fd == nil {
		return false
	}
	WriteBlogHead(baseBody.fd)

	if TailInit(file + ".tdat"); baseTail.fd == nil {
		return false
	}
	return true
}

func Close(file string) {
	baseHead.fd.Close()
	baseBody.fd.Close()
	baseTail.fd.Close()
}

func Tar(file string) {
	os.Remove(file + ".hdat")
	os.Remove(file + ".tdat")
	if err := Gz(file+".bdat", file); err != nil {
		os.Rename(file+".bdat", file)
		return
	}
	if IsDir("/tmp/") {
		if m, err := Md5(file); err == nil {
			Move(file+".bdat", "/tmp/"+m+".dat")
			return
		}
	}
	os.Remove(file + ".bdat")
}

/*
baseHead Body Body ...//baseHead: "Afflux dataBase" version BodyOffset(64) HeadOffset(64) TailOffset(64)
Body Body Body Body...//读取body不需要添加偏移
HeadSize Head Head ...		HeadSize(32): head的大小,已包含HeadSize大小
Head Head Head Head...//读取head需要添加偏移，偏移不包括HeadSize
TailSize Tail Tail ...		TailSize(32): tail的大小,已包含TailSize大小
Tail Tail Tail Tail...
*/
func BuildLogMain(file string) {
	if ok := Init(file); !ok {
		return
	}

	for cmd := range BaseChan {
		body := baseHead.WriteHead(cmd)
		tail := baseBody.WriteBody(body)
		baseTail.AddTail(tail, cmd.puid)
	}
	baseTail.WriteTail()
	offHead := HeadCopy(baseHead.fd)
	offTail := HeadCopy(baseTail.fd)
	baseHeadWrite(offHead, offTail)
	Close(file)
}

func CheckBlog(b []byte) bool {
	offHead, offBody, offTail := ReadBlogHead(b[BlogHeadSize-8*3:])
	if offHead == 0 || offBody == 0 || offTail == 0 {
		fmt.Printf("CheckBlog failed! H:%#v, B:%#v T:%#v\n", offHead, offBody, offTail)
		return false
	}

	sizeHead := int64(binary.BigEndian.Uint32(b[offHead:]))
	sizeTail := int64(binary.BigEndian.Uint32(b[offTail:]))
	if offHead+sizeHead != offTail {
		fmt.Printf("CheckBlog for head failed! %d+%d!=%d\n", offHead, sizeHead, offTail)
		return false
	}
	if offTail+sizeTail != int64(len(b)) {
		fmt.Printf("CheckBlog for tail failed! %d+%d!=%d\n", offTail, sizeTail, len(b))
		return false
	}

	return true
}

func WriteBlogHead(fd *os.File) {
	buffer := new(bytes.Buffer)
	buffer.WriteString(BlogHead)
	buf := make([]byte, BlogHeadSize-len(BlogHead))
	buf[0] = 0x1 //数据库版本
	buffer.Write(buf)
	buffer.WriteTo(fd)
}

func ReadBlogHead(src []byte) (offH, offB, offT int64) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("ReadBlogHead failed! H:%#v, B:%#v T:%#v\n", offH, offB, offT)
		}
	}()
	offB = int64(binary.BigEndian.Uint64(src[0:]))
	offH = int64(binary.BigEndian.Uint64(src[8:]))
	offT = int64(binary.BigEndian.Uint64(src[16:]))
	return
}

//写入head的size
func HeadWriteSize(fd *os.File, val int64) {
	buf := make([]byte, baseHeadSize)
	binary.BigEndian.PutUint32(buf, uint32(val))
	fd.Write(buf)
	return
}

//填充Head头并copy主体
func HeadCopy(fd *os.File) uint64 {
	fd.Sync() //flush
	off, _ := baseBody.fd.Seek(0, 2)
	val, _ := fd.Seek(0, 2)
	HeadWriteSize(baseBody.fd, val+4)
	fd.Seek(0, 0)
	io.Copy(baseBody.fd, fd)
	return uint64(off)
}

func baseHeadWrite(offH uint64, offT uint64) {
	buf := make([]byte, 8*3)
	binary.BigEndian.PutUint64(buf[0:], uint64(BlogHeadSize))
	binary.BigEndian.PutUint64(buf[8:], offH)
	binary.BigEndian.PutUint64(buf[16:], offT)
	baseBody.fd.WriteAt(buf, int64(BlogHeadSize-8*3))
	return
}
