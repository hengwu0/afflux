package main

import (
	"fmt"
	"regexp"
)

const Vomit = `sh$`

func main() {
	reg := regexp.MustCompile(Vomit)
	head := "/bin/sh"
	r := reg.FindStringIndex(head)
	fmt.Printf("reg: %#v\n", r)
}
