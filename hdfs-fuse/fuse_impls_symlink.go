package hdfs_fuse

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	log "github.com/sirupsen/logrus"
	"syscall"
)

var _ = (fs.NodeSymlinker)((*Node)(nil))

func (n *Node) Symlink(ctx context.Context, target, name string, out *fuse.EntryOut) (node *fs.Inode, errno syscall.Errno) {
	log.Warnf("* Symlink path:%s,target:%s,name:%s set attr not support", n.path, target, name)

	return nil, syscall.ENOTSUP
}


