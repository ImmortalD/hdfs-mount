package nameserver

import (
	"context"
	log "github.com/sirupsen/logrus"
	"net"
)

type MemNameServer struct {
	// 域名解析文件
	// Config string
	// 内存中保存的域名ip对应关系  host --> ip
	MemConfig map[string]string
}
type domain struct {
	Ns map[string]string
	//Host string `yaml:"host"`
	//Ip   string `yaml:"ip"`
}

func (fr *MemNameServer) Init() error {
	return nil
}

func (fr *MemNameServer) GetType() string {
	return "file"
}

//go:noinline
func (fr *MemNameServer) Resolve(r *net.Resolver, ctx context.Context, network, host string) ([]net.IPAddr, error) {
	log.Infof("NameServer domain resolve network: %s, host: %s", network, host)

	ip := fr.MemConfig[host]
	if ip == "" {
		log.Infof("can not resolver domain %s,will use system domain resolver", host)
		return OriginLookupIPAddr(r, ctx, network, host)
	}

	log.Infof("use file domain resolver success: [%s] --> [%s]", host, ip)
	return []net.IPAddr{{
		IP:   net.ParseIP(ip),
		Zone: "",
	}}, nil

}
