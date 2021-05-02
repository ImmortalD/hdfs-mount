package krb5

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/credentials"
	"github.com/jcmturner/gokrb5/v8/keytab"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

type Krb5Client struct {
	// krb5.conf 文件
	Krb5Conf string
	// 账号 使用keytab或账号密码登录时使用
	UserName string
	// 密码 使用账号密码登录时使用,配置了keytab,优先使用keytab
	Password string
	// "TEST.GOKRB5"
	Realm string
	// keytab
	// keytab文件,优先使用文件
	KeytabFile string
	// 16进制的keytab字符串
	KeytabHex string
	// ccahe 登录,最后尝试使用ccahe登录
	CCache string

	// krb5.conf 对象
	config *config.Config
	client *client.Client
}

func (kc *Krb5Client) LoadKrb5Client() (*client.Client, error) {
	krb5Client, err := kc.newKrb5Client(kc.Krb5Conf)
	if err != nil {
		kc.client = krb5Client
		return nil, err
	}

	err = krb5Client.Login()
	if err != nil {
		log.Errorf("login kerberos err:%v", err)
	}
	return krb5Client, err
}

func (kc *Krb5Client) newKrb5Client(krb5conf string) (*client.Client, error) {
	var err error
	if kc.config, err = config.Load(krb5conf); err != nil {
		return nil, err
	}
	// c.config.ResolveRealm(c.Realm)
	// Set to lookup KDCs in DNS
	// c.config.LibDefaults.DNSLookupKDC = true
	// Blank out the KDCs to ensure they are not being used
	// c.config.Realms = []config.Realm{}

	// 配置了keytab,优先使用
	if kc.KeytabFile != "" || kc.KeytabHex != "" {
		log.Infof("use username and keytab file login kerberos,username:%s")
		if kc.UserName == "" {
			return nil, errors.New("UserName not set")
		}

		kt, err := kc.getKeytab()
		if err == nil {
			return client.NewWithKeytab(kc.UserName, kc.Realm, kt, kc.config), nil
		} else {
			return nil, err
		}
	}

	// 使用账号密码
	if kc.UserName != "" && kc.Password != "" {
		log.Infof("use username and password login kerberos,username:%s,password:*****", kc.UserName)
		return client.NewWithPassword(kc.UserName, kc.Realm, kc.Password, kc.config, func(settings *client.Settings) {
		}), nil
	}

	// 最后尝试使用ccahe
	if kc.CCache != "" {
		log.Infof("use CCache login kerberos,CCache file:%s", kc.CCache)
		ccache, err := credentials.LoadCCache(kc.CCache)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Couldn't load keytab freom [%s]", kc.CCache))
		}

		return client.NewFromCCache(ccache, kc.config)
	}

	return nil, errors.New(`please specify a way to login to krb5
1.keytab
2.UserName and Password
3.kerberos ccahe`)

}

// 获取keytab
func (kc *Krb5Client) getKeytab() (*keytab.Keytab, error) {
	var ktBytes []byte
	var err error
	// 优先使用keytab文件
	if kc.KeytabFile == "" {
		log.Info("KeytabFile is not set")
		if kc.KeytabHex == "" {
			log.Info("KeytabHex is not set")
			return nil, errors.New("ERROR KeytabFile and KeytabHex is not set")
		} else {
			if ktBytes, err = hex.DecodeString(kc.KeytabHex); err != nil {
				return nil, err
			}
		}
	} else {
		if ktBytes, err = readKeytab(kc.KeytabFile); err != nil {
			return nil, err
		}
	}
	// 创建keytab
	kt := keytab.New()
	if err := kt.Unmarshal(ktBytes); err != nil {
		return nil, err
	} else {
		return kt, nil
	}
}

// 读取文件
func readKeytab(keytabFile string) ([]byte, error) {
	log.Infof("keytab file:%s", keytabFile)
	ktFile, err := os.Open(keytabFile)
	if err != nil {
		return nil, err
	}
	defer ktFile.Close()

	return ioutil.ReadAll(ktFile)
}
