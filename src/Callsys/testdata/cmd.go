package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	//"time"
)

func main() {
	fmt.Printf("pid= %d\n", os.Getpid())
	os.Chdir("..")
	//signal.Notify(c, syscall.SIGTTOU, syscall.SIGTTIN)
	//go signalProcess()
	Process()
	// var s string
	// Scanf(&s)
	// fmt.Println("exited!")
}

func Scanf(a *string) {
	fmt.Println("Please input a string:")
	data := make([]byte, 10)
	//os.Stdin.Read(data)
	*a = string(data)
	fmt.Println(*a)
}

var c = make(chan os.Signal)

//SIGTTIN,SIGTTOU信号处理
func Process() {

	cmd := exec.Command("ls", "-alh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Start()
	fmt.Printf("%s [%d] background running...\n", os.Args[0], cmd.Process.Pid)
	//time.Sleep(time.Second*10)
	os.Exit(4)
}

func signalProcess() {
	for {
		s := <-c
		switch s {
		case syscall.SIGTTIN, syscall.SIGTTOU:
			fmt.Println("Switched to background!")
			signal.Ignore(s)
			Process()
		default:
			fmt.Printf("Recv signal:%#v\n", s)
		}
	}
}
