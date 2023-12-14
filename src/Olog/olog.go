package olog

import (
	"Blog"
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"syscall"
)

var baseHead *blog.BaseHead = new(blog.BaseHead)
var baseBody *blog.BaseBody = new(blog.BaseBody)
var baseTail *blog.BaseTail = new(blog.BaseTail)
var offHead, offBody, offTail int64
var logs []byte
var fdOutput *os.File
var w *bufio.Writer

func Init(fin, fout string) bool {
	var err error
	var fd *os.File

	if fout != "" {
		if fd, err = os.Create(fout); err != nil {
			fmt.Printf("Create output file(%s) failed! %v\n", fout, err)
			return false
		}
		fdOutput = fd
	} else {
		fdOutput = os.Stdout
	}
	w = bufio.NewWriterSize(fdOutput, 40960)

	if fd, err = parseHead(fin + ".bdat"); fd == nil {
		if fd, err = parseHead(fin); err != nil {
			fmt.Printf("parseHead dataBase file(%s) failed! %v\n", fin, err)
			return false
		} else if fd == nil {
			if m, err := blog.Md5(fin); err == nil {
				var newF string
				if blog.IsDir("/tmp/") {
					newF = "/tmp/" + m + ".dat"
				} else {
					newF = "./" + m + ".tmp"
				}
				if !blog.IsExist(newF) {
					blog.Ungz(fin, newF)
				}
				fd, err = parseHead(newF)
			} else {
				fmt.Printf("read fileMd5 failed!\n")
				return false
			}
		}
	}
	if fd == nil {
		fmt.Printf("Olog Init faild! Can't parseHead file(%s)%v\n", fin, err)
		return false
	}
	//在映射完后就可以立刻关闭文件了，打开文件是为了获取文件描述符来映射文件，
	//文件是否打开对映射内存的操作没有关系
	defer fd.Close()

	end, _ := fd.Seek(0, 2)
	fd.Seek(0, 0)
	if logs, err = syscall.Mmap(int(fd.Fd()), 0, int(end), syscall.PROT_READ, syscall.MAP_PRIVATE); err != nil {
		fmt.Printf("Mmap dataBase file(%s) failed! %v\n", fin, err)
		return false
	}

	//文件头校验
	if !blog.CheckBlog(logs) {
		fmt.Printf("dat check failed!!!\n")
		return false
	}

	return true
}

func parseHead(fin string) (fd *os.File, err error) {
	defer func() {
		if r := recover(); r != nil {
			fd = nil
			err = fmt.Errorf("parseHead failed!\n")
		}
	}()
	if fd, err = os.Open(fin); err != nil {
		return nil, err
	}
	buf := make([]byte, 6)
	io.ReadFull(fd, buf)

	switch {
	case string(buf[:6]) == "Afflux":
		return fd, nil
	case buf[0] == 0x1F && buf[1] == 0x8B:
		fd.Close()
		return nil, nil
	}
	return nil, fmt.Errorf("unknow fileHead!\n")

}

func Close() {
	w.Flush()
	fdOutput.Close()
	syscall.Munmap(logs)
}

func BuildLogMain(fin, fout string, pids []int, omit, sort string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("dataBase(%s) corrupted!!!\n", fin)
		}
	}()

	if !Init(fin, fout) {
		return
	}

	offHead, offBody, offTail = blog.ReadBlogHead(logs[blog.BlogHeadSize-8*3:])
	offHead += 4 //head大小目前暂未使用
	ReadTail(logs[offTail:], sort)

	if pids == nil {
		OutputLog(omit)
	} else {
		Query(pids)
	}

	Close()
}

func ReadTail(src []byte, sort string) {
	size := int64(binary.BigEndian.Uint32(src[0:]))
	buffer := bytes.NewReader(src[4:size])
	baseTail.Deserialize(buffer)
	baseTail.Sort(sort, logs)
	return
}

func BuildLogServe(fin, ipport string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("dataBase(%s) corrupted!!!\n", fin)
		}
	}()

	if !Init(fin, "") {
		return
	}
	defer Close()

	offHead, offBody, offTail = blog.ReadBlogHead(logs[blog.BlogHeadSize-8*3:])
	offHead += 4 //head大小目前暂未使用
	ReadTail(logs[offTail:], "id")
	
	Serve(ipport)
}
