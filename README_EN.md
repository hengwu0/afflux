[![996.icu](https://img.shields.io/badge/link-996.icu-red.svg)](https://996.icu)

# Introduction
`Afflux` is a tool to trace the cmds and draw processes tree. It can trace x86 and x86_64 programs on GNU/Linux amd64 OS. It may also work fine on GNU/Linux x86 OS.

# Quick Start
Put the command in quotation marks, and pass it to `afflux`.

__e.g.__  
```
  [wuheng@show]$ afflux "gcc a.c -o A"

  explain: ───(1234:5ms:OK:add) {/bin/sh -c ./example} {./example}
                 │   │  │   │   └──exec commands
                 │   │  │   └──addition msg(if had)
                 │   │  └──exit status
                 │   └──spend time
                 └──unique id

  Output message:
  {.
  └───{(1:32.4ms:OK) gcc a.c -o A
      ├─── (2:7.5ms:OK) /usr/libexec/gcc/x86_64-redhat-linux/4.4.6/cc1 .....
      ├─── (3:3.78ms:OK) as -Qy -o /tmp/ccbGn3Df.o /tmp/ccws3MYE.s
      └───{(4:15ms:OK) /usr/libexec/gcc/x86_64-redhat-linux/4.4.6/collect2 .....
          └─}}}(5:12.8ms:OK) /usr/bin/ld --eh-frame-hdr ... ...
```
The `Output message` show the processes tree "`gcc a.c -o A`" command included: The gcc command evoked cc1,as,collect2 to do the compile; And then, The collect2 evoked ld to do the final link.

# What afflux can do
* When a process group is running, we can use `pstree` command to print the running processes as a tree. But, it can't show the processes which has exited! the `Afflux` can record the whole processes tree even it exited.
* It is unconvenient to grab some subcommands in shell scripts. Such as _an error gcc compile command_ in _Makefile_. And it is much more hard to replay the error without _shell env_.
* `Afflux` can help you to grap the err command and replay it easily.

# Full help
There is two mode of afflux, see `afflux -h`:  
```
Strace mode:
  -f             Exec commands from the file
  -e             Exit immediately while child process error exited
  -t             Use SysTime in /proc instead of afflux clocker
  -loder         Specify a loder to exec cmd (Default: '/bin/sh -c')
  -add           get addition msg from /proc/pid/status file with leftmost regexp
                         example: -add 'VmHWM:\s*(.*)'
Query mode:
  -b             Specify a dataBase FileName
  -i             Search "ids" in dataBaseFile for Full message

Common flags:
  -o             Output message to file
  -omit          omit programs in Output message with leftmost regexp
  -O1            omit GNU compilers in Output message
  -O2            omit GNU/go compilers in Output message
  -O3            omit compilers/*sh in Output message
  -sort          sort programs in Output message with `id[^/v]/time[^/v]/add[^/v]` (default with id)
                         example: -sort time^
```

　In _Strace mode_, it can trace any cmds and output the processes tree executed. `Quick Start` shows the simplest example of a gcc command. The commands will be puted in new line with `├───`, if there is an __fork__ system call. And they will be enclosed in curly braces, if there is an __exec__ system call. You may see there are two or more groups of curly braces in one line, which means this program __exec__ an new program without __fork__, so the pid will not change.  
　`-f file`: Exec commands from the _file_. Afflux will use `sh file` command to start work.  
　`-loder loder`: Specify a _loder_ to exec user's cmd (Afflux will default to use _loader_ `/bin/sh -c "command_string"`to start work).  
　`-e`: Afflux will check the return value of each processes traced. And abort all processes with afflux itself when the return value is not zero.  
　`-add regexp`: Afflux can get addition msg form /proc/pid/status file with leftmost _regexp_ when the progress finished but before it exit. So, it can help you to statistic data such as the Maximum memory footprint of the process. The addition msg will be showed in Output message. You shoud always use single quotation marks in case of shell's parsing.  
　Press ctrl-c can abort all processes with afflux itself immediately , and the `Output message` will be printed before quit.
　　   
　Afflux will record the process's Full messages into an afflux `dataBase file` when tracing, which can be used in _Query mode_.   
　　    
　In _Query mode_, it can print the Full messages of an process by ids.  
　`-b file`: Specify a dataBase _file_. Afflux will not do the trace, but open this dataBase and extract information to Output message.   
　`-i ids`: Search "ids" in dataBaseFile for Full message. You can search multiple ids together splited by blank space key/keys.  
      
　The _Common flags_ means the flags can be used in both _Strace mode_ and _Query mode_.  
　`-o file`: Output any messages to _file_ instead of terminal.  
　`omit regexp`: omit programs in Output message with leftmost regexp, so you can't see the processes tree evoked by this process. It will also let afflux __detach__ this program if afflux is doing tracing.  
　`-O1`: omit GNU compilers in Output message. Afflux will automatic use `--omit ([^\w|]+|^)ld$|([^\w|]+|^)gcc$|([^\w|]+|^)cc$|([^\w|]+|^)$g\+\+|([^\w|]+|^)$c\+\+|([^\w|]+|^)$ldd` to omit gcc/g++/ld/ldd programs. It does not conflict with the previous parameter, and they can work in the meantime.  
　`-O2`: omit GNU/go compilers in Output message. Afflux will automatic omit go/java/python on the basis of O1.  
　`-O3`: omit compilers/\*sh in Output message. Afflux will automatic omit any sh on the basis of O2.  
　`-sort id[^/v]/time[^/v]/add[^/v]`: sort programs in Output message with `id[^/v]/time[^/v]/add[^/v]` (default with id). If an process has multi-processes, Afflux can sort them by id, time, or addition msg(only if msg start of digital). And you can use `v` or `^` to specific ascending order or descending.  
  
### Where is `dataBase file`
The `dataBase file` will be auto created each time after _Strace mode_, or overwrite the old one. It's default name is _afflux.dat_. But if you use `-o` flag, the `dataBase file`'s name will also affected by the flag. It will use the file name and file path you specified, but modify the filename extension to `.dat`.

### How to use `dataBase file`
  Here shows the simple use of `dataBase file` to show processes's Full messages:  
__e.g.__  
```
  [wuheng@show]$ afflux -b afflux.dat -i "1 2 3"

  id(1)'s cmd:
  gcc a.c -o A
  
  id(1)'s cmd chain:
  .
  └───(1:32.4ms:OK) gcc a.c -o A
  
  id(1)'s cwd:
  /home/wuheng/test/show
  
  id(1)'s env:
  nil.
  
  
  id(2)'s cmd:
  /usr/libexec/gcc/x86_64-redhat-linux/4.4.6/cc1 -quiet a.c -quiet -dumpbase a.c -mtune=generic -auxbase a -o /tmp/ccws3MYE.s
  
  id(2)'s cmd chain:
  .
  └───(1:32.4ms:OK) gcc a.c -o A
      └───(2:7.5ms:OK) /usr/libexec/gcc/x86_64-redhat-linux/4.4.6/cc1 -quiet a.c -quiet -dumpbase a.c -mtune=generic -auxbase a -o /tmp/ccws3MYE.s
  
  id(2)'s cwd:
  /home/wuheng/test/show
  
  id(2)'s env:
  COLLECT_GCC=gcc
  COLLECT_GCC_OPTIONS=''\''-o'\'' '\''A'\'' '\''-mtune=generic'\'''
  
  
  id(3)'s cmd:
  as -Qy -o /tmp/ccbGn3Df.o /tmp/ccws3MYE.s
  
  id(3)'s cmd chain:
  .
  └───(1:32.4ms:OK) gcc a.c -o A
      └───(3:3.78ms:OK) as -Qy -o /tmp/ccbGn3Df.o /tmp/ccws3MYE.s
  
  id(3)'s cwd:
  /home/wuheng/test/show
  
  id(3)'s env:
  COLLECT_GCC=gcc
  COLLECT_GCC_OPTIONS=''\''-o'\'' '\''A'\'' '\''-mtune=generic'\'''
  
  
  Common environment variable:
  SHELL=/bin/bash
  HISTSIZE=1000
  LESSOPEN='|/usr/bin/lesspipe.sh %s'
  ...
  
  That's all the messages.
```
It is convenient to replay the command if you got the `cwd`. Some command also needs the `environment variables` potential, such as golang. In the above output, each command's env var is separated into two parts by afflux: `id's env` and `Common environment variable`. Afflux separate the env vars in two parts just for layout and presentation, you should not focus on just one of them.
  
# Bugs
* afflux use ptrace system call to trace the commands, which will let process runs slowly(About 20% Performance loss).
* Programs that use the setuid bit do not have effective user ID privileges while being traced.
