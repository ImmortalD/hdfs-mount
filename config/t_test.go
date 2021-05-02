package config

import (
	"gopkg.in/yaml.v2"
	"os"
	"testing"
)

func TestA(t *testing.T) {
	var conf HdfsMountConf
	file, err := os.Open("nameserver.yaml")
	if err != nil {
		t.Fatalf("open config err:%v\n", err)
	}

	err = yaml.NewDecoder(file).Decode(&conf)
	if err != nil {
		t.Fatalf("open config err:%v\n", err)
	}

	t.Logf("config:%v\n", err)

}
