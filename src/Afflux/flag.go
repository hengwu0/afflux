package main

import (
	"Callsys"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

const O1string = `([^\w|]+|^)ld$|([^\w|]+|^)gcc$|([^\w|]+|^)cc$|([^\w|]+|^)$g\+\+|([^\w|]+|^)$c\+\+|([^\w|]+|^)$ldd`
const O2string = `([^\w|]+|^)go$|([^\w|]+|^)java$|([^\w|]+|^)python$`
const O3string = `sh$`

var VlogOut, VdatOut, VdatIn string
var Verr bool
var Vids []int
var Vomits string
var Vadditon string
var Vsort string
var Vcmd []string

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [-o OutputFile] [-f CommandFile] \"PROG [ARGS]\" \n", os.Args[0])
	fmt.Fprintf(os.Stderr, "   or: %s -b dataBase [-p \"id1 id2...\"] [-omit '.*sh|?cc|ld']\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nStrace mode:\n")
	fmt.Fprintf(os.Stderr, "  -%s \t\t %s\n", "f", "Exec commands from the file")
	fmt.Fprintf(os.Stderr, "  -%s \t\t %s\n", "e", "Exit immediately while child process error exited")
	fmt.Fprintf(os.Stderr, "  -%s \t\t %s\n", "t", "Use SysTime in /proc instead of "+path.Base(os.Args[0])+" clocker")
	fmt.Fprintf(os.Stderr, "  -%s \t %s\n", "loder", "Specify a loder to exec cmd (Default: '/bin/sh -c')")
	fmt.Fprintf(os.Stderr, "  -%s \t\t %s\n", "add", "get addition msg from /proc/pid/status file with leftmost regexp")
	fmt.Fprintf(os.Stderr, "  \t \t\t %s\n", `example1: -add 'VmHWM:\s*(.*)'`)
	fmt.Fprintf(os.Stderr, "  \t \t\t %s\n", `example2: -add 'VmHWM:\s*(.*)|VmPeak:\s*(.*)'`)
	fmt.Fprintf(os.Stderr, "\nQuery mode:\n")
	fmt.Fprintf(os.Stderr, "  -%s \t\t %s\n", "b", "Specify a dataBase FileName")
	fmt.Fprintf(os.Stderr, "  -%s \t\t %s\n", "i", "Search \"ids\" in dataBaseFile for full message")
	fmt.Fprintf(os.Stderr, "\nCommon flags:\n")
	fmt.Fprintf(os.Stderr, "  -%s \t\t %s\n", "o", "Output message to file")
	fmt.Fprintf(os.Stderr, "  -%s \t %s\n", "omit", "omit programs in Output message with leftmost regexp")
	fmt.Fprintf(os.Stderr, "  -%s \t\t %s\n", "O1", "omit GNU compilers in Output message")
	fmt.Fprintf(os.Stderr, "  -%s \t\t %s\n", "O2", "omit GNU/go compilers in Output message")
	fmt.Fprintf(os.Stderr, "  -%s \t\t %s\n", "O3", "omit compilers/*sh in Output message")
	fmt.Fprintf(os.Stderr, "  -%s \t %s\n", "sort", "sort programs in Output message with `id[^/v]/time[^/v]/add[^/v]` (default with id)")
	fmt.Fprintf(os.Stderr, "  \t \t\t %s\n", "example: -sort time^")
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(2)
}

