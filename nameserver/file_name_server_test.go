package nameserver

import (
	"context"
	"net"
	"testing"
)

func TestFileNameServer(t *testing.T) {
	var ns = MemNameServer{
		MemConfig: map[string]string{
			"a012p.com": "2.2.2.2",
		},
	}

	if err := ns.Init(); err != nil {
		t.Fatalf("Init err:%v\n", err)
	}

	paresDomain(t, ns.Resolve)

	if err := InstallNameServer(&ns); err != nil {
		t.Fatalf("InstallNameServer InstallHook err: %v\n", err)
	}

	paresDomain(t, func(r *net.Resolver, ctx context.Context, network, host string) ([]net.IPAddr, error) {
		return r.LookupIPAddr(ctx, host)
	})

}

type TestHookLookupIP func(r *net.Resolver, ctx context.Context, network, host string) ([]net.IPAddr, error)

func paresDomain(t *testing.T, f TestHookLookupIP) {
	resolver := net.Resolver{}
	ctx := context.TODO()

	addr, err := f(&resolver, ctx, "tcp", "a012p.com")

	if err != nil {
		t.Fatalf("Init err:%v\n", err)
	}

	if addr[0].String() != "2.2.2.2" {
		t.Fatalf("nameserver parse domain a012p.com fail\n")
	}
}
