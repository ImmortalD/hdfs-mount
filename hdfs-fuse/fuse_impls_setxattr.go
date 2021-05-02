package hdfs_fuse

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	log "github.com/sirupsen/logrus"
	"syscall"
)

var _ = (fs.NodeSetxattrer)((*Node)(nil))

func (n *Node) Setxattr(ctx context.Context, attr string, data []byte, flags uint32) syscall.Errno {
	log.Warnf("* Setxattr path:%s,attr:%s,data:%s,flags:%d,set xattr not support",
		n.path, attr, string(data), flags)

	// mv 用到
	return fs.OK
}
