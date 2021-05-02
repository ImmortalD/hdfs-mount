# hdfs-mount 
把hdfs文件系统挂载到本地

- build on windows
```bat
set GOARCH=amd64
set GOOS=linux
go build -trimpath
```

- build on linux
```shell
go build -trimpath
```


配置`config.yaml`
```yaml
# hdfs 连接池
pool:
  # 最小的连接,初始化时使用
  minNum: 1
  # 最大连接,根据需要调整,目前空闲回收算法不是很好,连接数变大后很难回收
  maxNum: 2
  # 空闲是保留的连接数量
  idleOpenNum: 1
mount:
  # 挂载到的本地目录 
  mountPoint: "/mnt/hdfs"
  # 是否输出fuse调试信息
  debug: true
  # 是否允许其他用户读取,如果为false,例如使用Linux用户centos用户挂载,其他用户都不能访问,即使是root用户也没权限访问.
  allowOther: false
  # entry在内核缓存时间,单位:秒
  entryTimeout: 10
  # attr在内核缓存时间,单位:秒
  attrTimeout: 10
  # mount -o的选项
  options: [ "rw" ]
app:
  # 使用CPU core数
  maxNumCPU: 2
log:
  # 日志文件名,每天一个滚动日志 
  logName: "hdfs_mount.log"
  # 日志级别,tarce debug info warn error fatal 
  logLevel: "info"
  # 保存日志最大的数量
  maxRemainNum: 1000
# pprof 用户程序诊断,内存泄漏,死锁,cpu使用过高,内存使用过高等,开启后在浏览器输入 http://ip:prot/debug/pprof/查看
# 没有认证,建议生产不要开启,除非程序出现BUG开启诊断来分析
pprof:
  enable: false
  # http://ip:port/debug/pprof/
  address: "0.0.0.0"
  prot: 8080

hdfs:
  # 挂载后映射到hdfs的根目录
  mountDir: "/"
  #  namenode 地址
  nnAddress: [ "aa.c.om:8020" ,"bb.c.om:8020"]
  # hdfs 用户,如果配置了kerberos,这个用户和kerberos用户保持一致
  user: "root"
  # 是否使用kerberos认证
  kerberos:
    enable: false
    # krb5.conf 文件
    krb5Conf: "/root/krb5.conf"
    # 账号 使用keytab或账号密码登录时使用
    userName: "root"
    # 密码 使用账号密码登录时使用,配置了keytab,优先使用keytab
    # password: ""
    # "TEST.GOKRB5"
    realm: ""
    principleName: "hdfs/_HOST"
    # keytab
    # keytab文件,优先使用文件
    keytabFile: "/root/root.keytab"
    # 16进制的keytab字符串
    keytabHex:
    # ccahe 登录,最后尝试使用ccahe登录
    CCache: /tmp/u_ccahe
  # 配置数据写入时自动加密,目前读取不支持解密   
  encrypt:
    enable: false
    # 使用aes-256-cbc模式的密码没,加密和openssl enc -aes-256-cbc -md sha512 -in plaint.txt -out encrypt.txt -pass pass:1 一样
    # 解密  openssl enc -d -md sha512 -aes-256-cbc -in encrypt.txt -out plaint.txt -pass pass:1
    password: "1"
    # key的长度
    keyLen: 256

# 配置程序内置DNS解析
customerNameServer:
  enable: true
  ns:
    # 域名: ip
    aa.c.com: "1.2.3.4"
    bb.c.om: "2.3.4.5"
    cc.c.om: "2.3.4.5"
```

- 挂载
```shell
mkdir -p /mnt/hdfs
./hdfs-mount -config config.yaml &
```

