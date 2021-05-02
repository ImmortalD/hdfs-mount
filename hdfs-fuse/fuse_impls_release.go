package hdfs_fuse

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	log "github.com/sirupsen/logrus"
	"syscall"
)

var _ = (fs.NodeReleaser)((*Node)(nil))

func (n *Node) Release(ctx context.Context, f fs.FileHandle) syscall.Errno {
	log.Infof("* Release path:%s", n.path)
	var fh *FileHandle
	var ok bool
	fh, ok = f.(*FileHandle)
	if !ok {
		return fs.OK
	}

	n.mutex.Lock()
	defer n.mutex.Unlock()

	if n.fileHandle == nil || len(n.fileHandle) == 0 {
		return fs.OK
	}

	n.RemoveHandle(f)
	// 关闭Reader 并且用处FileHandler
	{
		fh.readMutex.Lock()
		defer fh.readMutex.Unlock()
		if fh.reader != nil {
			err := fh.reader.Close()
			if err != nil {
				log.Errorf("close hdfs reader error,path:%s err:%v", n.path, err)
			}
		}
	}

	// 关闭Writer
	fh.writeMutex.Lock()
	defer fh.writeMutex.Unlock()

	if n.fileHandle == nil || len(n.fileHandle) == 0 && n.writer != nil {
		log.Infof("close hdfs writer path:%s", n.path)
		var err error
		var errWrite error
		if n.enc.enable && n.writeLen > 0 {
			encBuf := n.enc.encryptor.EncryptLast()
			_, errWrite = n.writer.Write(encBuf)
			if errWrite != nil {
				log.Errorf("write hdfs file last data error,path:%s err:%v", n.path, errWrite)
			}
		}

		err = n.writer.Close()
		if err != nil {
			log.Errorf("close hdfs writer error path:%s err:%v", n.path, errWrite)
		}
		if err != nil || errWrite != nil {
			return syscall.EIO
		}

		n.writer = nil
	}

	return fs.OK
}

func (n *Node) RemoveHandle(f fs.FileHandle) {
	fh := f.(*FileHandle)
	for i, h := range n.fileHandle {
		if h == fh {
			n.fileHandle = append(n.fileHandle[:i], n.fileHandle[i+1:]...)
			break
		}
	}
}
