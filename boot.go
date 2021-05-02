package main

import (
	"fmt"
	"github.com/colinmarc/hdfs/v2"
	log "github.com/sirupsen/logrus"
	"hdfs-mount/config"
	"hdfs-mount/hdfs-fuse"
	"hdfs-mount/hdfsclient"
	"hdfs-mount/krb5"
	"hdfs-mount/logger"
	"hdfs-mount/nameserver"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

//var conf *config.HdfsMountConf
var hdfsPool *hdfsclient.HdfsPool

func CreateHdfsClient(conf *config.HdfsMountConf) *hdfsclient.DefaultHdfsClient {
	var err error
	// 开启了kerberos认证
	kerberos := conf.Hdfs.Kerberos
	var krb5Client *krb5.Krb5Client
	if kerberos.Enable {
		log.Info("kerberos enable")
		krb5Client = &krb5.Krb5Client{
			Krb5Conf:   kerberos.Krb5Conf,
			UserName:   kerberos.UserName,
			Password:   kerberos.Password,
			Realm:      kerberos.Realm,
			KeytabFile: kerberos.KeytabFile,
			KeytabHex:  kerberos.KeytabHex,
			CCache:     kerberos.CCache,
		}
	} else {
		log.Info("kerberos disable")
	}

	// 开启了自定义的dns解析
	nServer := conf.NameServer
	if nServer.Enable {
		log.Info("name server enable")
		namedServer := nameserver.MemNameServer{
			MemConfig: nServer.Ns,
		}
		if err = namedServer.Init(); err != nil {
			log.Fatalf("load customer name server err:%v", err)
		}
		if err = nameserver.InstallNameServer(&namedServer); err != nil {
			os.Exit(3)
		}
	} else {
		log.Info("name server disable")
	}

	// 创建 hdfs Client
	c := hdfsclient.DefaultHdfsClient{
		BaseDir:                      conf.Hdfs.MountDir,
		NnAddress:                    conf.Hdfs.NnAddress,
		User:                         conf.Hdfs.User,
		EnableKerberos:               conf.Hdfs.Kerberos.Enable,
		KerberosServicePrincipleName: conf.Hdfs.Kerberos.KerberosServicePrincipleName,
		HdfsPool:                     nil,
		Krb5Client:                   krb5Client,
	}

	err = c.Init()
	if err != nil {
		log.Errorf("init hdfs client err:%v", err)
		os.Exit(4)
	}

	// hdfs连接池
	pool := conf.Pool
	hdfsPool, err = hdfsclient.NewHdfsPool(pool.MinNum, pool.MaxNum, pool.IdleOpenNum, func() (*hdfs.Client, error) {
		return hdfs.NewClient(*c.GetClientOptions())
	})
	if err != nil {
		log.Errorf("create hdfs pool err:%v", err)
		os.Exit(5)
	}
	c.HdfsPool = hdfsPool
	return &c
}

func InterruptProc() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for range signalChan {
			if err := umount(); err != nil {
				continue
			}

			if hdfsPool != nil {
				hdfsPool.Shutdown()
				hdfsPool = nil
			}
		}
	}()
}

var server *hdfs_fuse.Server
var mutex sync.Mutex

func umount() error {
	mutex.Lock()
	defer mutex.Unlock()
	if server != nil {
		if err := server.Unmount(); err != nil {
			log.Errorf("umount err:%v", err)
			return err
		} else {
			log.Infof("umount success")
		}
		server = nil
	}
	return nil
}

func start(conf *config.HdfsMountConf) {
	var err error
	// 初始化日志
	logConf := conf.Log
	logger.InitLog(&logConf.LogName, &logConf.LogLevel, logConf.MaxRemainNum)

	// pprof
	if conf.PProf.Enable {
		addressAndPort := fmt.Sprintf("%s:%d", conf.PProf.Address, conf.PProf.Port)
		webAddress := fmt.Sprintf("%s:%d", GetLoaclIP(), conf.PProf.Port)
		log.Infof("pprof is enable,listen address http://%s", addressAndPort)
		log.Infof("web access address http://%s", webAddress)

		go http.ListenAndServe(addressAndPort, nil)
	}
	// 初始化Hdfs client
	client := CreateHdfsClient(conf)
	encrypt := conf.Hdfs.Encrypt
	var enc *hdfs_fuse.Enc
	if encrypt.Enable {
		enc, err = hdfs_fuse.NewEnc(encrypt.Password, encrypt.KeyLen, encrypt.Enable)
		if err != nil {
			log.Errorf("creat encryptor error:%v", err)
			os.Exit(10)
		}
	} else {
		enc = &hdfs_fuse.Enc{}
	}
	// 挂载卷
	mount := conf.Mount
	server, err = hdfs_fuse.NewServer(&mount, client, enc)

	if err != nil {
		log.Errorf("mount error:%v", err)
		os.Exit(15)
	}

	InterruptProc()
	server.Wait()
	umount()
}

func GetLoaclIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
