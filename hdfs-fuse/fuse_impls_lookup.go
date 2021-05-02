package hdfs_fuse

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	log "github.com/sirupsen/logrus"
	"syscall"
)

//var _ fs.NodeAccesser = &Node{}
var _ = (fs.NodeLookuper)((*Node)(nil))

// Lookup 在当前节点（目录）下查找文件
func (n *Node) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	log.Infof("* Lookup path:%s,name:%s", n.path, name)

	dir, err := n.hdfsClient.ReadDir(n.makeHdfsPath(n.path, ""))
	if err != nil {
		return nil, syscall.EIO
	}

	for _, file := range dir {
		if file.Name() == name {
			fullPath := n.makeNodePath(n.path, file.Name())
			enc, err := NewEnc(n.enc.password, n.enc.keyLen, n.enc.enable)
			if err != nil {
				log.Errorf("creat encryptor error. file %s,error:%v", fullPath, err)
				return nil, syscall.EIO
			}

			stat, isDir, err := n.hdfsClient.Stat(fullPath)
			if err != nil {
				return nil, syscall.EIO
			}

			// attr
			hdfsFileStatToAttr(stat, isDir, &out.Attr)

			child := NewNode(fullPath, file.IsDir(), n.hdfsClient, enc)
			mode := n.getMode(child.isDir, child.Mode())
			inode := n.NewInode(ctx, child, fs.StableAttr{
				Mode: mode,
				Ino:  n.getIno(file.Sys()),
				Gen:  child.Inode.StableAttr().Gen,
			})

			return inode, fs.OK
		}
	}

	return nil, syscall.ENOENT
}
