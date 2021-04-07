package olog

import (
	"fmt"
	"strings"
	"time"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

//查两遍，并不会减少多少程序性能
func Query(pids []int) {
	Env := make(map[string]int)
	count := 0
	for _, id := range pids {
		bufferBody = bufferBody[:0]
		if r := baseTail.Root.Search(uint32(id)); r != nil {
			s := baseBody.ReadSubject(logs[r[0]:])
			for _, e := range s.Env {
				env := baseHead.ReadString(logs[offHead+int64(e):])
				Env[env]++
				count = max(count, Env[env])
			}
		}
	}
	if len(pids) == 1 {
		if r := baseTail.Root.Search(1); r != nil {
			s := baseBody.ReadSubject(logs[r[0]:])
			for _, v := range s.Env {
				env := baseHead.ReadString(logs[offHead+int64(v):])
				if _, ok := Env[env]; ok {
					Env[env]++
					count = max(count, Env[env])
				}
			}
		}
	}
	if len(pids) == 1 && pids[0] == 1 {
		count++
	}

	for _, id := range pids {
		bufferBody = bufferBody[:0]
		if r := baseTail.Root.Search(uint32(id)); r == nil || r[0] == 0 {
			bufferBody = writestring(bufferBody, fmt.Sprintf("Can't find id = %d!", id))
			bufferBody = writestring(bufferBody, "\r\n")
		} else {

			s := baseBody.ReadSubject(logs[r[0]:])
			bufferBody = writestring(bufferBody, fmt.Sprintf("\r\nid(%d)'s cmd:\r\n", id))
			head := baseHead.ReadString(logs[offHead+int64(s.Cmd):])

			for i, v := range s.Args {
				s.Args[i] = formatStr(v)
			}
			if ss := strings.Join(s.Args, " "); strings.ContainsAny(ss, "\n\r") && head[0] != '{' { //判断\r或\n
				bufferBody = writestring(bufferBody, "{"+fmt.Sprintf("%s %s", head, ss)+"}")
			} else {
				bufferBody = writestring(bufferBody, fmt.Sprintf("%s %s", baseHead.ReadString(logs[offHead+int64(s.Cmd):]), ss))
			}
			bufferBody = writestring(bufferBody, "\r\n")

			bufferBody = writestring(bufferBody, fmt.Sprintf("\r\nid(%d)'s cmd chain:\r\n", id))
			l := len(r)
			bufferBody = writestring(bufferBody, ".\r\n")
			for i := l - 1; i >= 0; i-- {
				if r[i] == 0 {
					continue
				}
				s := baseBody.ReadSubject(logs[r[i]:])
				for j := 0; j < l-i-1; j++ {
					bufferBody = writestring(bufferBody, "    ")
				}
				bufferBody = writestring(bufferBody, "└───")

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

				head := baseHead.ReadString(logs[offHead+int64(s.Cmd):])
				for i, v := range s.Args {
					if strings.ContainsRune(v, ' ') || len(v) == 0 {
						s.Args[i] = `'` + s.Args[i] + `'`
					}
				}
				if ss := strings.Join(s.Args, " "); strings.ContainsAny(ss, "\n\r") && head[0] != '{' { //判断\r或\n
					bufferBody = writestring(bufferBody, fmt.Sprintf("{%s %s}", head, ss))
				} else {
					bufferBody = writestring(bufferBody, fmt.Sprintf("%s %s", head, ss))
				}
				bufferBody = writestring(bufferBody, "\r\n")
			}

			bufferBody = writestring(bufferBody, "\r\n")
			bufferBody = writestring(bufferBody, fmt.Sprintf("id(%d)'s cwd:", id))
			bufferBody = writestring(bufferBody, "\r\n")
			cwd := baseHead.ReadString(logs[offHead+int64(s.Cwd):])
			bufferBody = writestring(bufferBody, cwd)
			bufferBody = writestring(bufferBody, "\r\n")

			bufferBody = writestring(bufferBody, "\r\n")
			bufferBody = writestring(bufferBody, fmt.Sprintf("id(%d)'s env:", id))
			bufferBody = writestring(bufferBody, "\r\n")
			flag := true
			for _, e := range s.Env {
				env := baseHead.ReadString(logs[offHead+int64(e):])
				if Env[env] >= count && len(env) > 2 {
					continue
				}
				flag = false
				env = formatStr(env)
				bufferBody = writestring(bufferBody, env)
				bufferBody = writestring(bufferBody, "\r\n")
			}
			if flag {
				bufferBody = writestring(bufferBody, "nil.\r\n")
			}
			bufferBody = writestring(bufferBody, "\r\n")
		}
		w.Write(bufferBody)
	}
	bufferBody = bufferBody[:0]
	for k, v := range Env {
		if v >= count && len(k) > 2 {
			k = formatStr(k)
			bufferBody = writestring(bufferBody, k)
			bufferBody = writestring(bufferBody, "\r\n")
		}
	}
	if len(bufferBody) > 0 {
		w.Write([]byte("\r\nCommon environment variable:\r\n"))
	}
	bufferBody = writestring(bufferBody, "\r\nThat's all the messages.\r\n")
	w.Write(bufferBody)
}

func count(src []byte, b byte) (r int) {
	for i := 0; i < len(src); i++ {
		if src[i] == b {
			r++
		}
	}
	return
}

func formatStr(s string) string {
	str := []byte(s)
	a := str
	var b []byte
	for i, v := range str {
		if v == '=' {
			a = str[:i]
			b = str[i+1:]
			return string(formatstr(a)) + "=" + string(formatstr(b))
		}
	}
	return string(formatstr(a))
}

func formatstr(str []byte) string {
	if len(str) == 0 {
		return "''"
	}
	constom := []byte("*$;|&#!$><\"`%()[]{}^\\'")
	countAll := 0
	for i := 0; i < len(constom); i++ {
		countAll += count(str, constom[i])
	}
	if count(str, ' ') == 0 {
		if countAll == 0 {
			return string(str)
		} else if countAll < 5 {
			res := make([]byte, 0, len(str)+countAll*2+1)
			for i := 0; i < len(str); i++ {
				if count(constom, str[i]) > 0 {
					res = append(res, '\\', str[i])
				} else {
					res = append(res, str[i])
				}
			}
			return string(res)
		}
	}

	n := count(str, '\'')
	res := make([]byte, 0, len(str)+countAll*2+n*3+1)
	if n != 0 {
		for i := 0; i < len(str); i++ {
			if str[i] == '\'' {
				res = append(res, '\'', '\\', '\'', '\'')
			} else {
				res = append(res, str[i])
			}
		}
	} else {
		res = str
	}
	return `'` + string(res) + `'`
}
