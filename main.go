package main

import (
	"flag"
	"hdfs-mount/config"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strings"
)

const MOUNT_OPTS = `
Mount options
fd=N
The file descriptor to use for communication between the userspace filesystem and the kernel. The file descriptor must have been obtained by opening the FUSE device (‘/dev/fuse’).

rootmode=M
The file mode of the filesystem’s root in octal representation.

user_id=N
The numeric user id of the mount owner.

group_id=N
The numeric group id of the mount owner.

default_permissions
By default FUSE doesn’t check file access permissions, the filesystem is free to implement its access policy or leave it to the underlying file access mechanism (e.g. in case of network filesystems). This option enables permission checking, restricting access based on file mode. It is usually useful together with the ‘allow_other’ mount option.

allow_other
This option overrides the security measure restricting file access to the user mounting the filesystem. This option is by default only allowed to root, but this restriction can be removed with a (userspace) configuration option.

max_read=N
With this option the maximum size of read operations can be set. The default is infinite. Note that the size of read requests is limited anyway to 32 pages (which is 128kbyte on i386).

blksize=N
Set the block size for the filesystem. The default is 512. This option is only valid for ‘fuseblk’ type mounts.
`

func main() {
	var confFile string
	flag.StringVar(&confFile, "config", "", "配置文件")
	var ops string
	// -o debug,rw,allow_other
	flag.StringVar(&ops, "o", "", "-o fd=N,rootmode=M,user_id=N,group_id=N,default_permissions,"+
		"allow_other,max_read=N,blksize=N\n"+MOUNT_OPTS)
	flag.Parse()
	if confFile == "" {
		flag.PrintDefaults()
		os.Exit(-1)
	}
	// 加载配置
	var err error
	conf, err := config.LoadConfig(confFile)
	if err != nil {
		os.Exit(1)
	}
	// 命令函参数优先于配置文件
	if ops != "" {
		conf.Mount.Options = strings.Split(ops, ",")
	}

	if conf.App.MaxNumCPU > 0 {
		runtime.GOMAXPROCS(conf.App.MaxNumCPU)
	}
	start(conf)
}
