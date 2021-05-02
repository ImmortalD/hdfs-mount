package hdfs_fuse

import (
	"github.com/colinmarc/hdfs/v2"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"hdfs-mount/hdfsclient"
	"os/user"
	"path"
	"strconv"
	"sync"
	"syscall"
)

// Set file owners to the current user,
// otherwise in OSX, we will fail to start.

func init() {

}

// Node 文件系统中的树节点，它既充当目录又充当文件
type Node struct {
	fs.Inode
	// 是否是目录
	isDir bool
	// 当前文件的路径
	path string
	// 保护文件操作
	mutex sync.Mutex
	// hdfs操作客服端
	hdfsClient *hdfsclient.DefaultHdfsClient
	// 处理中的file handle
	// 其中的 writer *hdfs.FileWriter是下一个变量writer的引用
	// reader []*hdfs.FileReader每次创建一个
	fileHandle []*FileHandle
	// 写hdfs操作使用,hdfs每次只能打开一盒FileWriter,FileReader可以打开多个
	writer *hdfs.FileWriter
	// writer写入数据大小
	writeLen int64
	// 加密
	enc *Enc
}

const OpenAccessModeMask uint32 = syscall.O_ACCMODE

// Flags that can be seen in OpenRequest.Flags.
const (
	// Access modes. These are not 1-bit flags, but alternatives where
	// only one can be chosen. See the IsReadOnly etc convenience
	// methods.
	OpenReadOnly  uint32 = syscall.O_RDONLY
	OpenWriteOnly uint32 = syscall.O_WRONLY
	OpenReadWrite uint32 = syscall.O_RDWR

	// File was opened in append-only mode, all writes will go to end
	// of file. FreeBSD does not provide this information.
	OpenAppend    uint32 = syscall.O_APPEND
	OpenCreate    uint32 = syscall.O_CREAT
	OpenDirectory uint32 = syscall.O_DIRECTORY
	OpenExclusive uint32 = syscall.O_EXCL
	OpenNonblock  uint32 = syscall.O_NONBLOCK
	OpenSync      uint32 = syscall.O_SYNC
	OpenTruncate  uint32 = syscall.O_TRUNC
)

// Return true if OpenReadOnly is set.
func (n *Node) hasRead(flags uint32) bool {
	return flags&OpenAccessModeMask == OpenReadOnly || n.hasReadWrite(flags)
}

// Return true if OpenWriteOnly is set.
func (n *Node) hasWrite(flags uint32) bool {
	return flags&OpenAccessModeMask == OpenWriteOnly || n.hasReadWrite(flags)
}

func (n *Node) hasReadWrite(flags uint32) bool {
	return flags&OpenAccessModeMask == OpenReadWrite
}

// 生成读/写取hdfs文件的路径
func (n *Node) makeHdfsPath(parentPath, name string) string {
	return path.Join(n.hdfsClient.BaseDir, parentPath, name)
}

// 生产Linux系统中目录树的路径
func (n *Node) makeNodePath(parentPath, name string) string {
	return path.Join(parentPath, name)
}

// 获取文件inode
func (n *Node) getIno(stat interface{}) uint64 {
	s := stat.(*hdfs.FileStatus)
	return s.GetFileId()
}

// 获取Mode,是文件 or 目录
func (n *Node) getMode(isDir bool, mode uint32) uint32 {
	return getMode(isDir, mode)
}

func getMode(isDir bool, mode uint32) uint32 {
	if isDir {
		return uint32(syscall.S_IFDIR) | mode&07777
	} else {
		return uint32(syscall.S_IFREG) | mode&07777
	}
}

func atoi(id string, defaultId int) uint32 {
	nId, err := strconv.Atoi(id)
	if err != nil {
		return uint32(defaultId)
	} else {
		return uint32(nId)
	}

}

func hdfsFileStatToAttr(stat *hdfs.FileStatus, isDir bool, out *fuse.Attr) (uer, group string) {
	out.Ino = stat.GetFileId()
	out.Atime = stat.GetModificationTime() / 1000
	out.Mtime = stat.GetModificationTime() / 1000
	out.Ctime = stat.GetAccessTime() / 1000
	out.Atimensec = uint32(out.Atime % 1000)
	out.Mtimensec = uint32(out.Mtime % 1000)
	out.Ctimensec = uint32(out.Ctime % 1000)
	out.Mode = getMode(isDir, stat.GetPermission().GetPerm())
	out.Nlink = stat.GetBlockReplication()
	// Blksize is the preferred size for file system operations.
	out.Blksize = uint32(stat.GetBlocksize())
	// 是否是目录
	if isDir {
		out.Blocks = 0
		out.Size = 4096
	} else {
		out.Size = stat.GetLength()
		// out.Blocks = (stat.GetLength() + stat.GetBlocksize()) / (stat.GetBlocksize() + 1)
		// https://github.com/hanwen/go-fuse/pull/348
		// https://man7.org/linux/man-pages/man2/stat.2.html
		out.Blocks = (out.Size + 511) / 512
	}

	defOwner := fuse.CurrentOwner()
	u, err := user.Lookup(stat.GetOwner())
	if err != nil {
		out.Owner.Uid = defOwner.Uid
	} else {
		out.Owner.Uid = atoi(u.Uid, int(defOwner.Uid))
	}

	g, err := user.LookupGroup(stat.GetGroup())
	if err != nil {
		out.Owner.Gid = defOwner.Gid
	} else {
		out.Owner.Gid = atoi(g.Gid, int(defOwner.Gid))
	}

	return stat.GetOwner(), stat.GetGroup()
}
