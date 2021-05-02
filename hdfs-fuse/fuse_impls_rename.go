package hdfs_fuse

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	log "github.com/sirupsen/logrus"
	"path"
	"syscall"
)

var _ = (fs.NodeRenamer)((*Node)(nil))

func (n *Node) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
	log.Infof("* Rename path:%s,name:%s,new name:%s,flags:%d", n.path, name, newName, flags)

	node, okNewParent := newParent.(*Node)
	child := n.GetChild(name)
	oldNode, okOldNode := child.Operations().(*Node)

	if okNewParent && okOldNode {
		oldPath := path.Join(n.path, name)
		newPath := path.Join(node.path, newName)
		log.Infof("Rename path:%s,new path:%s", oldPath, newPath)

		hdfsClient, err := n.hdfsClient.GetClient()
		if err != nil {
			return syscall.EIO
		}
		defer n.hdfsClient.ReleaseClient(&hdfsClient)

		err = hdfsClient.Rename(oldPath, newPath)
		if err != nil {
			log.Errorf("Rename path:%s,new path:%s error,err:%v", oldPath, newPath, err)
			return syscall.EIO
		}

		oldNode.path = newPath
	} else {
		return syscall.EIO
	}

	return fs.OK
}
