package nameserver

import (
	"testing"
	_ "unsafe"
)

func TestInstallNameServer(t *testing.T) {
	ns := &MemNameServer{
		MemConfig: nil,
	}

	if err := InstallNameServer(ns); err != nil {
		t.Fatalf("InstallNameServer InstallHook err: %v\n", err)
	}

}
