package hdfs_fuse

import (
	"context"
	"github.com/colinmarc/hdfs/v2"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	log "github.com/sirupsen/logrus"
	"io"
	"strconv"
	"syscall"
)

var _ = (fs.NodeReader)((*Node)(nil))

func (n *Node) OpenHdfsFileReader() (*hdfs.FileReader, syscall.Errno) {
	log.Infof("open hdfs file reader,file:%s", n.path)

	hdfsClient, err := n.hdfsClient.GetClient()
	if err != nil {
		return nil, fs.OK
	}
	defer n.hdfsClient.ReleaseClient(&hdfsClient)

	fullPath := n.makeHdfsPath(n.path, "")
	reader, err := hdfsClient.Open(fullPath)
	if err != nil {
		log.Errorf("open hdfs file reader error path:%s err:%v", n.path, err)
		return nil, syscall.EIO
	}
	return reader, fs.OK
}

// Read Hdfs File TODO
func (n *Node) Read(ctx context.Context, f fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	log.Debugf("* Read file:%s offset:%d", n.path, off)
	fh := f.(*FileHandle)
	var readLen int
	var err error
	{
		fh.readMutex.Lock()
		defer fh.readMutex.Unlock()
		// 打开reader
		if fh.reader == nil {
			reader, err := n.OpenHdfsFileReader()
			if err != fs.OK {
				return nil, syscall.EIO
			}
			fh.reader = reader
		}
		// 读取数据
		readLen, err = fh.reader.ReadAt(dest, off)
	}

	log.Debugf("Read file:%s offset:%d read length:%d", n.path, off, readLen)
	if err != nil && err != io.EOF {
		if err.Error() == "invalid resulting offset: "+strconv.Itoa(int(off)) {
			return fuse.ReadResultData(nil), fs.OK
		}
		return nil, fs.ToErrno(err)
	}
	return fuse.ReadResultData(dest[0:readLen]), fs.OK
}
