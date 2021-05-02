package hdfsclient

import (
	"errors"
	"github.com/colinmarc/hdfs/v2"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
	"unsafe"
)

var (
	ErrInvalidConfig = errors.New("invalid pool config")
	ErrPoolClosed    = errors.New("pool closed")
	ErrPoolNoConnect = errors.New("can not get hdfs client from hdfs client pool")
)

type factory func() (*hdfs.Client, error)
type HdfsPool struct {
	sync.Mutex
	pool            chan *HdfsConnect
	minOpen         int  // 池中最少资源数
	maxOpen         int  // 池中最大资源数
	numOpen         int  // 当前池中资源数
	numIdleOpen     int  // 空闲时池中资源数
	currentIdleOpen int  // 池中空闲资源数
	closed          bool // 池是否已关闭
	closeCh         chan struct{}
	factory         factory
}

type HdfsConnect struct {
	*hdfs.Client
	lastAccessTime time.Time
}

func NewHdfsPool(minOpen, maxOpen, numIdleOpen int, f factory) (*HdfsPool, error) {
	if maxOpen <= 0 || minOpen > maxOpen {
		return nil, ErrInvalidConfig
	}
	p := &HdfsPool{
		maxOpen:     maxOpen,
		minOpen:     minOpen,
		numIdleOpen: numIdleOpen,
		pool:        make(chan *HdfsConnect, maxOpen),
		factory:     f,
	}

	for i := 0; i < minOpen; i++ {
		client, err := f()
		if err != nil {
			continue
		}
		renameClientName(client)
		p.numOpen++
		p.pool <- &HdfsConnect{Client: client, lastAccessTime: time.Now()}
	}

	go p.startCheck()
	return p, nil
}

func renameClientName(client *hdfs.Client) {
	/**
	type Client struct {
		namenode *rpc.NamenodeConnection
		...
	}

	type NamenodeConnection struct {
	ClientID   []byte
	ClientName string
	....
	}
	*/

	/**
	 NamenodeConnection = *(*int)(unsafe.Pointer(client))
	 *ClientName = uintptr(NamenodeConnection) + unsafe.Sizeof([]byte)
	千万不能出现这种用临时变量中转一下的情况。因为GC可能因为优化内存碎片的原因移动了这个对象。只保留了指针的地址是没有意义的。
	*/
	pClientID := (*[]byte)(unsafe.Pointer(uintptr(*(*int)(unsafe.Pointer(client)))))
	// *pClientID = []byte("sssssssssss")
	ClientID := *pClientID

	// "go-hdfs-" + string(clientId),
	var o []byte
	pClientName := (*string)(unsafe.Pointer(uintptr(*(*int)(unsafe.Pointer(client))) + unsafe.Sizeof(o)))
	*pClientName = "hadoop-hdfs-fs" + string(ClientID)
}

func (p *HdfsPool) removeIdleConnect() {
	if len(p.pool) <= p.numIdleOpen {
		return
	}

	for i := 0; i < len(p.pool); i++ {
		var client *HdfsConnect
		select {
		case client = <-p.pool:
		default:
			return
		}

		if time.Now().Sub(client.lastAccessTime).Seconds() > 60 {
			log.Infof("idle close,current open %d,currentIdleOpen:%d", p.numOpen, len(p.pool))
			p.Close(client.Client)
		} else {
			p.pool <- client
		}

		if len(p.pool) <= p.numIdleOpen {
			return
		}
	}

}
func (p *HdfsPool) startCheck() {
	tickTimer := time.NewTicker(10 * time.Second)
	i := 10
	for {
		select {
		case <-tickTimer.C:
			p.removeIdleConnect()
			i++
			if i%6 == 0 {
				log.Debugf("minOpen:%d maxOpen:%d numIdleOpen:%d currentIdleOpen:%d numOpen:%d poolsize:%d  ",
					p.minOpen, p.maxOpen, p.numIdleOpen, p.currentIdleOpen, p.numOpen, len(p.pool))
			}
		case <-p.closeCh:
			goto end
		}
	}

end:
	tickTimer.Stop()
}
func (p *HdfsPool) Acquire() (*HdfsConnect, error) {
	if p.closed {
		return nil, ErrPoolClosed
	}
	for {
		client, err := p.getOrCreate()
		if err != nil {
			return nil, err
		}
		// todo maxLifttime处理
		return client, nil
	}
}

func (p *HdfsPool) getOrCreate() (*HdfsConnect, error) {
	log.Debugf("get hdfs client,minOpen:%d maxOpen:%d numIdleOpen:%d currentIdleOpen:%d numOpen:%d poolsize:%d",
		p.minOpen, p.maxOpen, p.numIdleOpen, p.currentIdleOpen, p.numOpen, len(p.pool))

	p.Lock()
	tickTimer := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-tickTimer.C:
			p.Unlock()
			tickTimer.Stop()
			return nil, ErrPoolNoConnect
		default:
			if len(p.pool) > 0 {
				client := <-p.pool
				p.currentIdleOpen = len(p.pool)
				p.Unlock()
				return client, nil
			} else if p.numOpen < p.maxOpen {
				goto creat
			}
		}
	}
creat:
	defer p.Unlock()
	// 新建连接
	log.Infof("minOpen:%d maxOpen:%d numIdleOpen:%d currentIdleOpen:%d numOpen:%d poolsize:%d,will create hdfs client",
		p.minOpen, p.maxOpen, p.numIdleOpen, p.currentIdleOpen, p.numOpen, len(p.pool))
	client, err := p.factory()
	if err != nil {
		log.Errorf("create hdfs client error,err:%v", err)
		return nil, err
	}
	log.Infof("create hdfs client success")
	renameClientName(client)
	p.numOpen++
	return &HdfsConnect{Client: client, lastAccessTime: time.Now()}, nil
}

// 释放单个资源到连接池
func (p *HdfsPool) Release(client *HdfsConnect) error {
	log.Infof("release hdfs client")
	if p.closed {
		return ErrPoolClosed
	}

	client.lastAccessTime = time.Now()
	p.Lock()
	defer p.Unlock()
	p.pool <- client
	p.currentIdleOpen = len(p.pool)
	return nil
}

// 关闭单个资源
func (p *HdfsPool) Close(client *hdfs.Client) {
	log.Info("close hdfs client")
	defer func() {
		err := recover()
		if err != nil {
			log.Errorf("close hdfs client error,err:%v", err)
		}

		p.Lock()
		defer p.Unlock()
		p.numOpen--
		p.currentIdleOpen = len(p.pool)
	}()

	client.Close()
}

// 关闭连接池，释放所有资源
func (p *HdfsPool) Shutdown() error {
	log.Info("Shutdown hdfs client connect pool")
	if p.closed {
		return ErrPoolClosed
	}
	close(p.pool)
	for client := range p.pool {
		p.Close(client.Client)
	}
	p.closed = true
	return nil
}
