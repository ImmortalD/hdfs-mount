package hdfs_fuse

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	log "github.com/sirupsen/logrus"
	"syscall"
)

var _ = (fs.NodeReaddirer)((*Node)(nil))

// 读取hdfs目录信息
func (n *Node) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	log.Infof("* Readdir path:%s", n.path)

	dir, err := n.hdfsClient.ReadDir(n.makeHdfsPath(n.path, ""))
	if err != nil {
		log.Errorf("Readdir error,path:%s,err:%v", n.path, err)
		return nil, syscall.EIO
	}

	entries := make([]fuse.DirEntry, 0, len(dir))
	entries = append(entries, fuse.DirEntry{Mode: n.getMode(true, 0), Name: "."})
	entries = append(entries, fuse.DirEntry{Mode: n.getMode(true, 0), Name: ".."})

	for _, file := range dir {
		d := fuse.DirEntry{
			Mode: n.getMode(file.IsDir(), uint32(file.Mode())),
			Name: file.Name(),
			Ino:  n.getIno(file.Sys()),
		}
		entries = append(entries, d)
	}

	return fs.NewListDirStream(entries), fs.OK
}
