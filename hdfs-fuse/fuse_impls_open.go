package hdfs_fuse

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	log "github.com/sirupsen/logrus"
	"strings"
	"syscall"
)

var _ = (fs.NodeOpener)((*Node)(nil))

func (n *Node) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	log.Infof("* Open file:%s flags:%d", n.path, flags)

	// n.test(ctx, flags)

	fullPath := n.makeNodePath(n.path, "")
	_, isDir, err := n.hdfsClient.Stat(fullPath)
	if err != nil {
		log.Errorf("stat hdfs file error,file:%s err:%v", n.path, err)

		if strings.LastIndex(err.Error(), "file does not exist") >= 0 {
			return nil, fuse.FOPEN_DIRECT_IO, syscall.ENOENT
		}
		return nil, fuse.FOPEN_DIRECT_IO, syscall.EIO
	}
	if isDir {
		return nil, fuse.FOPEN_DIRECT_IO, syscall.EISDIR
	}
	fileHandle := NewFileHandle(&n.path, nil, nil, n.hasWrite(flags), n.hasRead(flags))
	n.fileHandle = append(n.fileHandle, fileHandle)

	return fileHandle,fuse.FOPEN_DIRECT_IO, fs.OK
}

/**
O_ACCMODE <0003>;: 读写文件操作时，用于取出flag的低2位。
O_RDONLY<00>;: 只读打开
O_WRONLY<01>;: 只写打开
O_RDWR<02>;: 读写打开
O_CREAT<0100>;: 文件不存在则创建，需要mode_t，not fcntl
O_EXCL<0200>;: 如果同时指定了O_CREAT，而文件已经存在，则出错， not fcntl
O_NOCTTY<0400>;: 如果pathname指终端设备，则不将此设备分配作为此进程的控制终端。not fcntl O_TRUNC<01000>;: 如果此文件存在，而且为只读或只写成功打开，则将其长度截短为0。not fcntl
O_APPEND<02000>;: 每次写时都加到文件的尾端
O_NONBLOCK<04000>;: 如果p a t h n a m e指的是一个F I F O、一个块特殊文件或一个字符特殊文件，则此选择项为此文件的本次打开操作和后续的I / O操作设置非阻塞方式。
O_NDELAY;;
O_SYNC<010000>;: 使每次write都等到物理I/O操作完成。
FASYNC<020000>;: 兼容BSD的fcntl同步操作
O_DIRECT<040000>;: 直接磁盘操作标识
O_LARGEFILE<0100000>;: 大文件标识
O_DIRECTORY<0200000>;: 必须是目录
O_NOFOLLOW<0400000>;: 不获取连接文件
O_NOATIME<01000000>;: 暂无
当新创建一个文件时，需要指定mode 参数，以下说明的格式如宏定义名称<实际常数值>;: 描述。
S_IRWXU<00700>;：文件拥有者有读写执行权限
S_IRUSR (S_IREAD)<00400>;：文件拥有者仅有读权限
S_IWUSR (S_IWRITE)<00200>;：文件拥有者仅有写权限
S_IXUSR (S_IEXEC)<00100>;：文件拥有者仅有执行权限
S_IRWXG<00070>;：组用户有读写执行权限
S_IRGRP<00040>;：组用户仅有读权限
S_IWGRP<00020>;：组用户仅有写权限
S_IXGRP<00010>;：组用户仅有执行权限
S_IRWXO<00007>;：其他用户有读写执行权限
S_IROTH<00004>;：其他用户仅有读权限
S_IWOTH<00002>;：其他用户仅有写权限
S_IXOTH<00001>;：其他用户仅有执行权限
*/
//const O_RDONLY = 0x0000  // open for reading only
//const O_WRONLY = 0x0001  // open for writing only
//const O_RDWR = 0x0002    // open for reading and writing
//const O_APPEND = 0x0008  // writes done at eof
//const O_CREAT = 0x0100   // create and open file
//const O_TRUNC = 0x0200   // open and truncate
//const O_EXCL = 0x0400    // open only if file doesn't already exist
//const O_ACCMODE = 0x0003 // open only if file doesn't already exist

const O_ACCMODE = 0x0003 // 读写文件操作时，用于取出flag的低2位。
const O_RDONLY = 0x0000  // 只读打开
const O_WRONLY = 0x0001  //  只写打开
const O_RDWR = 0x0002    //  读写打开
const O_CREAT = 0x0100   // 文件不存在则创建，需要mode_t，not fcntl
const O_EXCL = 0x0200    // 如果同时指定了O_CREAT，而文件已经存在，则出错， not fcntl
const O_NOCTTY = 0x0400  //  如果pathname指终端设备，则不将此设备分配作为此进程的控制终端。not fcntl O_TRUNC<01000>;: 如果此文件存在，而且为只读或只写成功打开，则将其长度截短为0。not fcntl
const O_APPEND = 0x02000 //  每次写时都加到文件的尾端
const O_TRUNC = 0x01000  // 如果此文件存在，而且为只读或只写成功打开，则将其长度截短为0

func (n *Node) test(ctx context.Context, flags uint32) {
	log.Warnf("==========================================")
	log.Warnf("test path:%s flags:%d ", n.path, flags)

	// client, err := n.hdfsClient.GetClient()
	//// "open";  "append";  "create";
	//	client.Create()
	if flags&O_ACCMODE == O_RDONLY {
		log.Warn("O_RDONLY")
	}

	if flags&O_ACCMODE == O_RDWR {
		log.Warn("O_RDWR")
	}
	if flags&O_ACCMODE == O_WRONLY {
		log.Warn("O_WRONLY   ")
	}

	if flags&O_EXCL > 0 {
		log.Warn("O_EXCL")
	}

	if flags&O_TRUNC > 0 {
		/* If we're opening for write or read/write, O_TRUNC means we should blow
		 * away the file which is there and create our own file.
		 * */
		log.Warn("O_TRUNC")
		//	return O_WRONLY;
	}
	if flags&O_APPEND > 0 {
		/* If we're opening for write or read/write, O_TRUNC means we should blow
		 * away the file which is there and create our own file.
		 * */
		log.Warn("O_APPEND")
		//	return O_WRONLY;
	}

	if flags&O_CREAT > 0 {
		/* If we're opening for write or read/write, O_TRUNC means we should blow
		 * away the file which is there and create our own file.
		 * */
		log.Warn("O_CREAT")
		//	return O_WRONLY;
	}
	log.Warnf("----------------------------------")
}
