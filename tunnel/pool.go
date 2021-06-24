package tunnel

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Pool struct {
	conn    chan net.Conn
	server  Endpoint
	remote  Endpoint
	config  *ssh.ClientConfig
	request chan bool
	exit    chan bool
	once    sync.Once
	client  *ssh.Client
}

func NewPool(server, remote Endpoint, config *ssh.ClientConfig) *Pool {
	return &Pool{
		conn:    make(chan net.Conn, 1),
		server:  server,
		remote:  remote,
		config:  config,
		request: make(chan bool, 1),
		exit:    make(chan bool, 1),
	}
}

func (p *Pool) Start(ctx context.Context) error {
	serverConn, err := ssh.Dial("tcp", p.server.String(), p.config)
	if err != nil {
		return errors.Wrap(err, "Server dial error")
	}
	p.client = serverConn
	go p.keepAlive(p.client, ctx)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-p.request:
				p.fillWaitConn()
			case <-p.exit:
				return
			}
		}
	}()
	return nil
}

func (p *Pool) Stop() {
	p.exit <- true
	close(p.request)
	close(p.conn)
	for conn := range p.conn {
		_ = conn.Close()
	}
}

func (p *Pool) Get() net.Conn {
	go func() {
		p.request <- true
	}()
	conn := <-p.conn
	return conn
}

func (p *Pool) fillWaitConn() {
	now := time.Now()
	conn, err := p.retrieveConn()
	logrus.Infof("take %s to retrieve connection \n", time.Since(now))
	if err != nil {
		logrus.Error(err)
		return
	}
	p.conn <- conn
}

func (p *Pool) WarnUp(ctx context.Context) error {
	serverConn, err := ssh.Dial("tcp", p.server.String(), p.config)
	if err != nil {
		return errors.Wrap(err, "Server dial error")
	}
	p.client = serverConn
	go p.keepAlive(p.client, ctx)

	//now := time.Now()
	//conn, err := p.retrieveConn()
	//logrus.Infof("take %s to retrieve connection \n", time.Since(now))
	//if err != nil {
	//	return err
	//}
	//
	//p.conn <- conn

	return nil
}

func (p *Pool) retrieveConn() (net.Conn, error) {
	//serverConn, err := ssh.Dial("tcp", p.server.String(), p.config)
	//if err != nil {
	//	return nil, errors.Wrap(err, "Server dial error")
	//}

	remoteConn, err := p.client.Dial("tcp", p.remote.String())
	if err != nil {
		return nil, errors.Wrap(err, "Remote dial error")
	}
	return remoteConn, nil
}

func (p *Pool) keepAlive(client *ssh.Client, ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	var aliveCount int32
	var wg sync.WaitGroup
	for {
		select {
		case <-ticker.C:
			if n := atomic.AddInt32(&aliveCount, 1); n > 12 {
				err := client.Close()
				if err != nil {
					logrus.Error(err)
				}
				logrus.Infof("(%v) SSH keep alive termination", n)
				return
			}
		case <-ctx.Done():
			return
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := client.SendRequest("keepalive@openssh.com", true, nil)
			if err == nil {
				atomic.StoreInt32(&aliveCount, 0)
			}
		}()
	}
}
