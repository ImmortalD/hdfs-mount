package hdfsclient

import (
	"errors"
	"fmt"
	"github.com/colinmarc/hdfs/v2"
	"github.com/colinmarc/hdfs/v2/hadoopconf"
	log "github.com/sirupsen/logrus"
	"hdfs-mount/krb5"
	"os"
	"path"
	"sync"
)

type HdfsClient interface {
	// 初始化
	Init() (*hdfs.Client, error)
}

type DefaultHdfsClient struct {
	BaseDir string
	// namenode 地址
	NnAddress []string
	// 用户指定客户端将充当哪个HDFS用户。
	// 除非启用了kerberos身份验证，否则它是必需的，
	// 在这种情况下，将从提供的凭据中确定是否为空。
	User string
	// 是否使用kerberos认证
	EnableKerberos bool
	// 使用kerberos认证后的Principle名称
	// 在core-site.xml的dfs.namenode.kerberos.principal配置的内容
	//  (<SERVICE>/<FQDN>) (例如: 'nn/_HOST')
	KerberosServicePrincipleName string
	Krb5Client                   *krb5.Krb5Client
	///////////////////////////////
	// hdfs pool
	HdfsPool *HdfsPool
	ops      *hdfs.ClientOptions
	mutex    sync.Mutex
}

// 获取Hadoop配置
func (c *DefaultHdfsClient) getHadoopConfig() (*hdfs.ClientOptions, error) {
	if c.NnAddress == nil {
		log.Infof("没有指定hdfs相关配置,将从来hadoop环境变量来获取配置,获取的顺序是HADOOP_CONF_DIR,HADOOP_HOME")
		conf, err := hadoopconf.LoadFromEnvironment()
		if err != nil || conf == nil {
			return nil, errors.New(fmt.Sprintf("Couldn't load ambient config  %v", err))
		}

		ops := hdfs.ClientOptionsFromConf(conf)
		if ops.Addresses == nil {
			return nil, errors.New(fmt.Sprintf("Missing namenode addresses in ambient config"))
		} else {
			return &ops, nil
		}

	}

	return &hdfs.ClientOptions{
		Addresses:                    c.NnAddress,
		User:                         c.User,
		UseDatanodeHostname:          false,
		KerberosClient:               nil,
		KerberosServicePrincipleName: c.KerberosServicePrincipleName,
	}, nil
}

func (c *DefaultHdfsClient) Init() error {
	var err error
	c.ops, err = c.getHadoopConfig()
	if err != nil {
		return errors.New(fmt.Sprintf("获取hadoop配置出错,error:%v", err))
	}

	if c.EnableKerberos {
		log.Info("kerberos is enable")
		client, err := c.Krb5Client.LoadKrb5Client()
		if err != nil {
			return err
		}
		c.ops.KerberosClient = client
		c.ops.User = client.Credentials.UserName()
		c.ops.User = ""
	} else {
		log.Info("kerberos is disable")
	}

	log.Infof("hadoop配置:%v ", c.ops)
	return err
}

func (c *DefaultHdfsClient) GetClientOptions() *hdfs.ClientOptions {
	return c.ops
}

func (c *DefaultHdfsClient) Stat(name string) (st *hdfs.FileStatus, isDir bool, err error) {
	defer func() {
		err := recover()
		if err != nil {
			log.Errorf("hdfs stat file:%s,painc:%v", name, err)
		}
	}()

	client, err := c.GetClient()
	if err != nil {
		return nil, false, err
	}
	defer c.ReleaseClient(&client)

	stat, err := client.Stat(path.Join(c.BaseDir, name))
	if err != nil {
		log.Errorf("hdfs stat file:%s,error:%v", name, err)
		return nil, false, err
	}

	return stat.Sys().(*hdfs.FileStatus), stat.IsDir(), nil
}

func (c *DefaultHdfsClient) ReadDir(dirname string) ([]os.FileInfo, error) {
	client, err := c.GetClient()
	if err != nil {
		return nil, err
	}
	defer c.ReleaseClient(&client)
	return client.ReadDir(dirname)
}

func (c *DefaultHdfsClient) GetClient() (*HdfsConnect, error) {
	client, err := c.HdfsPool.Acquire()
	if err != nil {
		log.Errorf("get hdfs connection error:%v", err)
	}
	return client, err
}

func (c *DefaultHdfsClient) ReleaseClient(client **HdfsConnect) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if client == nil || *client == nil {
		return
	}

	if err := c.HdfsPool.Release(*client); err != nil {
		log.Errorf("release hdfs connection error:%v", err)
	}
	*client = nil
}
