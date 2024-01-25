package olog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type CmdStatus struct {
	Id       int    `json:"id"`
	Time     uint64 `json:"time"`
	Ret      string `json:"ret"`
	Addition string `json:"add"`
}
type CmdObject struct {
	CmdStatus
	Cmd      string       `json:"cmd"`
	Children []*CmdObject `json:"children"`
}
type CmdDetail struct {
	Cmd       string     `json:"cmd"`
	CmdChain  *CmdObject `json:"children"`
	Cwd       string     `json:"cwd"`
	Env       []string   `json:"env"`
	CommonEnv []string   `json:"common_env"`
}

func Serve(ipport string) {
	http.HandleFunc("/get-childrencmds", getChildrenCmds)
	http.HandleFunc("/get-cmd-by-id", getCmdbyId)

	fmt.Printf("Server listening on %s...\n", ipport)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Errorf("Server error: %s\n", err)
	}
}

func getCmdbyId(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("error curred\n")
		}
	}()

	if r.Method != "GET" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("no support."))
		return
	}

	var id int
	var err error
	params := r.URL.Query()
	if params == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no support param."))
		return
	}

	if val, ok := params["id"]; ok {
		id, err = strconv.Atoi(val[0])
	} else {
		err = fmt.Errorf("no id param")
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no support param."))
		return
	}

	res := GetCmdDetail(id)
	resstr, _ := json.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resstr)
}

func GetCmdDetail(id int) *CmdDetail {
	Env := make(map[string]int)
	count := 0
	var searchDetail []int64
	res := new(CmdDetail)

	// 查1次env
	bufferBody = bufferBody[:0]
	if r := baseTail.Root.Search(uint32(id)); r != nil {
		s := baseBody.ReadSubject(logs[r[0]:])
		for _, e := range s.Env {
			env := baseHead.ReadString(logs[offHead+int64(e):])
			Env[env]++
			count = max(count, Env[env])
		}
		searchDetail = r
	} else {
		return nil
	}

	// 查一次root的env
	if r := baseTail.Root.Index; r != 0 {
		s := baseBody.ReadSubject(logs[r:])
		for _, v := range s.Env {
			env := baseHead.ReadString(logs[offHead+int64(v):])
			if _, ok := Env[env]; ok {
				Env[env]++
				count = max(count, Env[env])
			}
		}
		if s.Pid == uint32(id) {
			count++
		}
	}

	s := baseBody.ReadSubject(logs[searchDetail[0]:])
	for i, v := range s.Args {
		s.Args[i] = formatStr(v)
	}
	res.Cmd = fmt.Sprintf("%s %s", baseHead.ReadString(logs[offHead+int64(s.Cmd):]), strings.Join(s.Args, " "))

	l := len(searchDetail)
	tmp := new(CmdObject)
	beforeCmd := tmp
	for i := l - 1; i >= 0; i-- {
		if searchDetail[i] == 0 {
			continue
		}
		cmd := new(CmdObject)

		s := baseBody.ReadSubject(logs[searchDetail[i]:])
		cmd.Id = int(s.Pid)
		cmd.Time = s.Time
		cmd.Addition = s.Addition
		cmd.Ret = getRet(s.Status, s.Signal)

		for i, v := range s.Args {
			if strings.ContainsRune(v, ' ') || len(v) == 0 {
				s.Args[i] = `'` + s.Args[i] + `'`
			}
		}
		cmd.Cmd = fmt.Sprintf("%s %s", baseHead.ReadString(logs[offHead+int64(s.Cmd):]), strings.Join(s.Args, " "))
		cmd.Children = nil
		beforeCmd.Children = []*CmdObject{cmd}
		beforeCmd = cmd
	}
	res.CmdChain = tmp

	res.Cwd = baseHead.ReadString(logs[offHead+int64(s.Cwd):])

	res.Env = make([]string, 0, len(s.Env))
	for _, e := range s.Env {
		env := baseHead.ReadString(logs[offHead+int64(e):])
		if Env[env] >= count && len(env) > 2 {
			continue
		}
		res.Env = append(res.Env, formatStr(env))
	}

	res.CommonEnv = make([]string, 0, len(s.Env))
	for k, v := range Env {
		if v >= count && len(k) > 2 {
			res.CommonEnv = append(res.CommonEnv, formatStr(k))
		}
	}

	return res
}

func getChildrenCmds(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("error curred\n")
		}
	}()

	if r.Method != "GET" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("no support."))
		return
	}

	var id, depth int
	var err error
	params := r.URL.Query()
	if params == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no support param."))
		return
	}

	if val, ok := params["id"]; ok {
		id, err = strconv.Atoi(val[0])
	} else {
		err = fmt.Errorf("no id param")
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no support param."))
		return
	}

	if val, ok := params["depth"]; ok {
		depth, err = strconv.Atoi(val[0])
	} else {
		err = fmt.Errorf("no depth param")
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no support param."))
		return
	}

	res := searchChildrenCmds(id, depth)
	resstr, _ := json.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resstr)
}

func searchChildrenCmds(id int, depth int) []*CmdObject {
	if depth == 0 {
		return nil
	} else {
		depth--
	}

	if r := baseTail.Root.SearchChildren(uint32(id)); r != nil {
		res := make([]*CmdObject, 0, len(r))
		for i := range r {
			if r[i] == 0 {
				continue
			}
			cmd := new(CmdObject)
			s := baseBody.ReadSubject(logs[r[i]:])
			cmd.Cmd = fmt.Sprintf("%s %s", baseHead.ReadString(logs[offHead+int64(s.Cmd):]), strings.Join(s.Args, " "))
			cmd.Ret = getRet(s.Status, s.Signal)
			cmd.Addition = s.Addition
			cmd.Id = int(s.Pid)
			cmd.Time = s.Time
			cmd.Children = searchChildrenCmds(cmd.Id, depth)
			res = append(res, cmd)
		}
		return res
	}
	return nil
}
