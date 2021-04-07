package blog

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func createT(n uint32) *TailSubject {
	return &TailSubject{Subject: Subject{Pid: n}}
}

func TestTree1(t *testing.T) {
	var baseTail *BaseTail
	baseTail = new(BaseTail)
	//pid不能为0
	baseTail.AddTail(createT(1), 100)
	baseTail.AddTail(createT(2), 1)
	baseTail.AddTail(createT(3), 2)
	baseTail.AddTail(createT(4), 2)
	baseTail.AddTail(createT(5), 4)
	baseTail.AddTail(createT(6), 8)
	baseTail.AddTail(createT(7), 9)
	baseTail.AddTail(createT(8), 5)
	baseTail.AddTail(createT(100), 101)
	//   100    7
	//    1
	//   2
	// 3   4
	//    5
	//   8
	//  6

	if baseTail.Root.Pid != 100 {
		t.Errorf("TestTreeTail failed! (baseTail.Root.Pid=%#v)", baseTail.Root.Pid)
	}
	if baseTail.Root.Right.Pid != 7 {
		t.Errorf("TestTreeTail failed! (baseTail.Root.Right.Pid=%#v)", baseTail.Root.Right.Pid)
	}
	if baseTail.Root.Left.Pid != 1 {
		t.Errorf("TestTreeTail failed! (baseTail.Root.Left.Pid=%#v)", baseTail.Root.Left.Pid)
	}
	if baseTail.Root.Left.Left.Left.Right.Left.Left.Pid != 8 {
		t.Errorf("TestTreeTail failed! (baseTail.Root.Left.Left.Left.Right.Left.Left.Pid=%#v)", baseTail.Root.Left.Left.Left.Right.Left.Left.Pid)
	}
	if baseTail.Root.Left.Left.Left.Pid != 3 {
		t.Errorf("TestTreeTail failed! (baseTail.Root.Left.Left.Left.Pid=%#v)", baseTail.Root.Left.Left.Left.Pid)
	}
	if baseTail.Root.Left.Left.Left.Right.Left.Left.Left.Pid != 6 {
		t.Errorf("TestTreeTail failed! (baseTail.Root.Left.Left.Left.Right.Left.Left.Left.Pid=%#v)", baseTail.Root.Left.Left.Left.Right.Left.Left.Left.Pid)
	}
	if baseTail.Root.Deepth != 6 {
		t.Errorf("TestTreeTail failed! (baseTail.Root.Deepth=%#v)", baseTail.Root.Deepth)
	}
	if baseTail.Root.Left.Left.Left.Right.Deepth != 3 {
		t.Errorf("TestTreeTail failed! (baseTail.Root.Left.Left.Left.Right.Deepth=%#v)", baseTail.Root.Left.Left.Left.Right.Deepth)
	}
	if baseTail.Root.Right.Deepth != 0 {
		t.Errorf("TestTreeTail failed! (baseTail.Root.Right.Deepth=%#v)", baseTail.Root.Right.Deepth)
	}
	if baseTail.Root.Left.Left.Left.Right.Left.Deepth != 2 {
		t.Errorf("TestTreeTail failed! (baseTail.Root.Left.Left.Left.Right.Left.Deepth=%#v)", baseTail.Root.Left.Left.Left.Right.Left.Deepth)
	}

	buf := new(bytes.Buffer)
	baseTail.Serialize(buf)
	baseTail2 := new(BaseTail)
	baseTail2.Deserialize(buf)
	baseTail.Root.CleanUnexport()
	baseTail.Root.Deepth = 0
	baseTail2.Root.Deepth = 0
	if !reflect.DeepEqual(baseTail2.Root, baseTail.Root) {
		t.Errorf("\nTestBody want: %v,\nBut got:       %v", baseTail.Root, baseTail2.Root)
	}
}
func (s *TailSubject) CleanUnexport() {
	if s == nil {
		return
	}
	s.ppid = 0
	//s.Deepth = 0
	s.Left.CleanUnexport()
	s.Right.CleanUnexport()
}

//TODO: 调整为自动化测试random%30；判断父子节点及循环节点
func TestTree2(t *testing.T) {
	var baseTail *BaseTail
	baseTail = new(BaseTail)
	//pid不能为0
	baseTail.AddTail(createT(1), 100)
	baseTail.AddTail(createT(2), 1)
	//baseTail.AddTail(createT(4), 2)
	baseTail.AddTail(createT(3), 2)
	baseTail.AddTail(createT(5), 4)
	baseTail.AddTail(createT(6), 8)
	baseTail.AddTail(createT(10), 7)
	baseTail.AddTail(createT(7), 100)
	baseTail.AddTail(createT(8), 5)
	baseTail.AddTail(createT(100), 9)
	//      7
	//    100
	//    1
	//   2
	// 3   4
	//    5
	//   8
	//  6
	fmt.Printf("tree: %v\n", baseTail.Root)
	blank = make([]byte, 10)
	TransDeepth(baseTail.Root, 0)
	WriteTree(baseTail.Root)
}

func (s *TailSubject) String() string {
	if s == nil {
		return "nil"
	}
	return fmt.Sprint("Pid: ", s.Pid, " Deepth: ", s.Deepth, " Left:{ ", s.Left, " } Right:{ ", s.Right, " }")
}

//------------------------------------------------
//copy from oblog
var blank []byte
var w = os.Stdout

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
	w.WriteString("|──")

	bufferBody := fmt.Sprintf("%d(%d)", t.Pid, t.ppid)
	bufferBody += "\r\n"

	w.WriteString(bufferBody)
}
