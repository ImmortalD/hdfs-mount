package hdfs_fuse

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	log "github.com/sirupsen/logrus"
	"syscall"
)

var _ = (fs.NodeUnlinker)((*Node)(nil))

func (n *Node) Unlink(ctx context.Context, name string) syscall.Errno {
	log.Infof("* Unlink path:%s,name:%s", n.path, name)

	hdfsClient, err := n.hdfsClient.GetClient()
	if err != nil {
		return syscall.EIO
	}
	defer n.hdfsClient.ReleaseClient(&hdfsClient)

	err = hdfsClient.RemoveAll(n.makeHdfsPath(n.path, name))
	if err != nil {
		return syscall.EIO
	}

	n.RmChild(n.path)
	return fs.OK
}
