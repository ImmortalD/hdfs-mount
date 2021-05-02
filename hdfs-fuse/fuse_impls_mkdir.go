package hdfs_fuse

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	log "github.com/sirupsen/logrus"
	"os"
	"syscall"
)

var _ = (fs.NodeMkdirer)((*Node)(nil))

func (n *Node) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	log.Infof("* Mkdir dir path:%s,name:%s,mode:%d", n.path, name, mode)

	if n.GetChild(name) != nil {
		return nil, syscall.EEXIST
	}
	fullPath := n.makeHdfsPath(n.path, name)

	hdfsClient, err := n.hdfsClient.GetClient()
	if err != nil {
		return nil, syscall.EIO
	}
	defer func() {
		n.hdfsClient.ReleaseClient(&hdfsClient)
	}()

	err = hdfsClient.Mkdir(fullPath, os.FileMode(mode))
	n.hdfsClient.ReleaseClient(&hdfsClient)
	if err != nil {
		return nil, syscall.EIO
	}

	enc, err := NewEnc(n.enc.password, n.enc.keyLen, n.enc.enable)
	if err != nil {
		log.Errorf("creat encryptor error,path:%s,new dir:%s,error:%v", n.path, name, err)
		return nil, syscall.EIO
	}
	node := NewNode(n.makeNodePath(n.path, name), true, n.hdfsClient, enc)

	newInode := n.NewInode(ctx, node,
		fs.StableAttr{
			Mode: n.getMode(true, 0),
		},
	)

	n.AddChild(name, newInode, true)

	var attr fuse.AttrOut
	e := node.Getattr(ctx, nil, &attr)
	if e != fs.OK {
		return nil, e
	}

	out.Attr = attr.Attr
	return newInode, fs.OK
}
