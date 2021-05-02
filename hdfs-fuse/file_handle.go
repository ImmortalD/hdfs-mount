package hdfs_fuse

import (
	"github.com/colinmarc/hdfs/v2"
	log "github.com/sirupsen/logrus"
	"hdfs-mount/encrypt"
	"hdfs-mount/encrypt/openssl"
	"sync"
)

var seq = 0

func NewFileHandle(path *string, reader *hdfs.FileReader, writer *hdfs.FileWriter,
	isWrite bool, isRead bool) *FileHandle {
	seq++
	return &FileHandle{
		path:      path,
		readMutex: sync.Mutex{},
		reader:    reader,
		writer:    writer,
		isWrite:   isWrite,
		isRead:    isRead,
		seq:       seq,
	}
}

// FileHandle 处理文件的一些请求
type FileHandle struct {
	// 当前文件的路径,挂载到Linux中的路径,不是hdfs文件路径
	// 如果挂载到hdfs的/目录,那么和hdfs文件路径一样
	path *string
	// 保护读操作
	readMutex sync.Mutex
	// 保护写操作
	writeMutex sync.Mutex
	// 读取hdfs数据
	reader *hdfs.FileReader
	// 写入hdfs数据
	writer *hdfs.FileWriter
	// 是否可写,目前只是标记,不使用,可能标记不准确
	isWrite bool
	// 是否可读,目前只是标记,不使用,可能标记不准确
	isRead bool
	// 创建的顺序,目前不使用
	seq int
}

type Enc struct {
	enable    bool
	password  string
	keyLen    int
	encryptor encrypt.Encrypter
}

func NewEnc(password string, keyLen int, enable bool) (*Enc, error) {
	// 开启加密写入
	if enable {
		enc, err := openssl.NewAesEncrypterByPass(password, keyLen, openssl.BytesToKeySHA512, openssl.CBCEncrypter)
		if err != nil {
			log.Errorf("create openssl aes cbc encryptor error:%v", err)
			return nil, err
		}
		return &Enc{
			enable:    true,
			password:  password,
			keyLen:    keyLen,
			encryptor: enc,
		}, nil
	}

	return &Enc{
		enable:    false,
		password:  "",
		keyLen:    0,
		encryptor: nil,
	}, nil
}
