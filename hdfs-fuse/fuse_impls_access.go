package hdfs_fuse

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	log "github.com/sirupsen/logrus"
	"syscall"
)

var _ = (fs.NodeAccesser)((*Node)(nil))

func (n *Node) Access(ctx context.Context, mask uint32) syscall.Errno {
	log.Infof("* Access path:%s,mask:%d", n.path, mask)
	return fs.OK
}
