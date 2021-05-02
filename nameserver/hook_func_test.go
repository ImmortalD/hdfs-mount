package nameserver

import (
	"testing"
	_ "unsafe"
)

func TestInstallHook(t *testing.T) {
	ns := &MemNameServer{
		MemConfig: nil,
	}

	if err := InstallHook(ns); err != nil {
		t.Fatalf("InstallHook err: %v\n", err)
	}

}
