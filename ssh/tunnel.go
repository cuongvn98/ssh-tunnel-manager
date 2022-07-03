package ssh

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"golang.org/x/crypto/ssh"
	"ssh-tunnel/config"
)

type TunnelConnection struct {
	listener net.Listener
	client   *ssh.Client
	remote   *net.TCPAddr

	localAddress  string
	remoteAddress string
}

func NewTunnelConnection(endpoint config.Endpoint, client *ssh.Client) (*TunnelConnection, error) {
	ln, err := net.Listen("tcp", endpoint.LocalAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	addr, err := net.ResolveTCPAddr("tcp", endpoint.RemoteAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve remote address: %w", err)
	}

	return &TunnelConnection{
		listener:      ln,
		client:        client,
		remote:        addr,
		localAddress:  endpoint.LocalAddress,
		remoteAddress: endpoint.RemoteAddress,
	}, nil
}

func (t *TunnelConnection) Tunnel(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("stop listening %s %s", t.localAddress, t.remoteAddress)
			return
		default:
			conn, err := t.listener.Accept()
			if err != nil {
				log.Printf("accept error %s %s %v", t.localAddress, t.remoteAddress, err)
				continue
			}

			log.Printf("accept %s %s", t.localAddress, t.remoteAddress)

			remoteConn, err := t.client.DialTCP("tcp", nil, t.remote)
			if err != nil {
				log.Printf("dial error %s %s", t.localAddress, t.remoteAddress)
				continue
			}

			tunnel(conn, remoteConn)
		}
	}
}

func (t *TunnelConnection) Close() error {
	return t.listener.Close()
}

func NewTunnel(c context.Context, endpoint config.Endpoint, server *ssh.Client) (func(), error) {
	remote := endpoint.RemoteAddress
	local := endpoint.LocalAddress

	addr, err := net.ResolveTCPAddr("tcp", remote)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve remote address: %w", err)
	}

	listener, err := net.Listen("tcp", local)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	log.Printf("listening on %s", local)

	go func() {
		for {
			select {
			case <-c.Done():
				listener.Close()
				log.Printf("stop listening %s %s", local, remote)
				return
			default:
				conn, err := listener.Accept()
				if err != nil {
					log.Printf("accept error %s %s %v", local, remote, err)
					continue
				}

				log.Printf("accept %s %s", local, remote)

				remoteConn, err := server.DialTCP("tcp", nil, addr)
				if err != nil {
					log.Printf("dial error %s %s", local, remote)
					continue
				}

				tunnel(conn, remoteConn)
			}
		}
	}()

	return func() {
		listener.Close()
	}, nil
}

func tunnel(localConn, remoteConn net.Conn) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	copyConn := func(writer, reader net.Conn) {
		defer wg.Done()
		CopyAndMeasureThroughput(writer, reader)
	}

	go copyConn(localConn, remoteConn)
	go copyConn(remoteConn, localConn)
	go func() {
		wg.Wait()
		log.Printf("tunnel closed : %s %s", localConn.RemoteAddr(), remoteConn.RemoteAddr())
		_ = localConn.Close()
		_ = remoteConn.Close()
	}()
}
