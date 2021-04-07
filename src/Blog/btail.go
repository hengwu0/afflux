package blog

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

var baseTail *BaseTail

type BaseTail struct {
	//不能用bufio库，不能用缓存
	fd   *os.File
	Root *TailSubject
}

func TailInit(file string) {
	baseTail = new(BaseTail)
	baseTail.fd = NewTailFile(file)
}

func TailClose() {
	baseTail.fd.Close()
}

//多叉树
//left子节点；right兄弟节点
type TailSubject struct {
	Subject
	ppid        uint32
	sort        float64
	Deepth      uint32 //深度，用于排序
	Left, Right *TailSubject
}

//序列化内容格式，body可以大于4G故需使用int64
type Subject struct {
	Pid   uint32
	Index int64
}

func NewTailFile(file string) *os.File {
	if fd, err := os.Create(file); err != nil {
		fmt.Printf("Create dateBaseTail file failed! %v", err)
	} else {
		return fd
	}
	return nil
}

func max(x, y uint32) uint32 {
	if x < y {
		return y
	}
	return x
}

//插入到儿子群中
func (s *TailSubject) insert(newT *TailSubject) {
	s.Deepth = max(s.Deepth, newT.Deepth+1)
	if s.Left == nil {
		s.Left = newT
		return
	}
	s = s.Left
	for {
		if s.Right == nil {
			s.Right = newT
			return
		}
		s = s.Right
	}
}

//bfs查找ppid
func (b *BaseTail) FindParent(s, newT *TailSubject) uint32 {
	if s == nil {
		return 0
	}
	for {
		//找到了，插入到儿子群中
		if s.Pid == newT.ppid {
			//b.UpdateRoot(s, newT)
			s.insert(newT)
			return s.Deepth + 1
		}
		//先找兄弟群
		if d := b.FindParent(s.Right, newT); d == 0 {
			//再找儿子群，找不到不能跟新深度值
			if d = b.FindParent(s.Left, newT); d != 0 {
				s.Deepth = max(s.Deepth, d)
				return s.Deepth + 1
			} else {
				return 0
			}
		} else {
			//兄弟群更新深度
			return max(s.Deepth, d)
		}
	}
}

//孤儿都只在root的兄弟节点
func (b *BaseTail) FindAllChilds(newT *TailSubject) {
	s := b.Root
	//root单独处理
	if b.Root.ppid == newT.Pid {
		newT.insert(b.Root)
		newT.Right = b.Root.Right
		b.Root.Right = nil //删除root原有的兄弟节点
		b.Root = newT
		s = newT //为了继续寻找Childs
	}

	for s.Right != nil {
		if s.Right.ppid == newT.Pid {
			newT.insert(s.Right)
			tmp := s.Right
			s.Right = s.Right.Right
			tmp.Right = nil
		} else {
			s = s.Right
		}
	}
	return
}

//寻找并挂parent
func (s *TailSubject) SetRootParent(root *TailSubject) bool {
	if s == nil {
		return false
	}
	if root.ppid == s.Pid {
		s.insert(root)
		return true
	}
	return s.Left.SetRootParent(root) || s.Right.SetRootParent(root)
}

func (b *BaseTail) FindRootParent() {
	s := b.Root
	for s.Right != nil {
		//先挂parent
		if b.Root.ppid == s.Right.Pid {
			s.Right.insert(b.Root)
		}
		//再处理root的兄弟节点
		if b.Root.ppid == s.Right.Pid || s.Right.Left.SetRootParent(b.Root) {
			rtmp := b.Root
			b.Root = s.Right
			s.Right = s.Right.Right
			b.Root.Right = rtmp.Right
			rtmp.Right = nil
			return
		}
		s = s.Right
	}
}

//插入一个节点到树中
//先寻找到孤儿，若孤儿不是root，则再找父节点
func (b *BaseTail) AddTail(newT *TailSubject, ppid uint32) {
	newT.ppid = ppid
	if b.Root == nil {
		b.Root = newT
		return
	}

	//先找儿子
	b.FindAllChilds(newT)

	//为root找parent
	if b.Root == newT {
		b.FindRootParent()
		return
	}
	//找parent(根据ppref，优化)
	//pstree保证先发送完毕子节点，再发送父节点，此处不再需要FindParent
	//if b.FindParent(b.Root, newT) == 0 {
	{
		//找不到？？？插入到root的兄弟群
		s := b.Root
		for {
			if s.Right == nil {
				s.Right = newT
				return
			}
			s = s.Right
		}
	}
}

//这里不使用seek，因为要使用buffer
func (b *BaseTail) Serialize(w io.Writer) {
	b.Root.serialize(w)
}

func (s *TailSubject) serialize(w io.Writer) {
	if s == nil {
		binary.Write(w, binary.BigEndian, new(Subject))
		return
	}
	binary.Write(w, binary.BigEndian, s.Subject)
	s.Left.serialize(w)
	s.Right.serialize(w)
}

func (b *BaseTail) Deserialize(r io.Reader) {
	var d uint32
	b.Root, d = deserialize(r)
	if b.Root != nil {
		b.Root.Deepth = d
	}
	return
}

