package hdfs_fuse

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	log "github.com/sirupsen/logrus"
	"syscall"
)

var _ = (fs.NodeFlusher)((*Node)(nil))

func (n *Node) Flush(ctx context.Context, fh fs.FileHandle) syscall.Errno {
	log.Debugf("* Flush path:%s", n.path)
	if n.writer != nil {
		err := n.writer.Flush()
		if err != nil {
			log.Errorf("flush file error,file:%s", n.path)
			return syscall.EIO
		}
	}
	return fs.OK
}
