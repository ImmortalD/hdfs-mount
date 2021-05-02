package hdfs_fuse

import (
	"context"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	log "github.com/sirupsen/logrus"
	"hdfs-mount/config"
	"hdfs-mount/hdfsclient"
	"strings"
	"sync"
	"time"
)

type Server struct {
	*fuse.Server
	MountPoint   string
	Debug        bool
	EntryTimeout int
	AttrTimeout  int
	Options      []string
}

func NewNode(path string, isDir bool, hdfsClient *hdfsclient.DefaultHdfsClient, enc *Enc) *Node {
	return &Node{
		Inode:      fs.Inode{},
		isDir:      isDir,
		enc:        enc,
		path:       path,
		mutex:      sync.Mutex{},
		hdfsClient: hdfsClient,
	}
}

func NewServer(conf *config.MountConf, hdfsClient *hdfsclient.DefaultHdfsClient, enc *Enc) (*Server, error) {
	opts := &fs.Options{}
	entryTimeoutSec := time.Second * time.Duration(conf.EntryTimeout)
	opts.EntryTimeout = &entryTimeoutSec
	attrTimeoutSec := time.Second * time.Duration(conf.AttrTimeout)
	opts.AttrTimeout = &attrTimeoutSec
	opts.Debug = conf.Debug
	opts.AllowOther = conf.AllowOther
	opts.Options = conf.Options
	opts.DisableXAttrs = true
	opts.NullPermissions = true
	opts.Name = "hdfs-mount"
	// FsName
	opts.FsName = "hdfs-mount#"
	nnAddress := "dfs://"
	if hdfsClient.NnAddress != nil {
		for _, address := range hdfsClient.NnAddress {
			nnAddress = nnAddress + address + ";"
		}
	}
	nnAddress = strings.TrimSuffix(nnAddress, ";")
	opts.FsName = opts.FsName + nnAddress + hdfsClient.BaseDir

	opts.ExplicitDataCacheControl = false
	opts.OnAdd = func(ctx context.Context) {
		log.Infof("hdfs mount start ...")
	}

	log.Infof("hdfs dir[%s] mount [%s]", nnAddress, conf.MountPoint)

	root := NewNode("/", true, hdfsClient, enc)

	server, err := fs.Mount(conf.MountPoint, root, opts)
	if err != nil {
		log.Fatalf("Mount error: %v", err)
		return nil, err
	}

	return &Server{
		Server:       server,
		MountPoint:   conf.MountPoint,
		Debug:        conf.Debug,
		EntryTimeout: conf.EntryTimeout,
		AttrTimeout:  conf.AttrTimeout,
		Options:      conf.Options,
	}, nil

}
