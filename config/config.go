package config

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"os"
)

type HdfsMountConf struct {
	Pool       PoolConf               `yaml:"pool"`
	Hdfs       HdfsConf               `yaml:"hdfs"`
	NameServer CustomerNameServerConf `yaml:"customerNameServer"`
	Mount      MountConf              `yaml:"mount"`
	Log        LogConf                `yaml:"log"`
	PProf      PProfConf              `yaml:"pprof"`
	App        AppConf                `yaml:"app"`
}
type AppConf struct {
	MaxNumCPU int
}
type PProfConf struct {
	Enable bool `yaml:"enable"`
	// 0.0.0.0
	Address string `yaml:"address"`
	// 6060
	Port int `yaml:"port"`
}
type LogConf struct {
	LogName      string `yaml:"logName"`
	LogLevel     string `yaml:"logLevel"`
	MaxRemainNum uint   `yaml:"maxRemainNum"`
}
type EncryptConf struct {
	Enable   bool   `yaml:"enable"`
	Password string `yaml:"password"`
	KeyLen   int    `yaml:"keyLen"`
}
type MountConf struct {
	MountPoint   string   `yaml:"mountPoint"`
	Debug        bool     `yaml:"debug"`
	AllowOther   bool     `yaml:"allowOther"`
	EntryTimeout int      `yaml:"entryTimeout"`
	AttrTimeout  int      `yaml:"attrTimeout"`
	Options      []string `yaml:"options"`
}

type HdfsConf struct {
	MountDir  string       `yaml:"mountDir"`
	NnAddress []string     `yaml:"nnAddress"`
	User      string       `yaml:"user"`
	Kerberos  kerberosConf `yaml:"kerberos"`
	Encrypt   EncryptConf  `yaml:"encrypt"`
}
type PoolConf struct {
	MinNum      int `yaml:"minNum"`
	MaxNum      int `yaml:"maxNum"`
	IdleOpenNum int `yaml:"idleOpenNum"`
}

type kerberosConf struct {
	Enable   bool   `yaml:"enable"`
	Krb5Conf string `yaml:"krb5Conf"`
	// 账号 使用keytab或账号密码登录时使用
	UserName string `yaml:"userName"`
	// 密码 使用账号密码登录时使用
	Password                     string `yaml:"password"`
	KerberosServicePrincipleName string `yaml:"principleName"`
	// "TEST.GOKRB5"
	Realm string `yaml:"realm"`
	// keytab
	// keytab文件,优先使用文件
	KeytabFile string `yaml:"keytabFile"`
	// 16进制的keytab字符串
	KeytabHex string `yaml:"keytabHex"`
	// ccahe 登录,最后尝试使用ccahe登录
	CCache string `yaml:"CCache"`
}

type CustomerNameServerConf struct {
	Enable bool              `yaml:"enable"`
	Ns     map[string]string `yaml:"ns"`
}

// LoadConfig 加载配置
func LoadConfig(config string) (*HdfsMountConf, error) {
	var conf HdfsMountConf
	file, err := os.Open(config)
	if err != nil {
		log.Errorf("open config err:%v", err)
		return nil, err
	}

	err = yaml.NewDecoder(file).Decode(&conf)
	if err != nil {
		log.Errorf("decode config err:%v", err)
		return nil, err
	}
	return &conf, nil
}
