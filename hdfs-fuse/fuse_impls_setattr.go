package hdfs_fuse

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	log "github.com/sirupsen/logrus"
	"os"
	"os/user"
	"strconv"
	"syscall"
	"time"
)

var _ = (fs.NodeSetattrer)((*Node)(nil))


// Setattr Implement Setattr to support truncation
// return syscall.EOPNOTSUPP
// touch  echo > 用到
func (n *Node) Setattr(ctx context.Context, fh fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	log.Warnf("* Setattr path:%s,set attr not support", n.path)
	// 获取原来的属性
	stat, dir, err := n.hdfsClient.Stat(n.makeNodePath(n.path, ""))
	if err != nil {
		log.Errorf("Setattr call Stat error,path:%s,err:%v", n.path, err)
		return syscall.EIO
	}
	// 原来的User Group
	oriUser, oriGroup := hdfsFileStatToAttr(stat, dir, &out.Attr)

	hdfsClient, err := n.hdfsClient.GetClient()
	if err != nil {
		return syscall.EIO
	}
	defer n.hdfsClient.ReleaseClient(&hdfsClient)

	fullPath := n.makeHdfsPath(n.path, "")
	nodePath := n.makeNodePath(n.path, "")

	if in.Valid&fuse.FATTR_MODE != 0 {
		mode, _ := in.GetMode()
		if mode > 01777 {
			return  syscall.EINVAL
		}

		out.Mode = mode
		err = hdfsClient.Chmod(fullPath, os.FileMode(mode))
		if err != nil {
			log.Errorf("chmod error,path:%s mode:%v err:%v\n", nodePath, out.Mode, err)
			return syscall.EIO
		}
	} else if in.Valid&fuse.FATTR_ATIME != 0 || in.Valid&fuse.FATTR_MTIME != 0 {
		aTime := time.Unix(int64(out.Atime), int64(out.Atimensec))
		mTime := time.Unix(int64(out.Mtime), int64(out.Mtimensec))
		if in.Valid&fuse.FATTR_ATIME != 0 {
			aTime, _ = in.GetATime()
			out.Atime = uint64(aTime.Second())
			out.Atimensec = uint32(aTime.Nanosecond())
		}

		if in.Valid&fuse.FATTR_MTIME != 0 {
			mTime, _ = in.GetMTime()
			out.Mtime = uint64(mTime.Second())
			out.Mtimensec = uint32(mTime.Nanosecond())
		}
		err = hdfsClient.Chtimes(fullPath, aTime, mTime)
		if err != nil {
			log.Errorf("Chtimes error,path:%s ATime:%v MTime:%v\n", nodePath, aTime, mTime)
			return syscall.EIO
		}
	} else if in.Valid&fuse.FATTR_GID != 0 || in.Valid&fuse.FATTR_UID != 0 {
		if in.Valid&fuse.FATTR_UID != 0 {
			u, err := user.LookupId(strconv.Itoa(int(in.Owner.Uid)))
			if err != nil {
				return syscall.ENOENT
			} else {
				oriUser = u.Username
				out.Uid = atoi(u.Gid, int(out.Uid))
			}
		}
		if in.Valid&fuse.FATTR_GID != 0 {
			g, err := user.LookupGroupId(strconv.Itoa(int(in.Owner.Gid)))
			if err != nil {
				return syscall.ENOENT
			} else {
				oriGroup = g.Name
				out.Gid = atoi(g.Gid, int(out.Gid))
			}

			err = hdfsClient.Chown(fullPath, oriUser, oriGroup)
			if err != nil {
				log.Errorf("Chown error,path:%s User:%v Group:%v\n", nodePath, oriUser, oriGroup)
				return syscall.EIO
			}
		}

	}

	return fs.OK
}
