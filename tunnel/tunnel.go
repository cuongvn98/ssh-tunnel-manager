package tunnel

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"io"
	"net"
	"os"
	"time"
)

//type Tunnels struct {
//	tunnels map[string]*SSHtunnel
//}
//
//func NewTunnels(tunnels ...*SSHtunnel) *Tunnels {
//	m := make(map[string]*SSHtunnel, len(tunnels))
//	for _, tunnel := range tunnels {
//		m[tunnel.Name] = tunnel
//	}
//	return &Tunnels{tunnels: m}
//}
//
//func (t *Tunnels) Start(ctx context.Context, name string) {
//	if tunnel, ok := t.tunnels[name]; ok && !tunnel.Running {
//		ctx := context.WithValue(ctx, "name", name)
//		if err := tunnel.Start(ctx); err != nil {
//			logrus.Error(err)
//		}
//	}
//}
//
//func (t *Tunnels) Stop(name string) {
//	if tunnel, ok := t.tunnels[name]; ok {
//		tunnel.Stop()
//	}
//}

//func (t *Tunnels) Start(ctx context.Context) error {
//	errs := make(chan error)
//	for _, tunnel := range t.tunnels {
//		go func(tunnel *SSHtunnel) {
//			errs <- tunnel.Start(ctx)
//		}(tunnel)
//	}
//	return <-errs
//}

//func (t *Tunnels) start(ctx context.Context, tunnel *SSHtunnel) error {
//	times := 0
//	for {
//		if times > 5 {
//			logrus.Infof("can't restart connection to: %s", tunnel.Local)
//			return fmt.Errorf("exceeded try times")
//		} else if times >= 1 {
//			logrus.Infof("wait to restart connection to: %s", tunnel.Local)
//			time.Sleep(2 * time.Second)
//			logrus.Infof("try to restart connection to: %s", tunnel.Local)
//		}
//		if err := tunnel.Start(ctx); err != nil {
//			logrus.Errorf("connection to %s occurred error: %s", tunnel.Local, err)
//			times++
//
//			continue
//		}
//		logrus.Infof("connection to %s already graceful shutdown", tunnel.Local)
//		return nil
//	}
//}

const (
	StartingStatus = iota
	RunningStatus
	StopStatus
)

type Endpoint struct {
	Host string
	Port int
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

type SSHtunnel struct {
	Name   string
	Status int

	exit chan bool

	Local  Endpoint
	Server Endpoint
	Remote Endpoint

	config *ssh.ClientConfig
	pool   *Pool
	Hook
}

type Status struct {
	Name    string `json:"name"`
	Running bool   `json:"running"`
}

func NewSSHTunnel(user string, local, remote, server Endpoint) *SSHtunnel {
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			sshAgent(),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 5,
	}

	pool := NewPool(server, remote, sshConfig)

	tunnel := &SSHtunnel{
		config: sshConfig,
		Local:  local,
		Server: server,
		Remote: remote,
		exit:   make(chan bool),
		pool:   pool,
		Hook:   Hook{},
		Status: StopStatus,
	}
	return tunnel
}

func (tunnel *SSHtunnel) Start(ctx context.Context) error {
	tunnel.Status = StartingStatus
	tunnel.fire(tunnel.Status)
	listener, err := net.Listen("tcp", tunnel.Local.String())
	if err != nil {
		return err
	}
	logrus.Infof("Tunnel to %s is establited", tunnel.Local.String())

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			logrus.Error("failed while close listener :", err)
		}
	}(listener)

	defer func() {
		logrus.Infof("Tunnel to %s is stopped", tunnel.Local.String())
		tunnel.Status = StopStatus
		tunnel.fire(StopStatus)
		tunnel.Hook.UnSubscribe()
	}()

	if err := tunnel.pool.Start(ctx); err != nil {
		return err
	}
	var errs = make(chan error)
	tunnel.Status = RunningStatus
	tunnel.fire(tunnel.Status)

	logrus.Infof("Tunnel to %s is started", tunnel.Local.String())

	connChan := make(chan net.Conn, 5)
	go func(errs chan<- error, conns chan<- net.Conn) {
		for {
			conn, err := listener.Accept()
			if err != nil {
				errs <- err
				return
			}
			connChan <- conn
		}
	}(errs, connChan)
	//
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tunnel.exit:
			return nil
		case err := <-errs:
			return err
		case conn := <-connChan:
			go tunnel.forward(conn)
		}
	}
}

func (tunnel *SSHtunnel) Stop() {
	tunnel.exit <- true
	logrus.Infof("Tunnel to %s is closed", tunnel.Local.String())
}

//func (t *Tunnels) Services() []Status {
//	status := make([]Status, 0, len(t.tunnels))
//	for _, htunnel := range t.tunnels {
//		status = append(status, Status{
//			Name:    htunnel.Name,
//			Running: htunnel.Running,
//		})
//	}
//
//	return status
//}

func (tunnel *SSHtunnel) forward(localConn net.Conn) {
	remote := tunnel.pool.Get()
	go tunnel.IOCopy(localConn, remote)
	tunnel.IOCopy(remote, localConn)
}
func (tunnel *SSHtunnel) IOCopy(writer, reader net.Conn) {
	defer func(writer net.Conn) {
		err := writer.Close()
		if err != nil {
			logrus.Errorf("close writer error: %s", err)
		}
	}(writer)
	defer func(reader net.Conn) {
		err := reader.Close()
		if err != nil {
			logrus.Errorf("close reader error: %s", err)
		}
	}(reader)

	_, err := io.Copy(writer, reader)
	if err != nil {
		logrus.Errorf("io.Copy error: %s", err)
	}
}

func sshAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

//func SSHPublic() ssh.AuthMethod {
//	pemBytes, err := ioutil.ReadFile("/home/hirosume/.ssh/id_rsa")
//	if err != nil {
//		return nil
//	}
//	signer, err := signerFromPem(pemBytes, []byte(""))
//	if err != nil {
//		return nil
//	}
//	return ssh.PublicKeys(signer)
//}
//
//func signerFromPem(pemBytes []byte, password []byte) (ssh.Signer, error) {
//
//	// read pem block
//	err := errors.New("Pem decode failed, no key found")
//	pemBlock, _ := pem.Decode(pemBytes)
//	if pemBlock == nil {
//		return nil, err
//	}
//
//	// generate signer instance from plain key
//	signer, err := ssh.ParsePrivateKey(pemBytes)
//	if err != nil {
//		return nil, fmt.Errorf("Parsing plain private key failed %v", err)
//	}
//
//	return signer, nil
//}
