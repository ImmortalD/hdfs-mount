package hdfs_fuse

import (
	"context"
	"github.com/colinmarc/hdfs/v2"
	"github.com/hanwen/go-fuse/v2/fs"
	log "github.com/sirupsen/logrus"
	"syscall"
)

var _ = (fs.NodeWriter)((*Node)(nil))

// Write saves to the internal "content" buffer
func (n *Node) Write(ctx context.Context, f fs.FileHandle, buf []byte, off int64) (written uint32, e syscall.Errno) {
	log.Debugf("* Write file:%s offset:%d write len:%d", n.path, off, len(buf))

	fh := f.(*FileHandle)
	fh.writeMutex.Lock()
	defer fh.writeMutex.Unlock()
	if fh.writer == nil {
		writer, err := n.OpenHdfsFileWriter()
		if err != fs.OK {
			return 0, err
		}
		fh.writer = writer
	}

	// 由于写的FileWriter始终是一个,要上锁
	n.mutex.Lock()
	defer n.mutex.Unlock()

	var writeLen int
	var err error

	if n.enc.enable {
		encBuf := n.enc.encryptor.Encrypt(buf)
		writeLen, err = n.writer.Write(encBuf)
	} else {
		writeLen, err = n.writer.Write(buf)
	}

	if err != nil {
		log.Errorf("wirite hdfs file %s,error:%v writeLen:%d", n.path, err, writeLen)
		return 0, syscall.EIO
	}

	log.Debugf("wirite hdfs file success,file:%s,real write size:%d write szie:%d", n.path, writeLen, len(buf))
	n.writeLen += int64(len(buf))
	return uint32(len(buf)), fs.OK
}

func (n *Node) OpenHdfsFileWriter() (*hdfs.FileWriter, syscall.Errno) {
	log.Infof("open hdfs file writer,file:%s", n.path)
	n.mutex.Lock()
	defer n.mutex.Unlock()

	// 文件还在写,直接返回
	if n.writer != nil {
		return n.writer, fs.OK
	}

	fullPath := n.makeHdfsPath(n.path, "")

	var fileLength uint64
	if n.enc.enable {
		stat, _, err := n.hdfsClient.Stat(n.makeNodePath(n.path, ""))
		if err != nil {
			log.Errorf("get hdfs file size error,file:%s err:%v", n.path, err)
			return nil, syscall.EIO
		}
		fileLength = stat.GetLength()
	}

	hdfsClient, err := n.hdfsClient.GetClient()
	if err != nil {
		return nil, syscall.EIO
	}
	defer n.hdfsClient.ReleaseClient(&hdfsClient)

	// 1. 空文件: 可以打开写句柄
	// 2. 非空文件: 没加密开Append
	if fileLength == 0 || (fileLength > 0 && !n.enc.enable) {
		writer, err := hdfsClient.Append(fullPath)
		if err != nil {
			log.Errorf("open hdfs file write error,path:%s err:%v", n.path, err)
			return nil, syscall.EIO
		}
		n.writer = writer
		return writer, fs.OK
	} else {
		log.Errorf("open hdfs file write error,because file size more then zero,path:%s file size:%d,encrypt:%v", n.path, fileLength, n.enc.enable)
		return nil, syscall.ENOTSUP
	}
}
