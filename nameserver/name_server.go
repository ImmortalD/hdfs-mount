package nameserver

import (
	"context"
	log "github.com/sirupsen/logrus"
	"net"
)

type NameServer interface {
	// 初始化
	Init() error
	// 获取类型
	GetType() string
	// 解析域名
	Resolve(r *net.Resolver, ctx context.Context, network, host string) ([]net.IPAddr, error)
}

// type ResolverFunc func(ctx context.Context, network string, host string) ([]net.IPAddr, error)

func InstallNameServer(server NameServer) error {
	if err := InstallHook(server); err != nil {
		log.Infof("NameServer安装失败,错误:%v", err)
		return err
	}
	log.Infof("NameServer安装成功")

	if err := server.Init(); err != nil {
		log.Errorf("NameServer init fail,err:%v", err)
		return err
	}

	return nil
}
