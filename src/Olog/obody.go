package olog

import (
	"Blog"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const OutputHead = `
explain: ───(1234:5ms:OK:add) {/bin/sh -c ./example} {./example}
               │   │  │   │   └──exec commands
               │   │  │   └──addition msg(if had)
               │   │  └──exit status
               │   └──spend time
               └──unique id

Output message:
`

var bufferBody []byte
var blank []byte
var reg *regexp.Regexp

func OutputLog(omit string) {
	if omit != "" {
		reg = regexp.MustCompile(omit)
	}
	WriteOutputHead()

	if baseTail.Root != nil {
		blank = make([]byte, baseTail.Root.Deepth+1)
	}

	TransDeepth(baseTail.Root, 0)
	w.WriteString("{.\r\n")
	WriteTree(baseTail.Root)
}

//将deepth由树高度值变成深度值
//子节点深度+1；兄弟节点深度不变
func TransDeepth(t *blog.TailSubject, val uint32) {
	if t == nil {
		return
	}
	t.Deepth = val
	TransDeepth(t.Left, val+1)
	TransDeepth(t.Right, val)
}

func WriteTree(t *blog.TailSubject) {
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

func WriteOutputHead() {
	w.WriteString(OutputHead)
}

func getRet(st uint16, sig uint16) string {
	status := int16(st)
	signal := int16(sig)
	if status == 0 {
		return "OK"
	}
	if signal != -1 {
		return fmt.Sprintf("Killed by signal,%d", signal)
	}
	return fmt.Sprintf("ERROR,%d", status)
}

func writestring(b []byte, s string) []byte {
	b = append(b, s...)
	return b
}

func WriteBody(t *blog.TailSubject) {
	bufferBody = bufferBody[:0]
	s := baseBody.ReadSubject(logs[t.Index:])
	head := baseHead.ReadString(logs[offHead+int64(s.Cmd):])
	if reg != nil && reg.FindStringIndex(head) != nil {
		t.Left = nil
	}

	var count int
	for i := uint32(0); i < t.Deepth; i++ {
		if blank[i] == 0 {
			bufferBody = writestring(bufferBody, "    ")
			count++
		} else {
			bufferBody = writestring(bufferBody, "│   ")
			count = 0
		}
	}
	switch {
	case t.Left == nil && t.Right != nil:
		bufferBody = writestring(bufferBody, "├─── ")
	case t.Left != nil && t.Right != nil:
		bufferBody = writestring(bufferBody, "├───{")
	case t.Left != nil && t.Right == nil:
		bufferBody = writestring(bufferBody, "└───{")
	case t.Left == nil && t.Right == nil:
		bufferBody = writestring(bufferBody, "└───")
		if count < 4 {
			tmp := ([]rune(string(bufferBody)))
			bufferBody = []byte(string(tmp[:len(tmp)-count]))
		}
		for count >= 0 {
			bufferBody = writestring(bufferBody, "}")
			count--
		}
	}

	bufferBody = writestring(bufferBody, fmt.Sprintf("(%d:", s.Pid))
	if time.Duration(s.Time) != -1 {
		bufferBody = writestring(bufferBody, fmt.Sprintf("%v:", time.Duration(s.Time)))
	} else {
		bufferBody = writestring(bufferBody, "OMITED:")
	}
	if s.Addition != "" {
		bufferBody = writestring(bufferBody, fmt.Sprintf("%s:%s) ", getRet(s.Status, s.Signal), s.Addition))
	} else {
		bufferBody = writestring(bufferBody, fmt.Sprintf("%s) ", getRet(s.Status, s.Signal)))
	}

	for i, v := range s.Args {
		if strings.ContainsRune(v, ' ') || len(v) == 0 {
			s.Args[i] = `'` + s.Args[i] + `'`
		}
	}
	if ss := strings.Join(s.Args, " "); strings.ContainsAny(ss, "\n\r") && head[0] != '{' { //判断\r或\n
		bufferBody = writestring(bufferBody, "{"+fmt.Sprintf("%s %s", head, ss)+"}")
	} else {
		bufferBody = writestring(bufferBody, fmt.Sprintf("%s %s", baseHead.ReadString(logs[offHead+int64(s.Cmd):]), ss))
	}
	bufferBody = writestring(bufferBody, "\r\n")

	w.Write(bufferBody)
}
