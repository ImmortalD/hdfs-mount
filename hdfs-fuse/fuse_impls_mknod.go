package hdfs_fuse

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	log "github.com/sirupsen/logrus"
	"syscall"
)

//var _ fs.NodeAccesser = &Node{}
var _ = (fs.NodeMknoder)((*Node)(nil))

func (n *Node) Mknod(ctx context.Context, name string, mode uint32, dev uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	log.Infof("* Mknod dir path:%s,name:%s,mode:%d ,dev:%d", n.path, name, mode, dev)

	return nil, fs.OK
}
