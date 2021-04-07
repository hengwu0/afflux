/*算法描述：
就是按现实情况插入到多插树中，
若存在ptrace中的情况2，则缓存到dolists列表中
实时摘除节点并发送到blog.BaseChan中，并保证发送完所有子节点，再发送父节点
*/
package main

import (
	"Blog"
	"Callsys"
)

var UidCount int
var root Node
var dolists map[int]*list_t
var pusage map[int]*callsys.RetUsage

type list_t struct {
	data []callsys.PtraceRet
}

func (q *list_t) push(data callsys.PtraceRet) {
	q.data = append(q.data, data)
}

func (q *list_t) pop() (r callsys.PtraceRet) {
	r = q.data[0]
	q.data[0] = nil //gc free
	q.data = q.data[1:]
	return
}

func (q *list_t) len() int {
	return len(q.data)
}

type Node struct {
	alive     bool
	pid, ppid int
	uid       int //unique id
	cmd       *blog.Cmd
	childs    map[int]*Node //key不可用会重复的pid，要用uid
	father    *Node
}

func isAlive(n *Node) bool { return n.alive }
func isKey(n *Node) bool   { return n.cmd != nil }
func newNode(f, c int) (new *Node) {
	UidCount++
	new = &Node{ppid: f, pid: c, alive: true, uid: UidCount}
	new.childs = make(map[int]*Node)
	return
}
func (node *Node) cmdAdd(r *callsys.CmdEnv) {
	buf := blog.Create(r.ReadCmdEnv())
	if node.cmd == nil {
		buf.Setuid(node.uid)
		node.cmd = buf
	} else {
		node.cmd.Add(buf)
	}
}

func Insert(r *callsys.Ship) {
	f, c := r.Getpid(), r.ReadShip()
	new := newNode(f, c)
	if !root.insert(new) {
		//创建新c节点失败，表明f节点还不存在，uid不是连续的，因为这里存在不插入
		//禁止使用UidCount--，避免uid重复
		if v, ok := dolists[f]; ok {
			v.push(r)
		} else {
			var tmp list_t
			tmp.push(r)
			dolists[f] = &tmp
		}
		return
	}
	//创建c新节点成功，输出dolist相关内容
	cleanlists(new, new.pid)
}
func (node *Node) insert(new *Node) bool {
	if node == nil {
		return false
	}
	if node.pid == new.ppid && isAlive(node) {
		node.childs[new.uid] = new
		new.father = node
		return true
	} else {
		for _, p := range node.childs {
			if p.insert(new) {
				return true
			}
		}
		return false
	}
}
func cleanlists(node *Node, c int) {
	if v, ok := dolists[c]; ok {
		delete(dolists, c)
		for v.len() > 0 {
			op := v.pop()
			switch r := op.(type) {
			case *callsys.Ship:
				new := newNode(c, r.ReadShip())
				node.insert(new)
				cleanlists(new, new.pid) //如果fork之前父进程没有执行exec，则认为该父进程无效
			case *callsys.CmdEnv:
				node.cmdAdd(r)
			case *callsys.RetUsage:
				pusage[r.Getpid()] = r
			case *callsys.RetVal:
				node.alive = false
				if node.cmd != nil {
					if v, ok := pusage[node.pid]; ok {
						node.cmd.SetUsage(v.ReadRetUsage(), v.ReadRetUsageAddition(), callsys.Clktck)
					}
					node.cmd.SetExitCode(r.ReadRetVal())
				}
				delete(pusage, node.pid)
				makeClean(node)
				c = 0 //已经脱节，还挂就挂到根节点去吧
			}
		}
	}
}

func cleanlistsAll() {
	for k, v := range dolists {
		if v.len() == 0 {
			delete(dolists, k)
			continue
		}
		new := newNode(root.pid, k)
		root.insert(new)
		cleanlists(new, new.pid)
	}

	makeCleanAll(&root)
}

func SetKey(r *callsys.CmdEnv) {
	key := r.Getpid()

	if !root.setkey(key, r) {
		if v, ok := dolists[key]; ok {
			v.push(r)
		} else {
			var tmp list_t
			tmp.push(r)
			dolists[key] = &tmp
		}
	}
}
func (node *Node) setkey(key int, r *callsys.CmdEnv) bool {
	if node == nil {
		return false
	}

	if node.pid == key && isAlive(node) {
		node.cmdAdd(r)
		return true
	} else {
		for _, p := range node.childs {
			if p.setkey(key, r) {
				return true
			}
		}
		return false
	}
}

func SetUsage(r *callsys.RetUsage) {
	pusage[r.Getpid()] = r
}

func Delete(r *callsys.RetVal) {
	x := r.Getpid()
	if !root.del(x, r) {
		if v, ok := dolists[x]; ok {
			v.push(r)
		} else {
			var tmp list_t
			tmp.push(r)
			dolists[x] = &tmp
		}
	}
}
func (node *Node) del(x int, r *callsys.RetVal) bool {
	if node == nil {
		return false
	}

	if node.pid == x && isAlive(node) {
		node.alive = false
		//移除节点
		if node.cmd != nil {
			if v, ok := pusage[node.pid]; ok {
				node.cmd.SetUsage(v.ReadRetUsage(), v.ReadRetUsageAddition(), callsys.Clktck)
			}
			node.cmd.SetExitCode(r.ReadRetVal())
		}
		delete(pusage, node.pid)
		makeClean(node)
		return true
	} else {
		for _, p := range node.childs {
			if p.del(x, r) {
				return true
			}
		}
		return false
	}
}
func makeClean(x *Node) {
	if isAlive(x) || len(x.childs) > 0 {
		return //stop the clean
	}
	if isKey(x) {
		cmd := x.cmd
		cmd.Setpuid(query(x))
		blog.BaseChan <- cmd
		x.cmd = nil //for gc

	}
	delete(x.father.childs, x.uid)
	makeClean(x.father)
	x.father = nil //for gc
	return
}

func makeCleanAll(x *Node) {
	if x != &root {
		x.alive = false
	}
	makeClean(x)
	for _, child := range x.childs {
		makeCleanAll(child)
	}
}

func query(node *Node) int {
	if isKey(node.father) {
		return node.father.uid
	}
	return query(node.father)
}
