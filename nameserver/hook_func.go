package nameserver

import (
	"context"
	"github.com/brahma-adshonor/gohook"
	"log"
	"net"
	_ "unsafe"
)

// 替换掉原来的函数
//go:linkname lookupIPAddr net.(*Resolver).lookupIPAddr
func lookupIPAddr(r *net.Resolver, ctx context.Context, network, host string) ([]net.IPAddr, error)

//go:noinline
func OriginLookupIPAddr(r *net.Resolver, ctx context.Context, network, host string) ([]net.IPAddr, error) {
	log.Printf("这里的代码不会执行,只是为了给函数生成的代码多一些,以便有足够的空间")
	log.Printf("1")
	log.Printf("2")
	log.Printf("3")
	return nil, nil
}

func InstallHook(s NameServer) error {
	return gohook.HookByIndirectJmp(lookupIPAddr, s.Resolve, OriginLookupIPAddr)
}
