go 1.16

module hdfs-mount

replace hdfs-mount => ../hdfs-mount

require (
	github.com/brahma-adshonor/gohook v1.1.9
	github.com/colinmarc/hdfs/v2 v2.2.0
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/hanwen/go-fuse/v2 v2.1.0
	github.com/jcmturner/gokrb5/v8 v8.4.1
	github.com/kr/pretty v0.1.0 // indirect
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible // indirect
	github.com/lestrrat-go/strftime v1.0.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5 // indirect
	github.com/sirupsen/logrus v1.8.1
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.4.0
)