func flagParse() int {
	var fcmd, ids, fout, bin, omits, addition, sort, loder string
	var O1, O2, O3, t bool
	flag.StringVar(&bin, "b", "", "Specify a dataBase `filename`")
	flag.StringVar(&ids, "i", "", "Search `\"ids\"` in DataFile for full message")
	flag.StringVar(&fout, "o", "", "Output message to `file`")
	flag.StringVar(&fcmd, "f", "", "Exec commands from the `file`")
	flag.StringVar(&omits, "omit", "", "omit programs in Output message with `regexp`")
	flag.StringVar(&sort, "sort", "", "sort programs in Output message with `id[^/v]/time[^/v]/add[^/v]`")
	flag.StringVar(&addition, "add", "", "get addition msg from /proc/pid/status file with leftmost `regexp`")
	flag.StringVar(&loder, "loder", "-1", "Specify `loder` to exec cmd (Default: '/bin/sh -c')")
	flag.BoolVar(&O1, "O1", false, "omit gcc/g++/ld/ldd programs")
	flag.BoolVar(&O2, "O2", false, "omit gcc/g++/ld/ldd/go/java/python/ programs")
	flag.BoolVar(&O2, "O3", false, "omit gcc/g++/ld/ldd/go/java/python/*sh programs")
	flag.BoolVar(&Verr, "e", false, "Exit immediately while child process error")
	flag.BoolVar(&t, "t", false, "Use SysTime in /proc instead of "+path.Base(os.Args[0]))
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() > 1 {
		//尝试将该arg自动移动到末尾，并再次解析
		origArgs := make([]string, len(flag.Args()))
		copy(origArgs, flag.Args())
		for i, v := range os.Args {
			if v == origArgs[0] {
				copy(os.Args[i:], os.Args[i+1:])
				os.Args[len(os.Args)-1] = v
				break
			}
		}
		flag.Parse()
		if flag.NArg() > 1 {
			fmt.Fprintf(os.Stderr, "Can Not parse: %v!\n", origArgs)
			fmt.Fprintf(os.Stderr, "Try '%s -h' for more information.\n", os.Args[0])
			return 0
		}
	}
	if (fcmd != "" || flag.NArg() == 1 || t || Verr || loder != "-1") &&
		(bin != "" || ids != "") {
		fmt.Fprintf(os.Stderr, "Do Not mix two modes!\n")
		fmt.Fprintf(os.Stderr, "Try '%s -h' for more information.\n", os.Args[0])
		return 0
	}

	//可选项处理
	callsys.GetSysTimeUsage = t
	if sort != "" {
		switch sort {
		case "id^", "id":
			Vsort = "id"
		case "idv":
			Vsort = "idv"
		case "time", "time^":
			Vsort = "time"
		case "timev":
			Vsort = "timev"
		case "add", "add^":
			Vsort = "add"
		case "addv":
			Vsort = "addv"
		default:
			fmt.Fprintf(os.Stderr, "Can't parse sort arg: %s\n", sort)
			fmt.Fprintf(os.Stderr, "Try '%s -h' for more information.\n", os.Args[0])
			return 0
		}
	}
	if addition != "" {
		if _, err := regexp.Compile(addition); err != nil {
			fmt.Fprintf(os.Stderr, "Can Not parse regexp: Compile( %s ): %v\n", addition, err.Error())
			return 0
		}
		Vadditon = addition
	}
	if O1 && omits != "" {
		Vomits = O1string + "|" + omits
	} else if O1 {
		Vomits = O1string
	}
	if O2 && omits != "" {
		Vomits = O1string + "|" + O2string + "|" + omits
	} else if O2 {
		Vomits = O1string + "|" + O2string
	}
	if O3 && omits != "" {
		Vomits = O1string + "|" + O2string + "|" + O3string + "|" + omits
	} else if O3 {
		Vomits = O1string + "|" + O2string + "|" + O3string
	}
	if Vomits != "" {
		if _, err := regexp.Compile(Vomits); err != nil {
			fmt.Fprintf(os.Stderr, "Can Not parse regexp: Compile( %s ): %v\n", omits, err.Error())
			return 0
		}
	}
	if Verr {
		fmt.Printf("WARNING: %s will EXIT immediately while child process error!\n", path.Base(os.Args[0]))
	}
	if fout != "" {
		VlogOut = fout
		VdatOut = strings.TrimSuffix(fout, path.Ext(fout)) + ".dat"
	} else {
		VdatOut = "afflux.dat"
	}

	//必选项处理
	//mode:
	//1:Strace mode
	//2:Query mode, output log
	//3:Query mode, query ids
	Vcmd = make([]string, 0, 3)
	if fcmd != "" && flag.NArg() == 0 {
		if _, err := ioutil.ReadFile(fcmd); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 0
		} else {
			if loder != "-1" {
				Vcmd = append(strings.Fields(loder), fcmd)
			} else {
				Vcmd = append([]string{"/bin/sh"}, fcmd)
			}
			return 1
		}
	} else if fcmd == "" && flag.NArg() == 1 {
		if loder != "-1" {
			Vcmd = append(strings.Fields(loder), flag.Arg(0))
		} else {
			Vcmd = append([]string{"/bin/sh"}, "-c", flag.Arg(0))
		}
		return 1
	} else if fcmd != "" && flag.NArg() == 1 {
		fmt.Fprintf(os.Stderr, "%s: Do Not mix '-f %s' with \"PROG [ARGS]\"!\n", path.Base(os.Args[0]), fcmd)
		fmt.Fprintf(os.Stderr, "Try '%s -h' for more information.\n", os.Args[0])
		return 0
	}

	if ids != "" && bin != "" {
		VdatIn = bin
		tmp := strings.Fields(ids)
		Vids = make([]int, 0, len(tmp)+1)
		for _, v := range tmp {
			if k, err := strconv.Atoi(v); err != nil {
				usage()
				return 0
			} else {
				Vids = append(Vids, k)
			}
		}
		if sort != "" {
			fmt.Fprintf(os.Stderr, "WARNING: -sort '%s' ignored!!\n", sort)
		}
		return 3
	} else if ids == "" && bin != "" {
		VdatIn = bin
		return 2
	}

	if fcmd == "" && flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "%s: must have \"PROG [ARGS]\"!\n", path.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Try '%s -h' for more information.\n", os.Args[0])
		return 0
	}

	fmt.Fprintf(os.Stderr, "Can Not parse: %v!!!\n", os.Args)
	fmt.Fprintf(os.Stderr, "Try '%s -h' for more information.\n", os.Args[0])
	return 0
}
