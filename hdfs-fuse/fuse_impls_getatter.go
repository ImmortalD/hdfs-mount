package hdfs_fuse

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	log "github.com/sirupsen/logrus"
	"syscall"
)

var _ = (fs.NodeGetattrer)((*Node)(nil))

// Getattr 输出文件属性
func (n *Node) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	log.Infof("* Getattr path:%s", n.path)

	stat, dir, err := n.hdfsClient.Stat(n.makeNodePath(n.path, ""))
	if err != nil {
		log.Errorf("Getattr error,path:%s,err:%v", n.path, err)
		return syscall.EIO
	}

	hdfsFileStatToAttr(stat, dir, &out.Attr)
	return fs.OK
}
