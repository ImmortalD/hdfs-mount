package hdfs_fuse

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	log "github.com/sirupsen/logrus"
	"syscall"
)

var _ = (fs.NodeCreater)((*Node)(nil))

// Create actually writes an empty value into etcd (as a placeholder)
func (n *Node) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (
	node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	log.Infof("* Create path:%s,name:%s,flags:%d,mode:%d", n.path, name, flags, mode)
	fullPath := n.makeHdfsPath(n.path, name)

	hdfsClient, err := n.hdfsClient.GetClient()
	if err != nil {
		return nil, nil, flags, syscall.EIO
	}
	defer func() {
		n.hdfsClient.ReleaseClient(&hdfsClient)
	}()

	e := hdfsClient.CreateEmptyFile(fullPath)
	n.hdfsClient.ReleaseClient(&hdfsClient)
	if e != nil {
		log.Errorf("creat hdfs file %s,error:%v", n.path, e)
		return nil, nil, 0, syscall.EIO
	}
	enc, err := NewEnc(n.enc.password, n.enc.keyLen, n.enc.enable)
	if err != nil {
		log.Errorf("creat encryptor error,path:%s,file:%s,error:%v", n.path, name, err)
		return nil, nil, 0, syscall.EIO
	}

	child := NewNode(n.makeNodePath(n.path, name), false, n.hdfsClient, enc)
	ch := n.NewInode(ctx, child, fs.StableAttr{Mode: child.getMode(child.isDir, child.Mode())})

	if ch != nil {
		fh, fuseFlags, err := child.Open(ctx, flags)
		return ch, fh, fuseFlags, err
	}

	// 1. 直接创建FileHandle
	nodePath := n.makeNodePath(n.path, name)
	fileHandle := NewFileHandle(&nodePath, nil, nil, n.hasWrite(flags), n.hasRead(flags))
	child.fileHandle = append(child.fileHandle, fileHandle)

	// 2.使用Open创建FileHandle
	/*fileHandle, fuseFlags, errno = child.Open(ctx, flags)
	if errno != fs.OK {
		log.Printf("Open hdfs file %s,error:%v", fullPath, e)
		return nil, nil, flags, syscall.EIO
	}
	ch := n.NewInode(ctx, child, fs.StableAttr{Mode: child.getMode(child.isDir)})
	return ch, fh, fuseFlags, fs.OK*/

	return ch, fileHandle, fuse.FOPEN_DIRECT_IO, fs.OK
}