//返回*TailSubject和当前深度+1
func deserialize(rd io.Reader) (*TailSubject, uint32) {
	var buf, empty Subject
	binary.Read(rd, binary.BigEndian, &buf)
	if buf == empty {
		return nil, 0
	}
	t := &TailSubject{Subject: buf}
	var r uint32
	t.Left, t.Deepth = deserialize(rd)
	t.Right, r = deserialize(rd)
	return t, max(t.Deepth+1, r) //r已经是深度+1的值了
}

func (b *BaseTail) WriteTail() {
	buf := new(bytes.Buffer)
	baseTail.Serialize(buf)
	b.fd.Seek(0, 2)
	buf.WriteTo(b.fd)
}

//------------------------------------------------
//copy from oblog, for debug.
/*
var blank []byte
var w *os.File

func (b *BaseTail) WriteTail_bak() {
	w, _ = os.Create("~/test/afflux/src/make.log")
	blank = make([]byte, 10)
	TransDeepth(baseTail.Root, 0)
	WriteTree(baseTail.Root)
	os.Exit(0)
}

func TransDeepth(t *TailSubject, val uint32) {
	if t == nil {
		return
	}
	t.Deepth = val
	TransDeepth(t.Left, val+1)
	TransDeepth(t.Right, val)
}
func WriteTree(t *TailSubject) {
	if t == nil {
		return
	}
	if t.Right != nil {
		blank[t.Deepth] = 1
	}

	WriteBody(t)
	WriteTree(t.Left)
	blank[t.Deepth] = 0
	WriteTree(t.Right)

}
func WriteBody(t *TailSubject) {
	for i := uint32(0); i < t.Deepth; i++ {
		if blank[i] == 0 {
			w.WriteString("    ")
		} else {
			w.WriteString("│   ")
		}
	}
	// switch {
	// case t.Left!=nil && t.Right!=nil:w.WriteString("├── ")
	// case t.Left!=nil && t.Right==nil:w.WriteString("└── ")
	// case t.Left==nil && t.Right!=nil:w.WriteString("├── ")
	// case t.Left==nil && t.Right==nil:w.WriteString("└──")
	// }
	w.WriteString("|── ")

	bufferBody := fmt.Sprintf("%d(%d)", t.Pid, t.ppid)
	bufferBody += "\r\n"

	w.WriteString(bufferBody)
}
*/

func (s *TailSubject) Search(pid uint32) []int64 {
	if s == nil {
		return nil
	}
	if s.Pid == pid {
		return []int64{s.Index}
	}
	if r := s.Left.Search(pid); r != nil {
		return append(r, s.Index)
	}
	if r := s.Right.Search(pid); r != nil {
		return r
	}
	return nil
}

//dfs排序
func (b *BaseTail) Sort(sort string, logs []byte) {
	fid := func(t *TailSubject) float64 {
		return float64(t.Pid)
	}
	fidv := func(t *TailSubject) float64 {
		return float64(t.Pid) * -1
	}
	ftime := func(t *TailSubject) float64 {
		return float64(baseBody.ReadTime(logs[t.Index:]))
	}
	ftimev := func(t *TailSubject) float64 {
		return float64(baseBody.ReadTime(logs[t.Index:])) * -1
	}
	fadd := func(t *TailSubject) float64 {
		return float64(baseBody.ReadAdd(logs[t.Index:]))
	}
	faddv := func(t *TailSubject) float64 {
		return float64(baseBody.ReadAdd(logs[t.Index:])) * -1
	}
	switch sort {
	case "time":
		writeSort(b.Root, ftime)
	case "timev":
		writeSort(b.Root, ftimev)
	case "add":
		writeSort(b.Root, fadd)
	case "addv":
		writeSort(b.Root, faddv)
	case "idv":
		writeSort(b.Root, fidv)
	case "id":
		writeSort(b.Root, fid)
	default:
	}

	for tmp := b.Root; tmp != nil; tmp = tmp.Right {
		tmp.sort_any()
	}
}

func writeSort(t *TailSubject, f func(t *TailSubject) float64) {
	if t == nil {
		return
	}
	t.sort = f(t)
	writeSort(t.Left, f)
	writeSort(t.Right, f)
}

func (s *TailSubject) sort_any() {
	if s.Left == nil {
		return
	}
	//TODO: 冒泡排序太慢啦(toolchain, 25万, 2秒)
	for head, headH := s.Left, (*TailSubject)(nil); head != nil; headH, head = head, head.Right {
		for tmp, tmpH := head.Right, head; tmp != nil; tmpH, tmp = tmp, tmp.Right {
			if (head.sort > tmp.sort) || (head.sort == tmp.sort && head.Pid > tmp.Pid) {
				exchange(s, head, tmp, headH, tmpH) //必须传入headH和tmpH,否则更慢！！！
				head, tmp = tmp, head
			}
		}
	}
	for tmp := s.Left; tmp != nil; tmp = tmp.Right {
		tmp.sort_any()
	}
}

//a, b无顺序要求
func exchange(s, a, b, aH, bH *TailSubject) {
	if aH != nil && bH != nil {
		aH.Right = b
		bH.Right = a
	} else if aH == nil {
		s.Left = b
		bH.Right = a
	} else {
		s.Left = a
		aH.Right = b
	}
	a.Right, b.Right = b.Right, a.Right
	if a.Right == a {
		a.Right = b
	} else if b.Right == b {
		b.Right = a
	}
}
