package hdfs_fuse

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	log "github.com/sirupsen/logrus"
	"syscall"
)

var _ = (fs.NodeStatfser)((*Node)(nil))

func (n *Node) Statfs(ctx context.Context, out *fuse.StatfsOut) syscall.Errno {
	log.Warnf("* Statfs path:%s", n.path)
	hdfsClient, err := n.hdfsClient.GetClient()
	if err != nil {
		return syscall.EIO
	}
	defer n.hdfsClient.ReleaseClient(&hdfsClient)

	statFs, err := hdfsClient.StatFs()
	if err != nil {
		log.Errorf("Statfs error,path:%s,err:%v", n.path, err)
		return syscall.EIO
	}

	summary, err := hdfsClient.GetContentSummary(n.makeHdfsPath("/", ""))
	if err != nil {
		log.Errorf("Statfs GetContentSummary error,path:%s,err:%v", n.path, err)
		return syscall.EIO
	}

	//
	//long f_type; /* 文件系统类型 */
	//long f_bsize; /* 经过优化的传输块大小 */
	//long f_blocks; /* 文件系统数据块总数 */
	//long f_bfree; /* 可用块数 */
	//long f_bavail; /* 非超级用户可获取的块数 */
	//long f_files; /* 文件结点总数 */
	//long f_ffree; /* 可用文件结点数 */
	//fsid_t f_fsid; /* 文件系统标识 */
	//long f_namelen; /* 文件名的最大长度 */
	// http://www.hechaku.com/Unix_Linux/open.html
	// Inodes total
	out.Files = 1024 * 1024 * 1024 * 1024
	// Inodes Free
	out.Ffree = out.Files - uint64(summary.FileCount()+summary.DirectoryCount()) - 1

	out.NameLen = 8000
	var blocksize uint64 = 1 // 1024 * 1024 * 128

	out.Bsize = uint32(blocksize)
	out.Frsize = out.Bsize

	out.Blocks = (uint64(summary.Size()) + statFs.Remaining) / blocksize // statFs.Capacity / blocksize
	out.Bfree = statFs.Remaining / blocksize
	out.Bavail = out.Bfree
	return fs.OK
}
