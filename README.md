[![996.icu](https://img.shields.io/badge/link-996.icu-red.svg)](https://996.icu)

# 引言
`Afflux`是一个用于跟踪用户命令执行过程的工具。他能跟踪x86、x86_64程序的执行过程，并记录下整个子命令树。该程序目前仅在GNU/Linux系统上测试通过。

# 快速上手
在终端中，在原始命令的执行路径下，将原始的命令使用双引号包含起来，并传递给`afflux`即可。

__例如：__  
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
`Output message`展示出命令"`gcc a.c -o A`"在执行过程中，包含的子命令，并以进程树的方式呈现。可以看到，gcc调用了cc1,as,collect2 完成整个编译过程，并且最后由collect2调用ld链接器执行链接。

# afflux有什么用
* 当一个进程组内进程都在执行时，我们可以使用`pstree`命令来打印整个进程树。但是对于那些已经执行结束的子进程它却无能为力。`Afflux`能够记录下整个子命令树，以及子命令的一些__相关__信息，以便用户可以方便的单独执行该子命令。
* 在shell脚本中，想要调试工作异常的脚本，是一件非常头疼和麻烦的事情。因为对于庞大的脚本内容，往往很难快速的定位到出错的语句。例如在一个庞大的Makefile中，父Makefile还会调用子Makefile，如果某个gcc编译命令出错了，你很难能单独复现这条gcc编译命令，而且你很可能需要花费很大的时间、精力先去学习和了解整个Makefile的编译框架，然后才再能去解决这gcc编译故障。
* `Afflux`能够帮助你抓取到这条出错的gcc编译命令，并让你能够立马复现该编译错误。从而无需对它的Makefile的编译框架有任何关心。

# 完全使用说明
在afflux中，有两种模式： _Strace mode_ 、 _Query mode_ 。如下展示了`afflux -h`命令:  
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

　在 _Strace mode_ 下，它能跟踪任何命令的执行过程，并且导出命令执行树。在__快速上手__章节中已经展示了一个gcc的简单示例，它列出了gcc命令的命令执行树。所有子命令会另起一行，并以`├───`开始，而且会以空格来表示该子命令的层级。若某命令采用的是通过 _exec_ 系统调用启动的新进程，而没有创建出新子进程来，则afflux会将两条命令显示在同一行，但会分别以大括号“{}”括起来，以表示他们是同一个进程，但里面包含了两条命令。在linux中，新的子进程是由 _fork_ 系统调用创建的，而装载新命令是由 _exec_ 系统调用实现的。afflux监控到 _fork_ 系统调用，便知道了创建子进程；监控到 _exec_ 系统调用，便知道了执行了新子命令。若某父进程只有fork创建子进程而没有exec执行新子命令，则afflux不会对子进程做任何记录，你只能看到该父进程的信息。  
　_Strace mode_ 参数：  
　`-f file`: 从指定的 _file_ 文件中执行初始命令. Afflux会使用`sh file`的方式来启动初始命令。  
　`-loder loder`: 指定执行用户命令的加载器 _loder_(Afflux默认使用`/bin/sh -c "command_string"`的方式来执行初始命令).  
　`-e`: Afflux将检查每条子命令的退出码，若检查到任意一个子命令的退出码非0，则立即终止所有子程序的执行。(afflux仍然会输出命令树)  
　`-add regexp`: Afflux可以帮助你在任何子进程将要结束时，从/proc/pid/status文件中，根据leftmost _regexp_ 准则抓取额外信息并记录下来。所以afflux还能让你轻松的获取任何子进程的统计数据，例如`-add 'VmHWM:\s*(.*)'`参数将抓取子进程的最大物理内存使用量。抓取到的信息会直接展示在`Output message`的子命令头部`(1234:5ms:OK:add)`中的add部分。若不添加`-add`参数，子命令头部则不会显示add部分。注意，你最好始终使用单引号将正则表达式引起来，以避免被shell解析字符。  
　按ctrl-c快捷键可以立即终止所有子进程，并立即打印命令树退出。  
　　   
　Afflux会记录子命令的所有相关信息到数据文件`dataBase file`中，这些信息可以在 _Query mode_ 展示并使用。   
　　    
　在 _Query mode_ 下，afflux可以帮助你根据输入的id，从数据文件中读取并打印出该子命令的完整信息，以协助你可以单独执行该子命令。若不输入-i参数，则会打印`Output message`。  
　_Query mode_ 参数：  
　`-b file`: 指定dataBase _file_ 文件。Afflux不会启动跟踪命令，而是打开数据文件并解析数据。  
　`-i ids`: 指定需要搜索的id，如有多个id请用引号包含，并空格分开。  
      
　其他参数(_Common flags_)：可以在 _Strace mode_ 和 _Query mode_ 同时使用的参数  
　`-o file`: 将Afflux的任何输出信息导出到 _file_ 文件中，而不是在终端显示  
　`omit regexp`: 忽略某些命令(根据leftmost _regexp_ 准则)，不显示该命令的所有子命令(即折叠不显示)，但该命令仍会记录。当Afflux在 _Strace mode_ 下时，会直接 __detach__ 符合正则的子进程。  
　`-O1`: 忽略GNU compilers相关程序的跟踪。Afflux会自动添加`--omit ([^\w|]+|^)ld$|([^\w|]+|^)gcc$|([^\w|]+|^)cc$|([^\w|]+|^)$g\+\+|([^\w|]+|^)$c\+\+|([^\w|]+|^)$ldd`来不展开gcc/g++/ld/ldd程序的执行过程。它不会对--omit参数产生干扰，并且两参数可同时生效。  
　`-O2`: 忽略GNU/go compilers相关程序的跟踪。Afflux会在-O1的基础上忽略go/java/python程序的执行展开。  
　`-O3`: 忽略compilers/\*sh相关程序的跟踪。Afflux会在-O2的基础上忽略任何sh命令的执行展开。  
　`-sort id[^/v]/time[^/v]/add[^/v]`: 将`Output message`中的命令按`id[^/v]/time[^/v]/add[^/v]`进行排序(默认根据id升序)，你也可以通过v或者^来指定升序或者降序排序，如`-sort id^`。需要注意的是，父子进程的缩进关系不会被打乱，其仅会对同一级的子进程进行排序(同一级子进程指由一个父进程直接衍生出的多个子进程)。  
  
### 如何产生`dataBase file`
`dataBase file`会在Afflux的 _Strace mode_ 时自动生成，若存在旧文件，则自动覆盖旧文件。它的默认名称为 _afflux.dat_ ，但是如果你指定了-o参数，则`dataBase file`的名称也会自动改变，它会使用你指定的文件名称(含路径)，并将文件扩展名调整为`.dat`。如`-o make.log`，则数据文件名为make.dat；如`-o log`，则数据文件名为log.dat。

### 如何使用`dataBase file`
  如下展示了`dataBase file`使用的简单示例，并展示出查询到的完整信息(可用于单独执行子命令)：  
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
从上述子命令的完整信息中，可以得到cmd、cmd chain、cwd、env等信息，这对于单独执行该子命令是十分必要的。其中，cwd指current work directory，表示该命令执行前的所在目录；cmd指可以直接黏贴到终端的完整子命令(含参数)；env指该程序执行前需准备的环境变量。若需单独执行该子命令，可以通过如下几部来执行：  
　1、cd到cwd目录：有些程序如gcc等其源文件可能通过相对路径方式传入，故必须调整到cwd路径下才能正常执行；  
　2、配置env：有些程序如golang等运行时会依赖环境变量来传入参数，故必须配置相关的环境变量值；  
　3、粘贴cmd来执行该子命令：Afflux会将命令行进行添加引号及反转义，以保证你可以直接黏贴执行(对于含空格的参数，会用单引号括起来；对于特殊字符，使用\\反转义)  
　注意，在上述输出中，env已被Afflux拆分为两组：`id's env`和`Common environment variable`。`id's env`指该id命令的env相比1号id的env不同部分，而`Common environment variable`指与1号id的env相同部分，两组env组合才是该id的全部环境变量。通常情况下，1号id的env为当前shell的env，分开显示有益于排版和显示。  
  
# Bugs
* `Afflux使用 _ptrace_ 系统调用，其会使整个命令执行时间变长(约20%性能损失)。
* 跟踪使用setuid位的程序时，会提示没有有效的用户ID权限。对于跟踪sudo命令时，需要将sudo提前。比如跟踪sudo sh make.sh，不能这样执行afflux "sudo sh make.sh"，需要这样执行：sudo afflux "sh make.sh"
