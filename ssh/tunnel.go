package ssh

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"

	"golang.org/x/crypto/ssh"
	"ssh-tunnel/config"
)

type TunnelConnection struct {
	listener net.Listener
	client   *ssh.Client
	remote   *net.TCPAddr

	localAddress  string
	remoteAddress string

	totalReadBytes  int64
	totalWriteBytes int64
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

func (t *TunnelConnection) Start(ctx context.Context) {
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

			t.tunnel(conn, remoteConn)
		}
	}
}

func (t *TunnelConnection) Close() error {
	return t.listener.Close()
}

func (t *TunnelConnection) tunnel(localConn, remoteConn net.Conn) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	copyConn := func(writer, reader net.Conn, bytesRecorder *int64) {
		defer wg.Done()
		t.copyAndMeasureThroughput(writer, reader, bytesRecorder)
	}

	go copyConn(localConn, remoteConn, &t.totalReadBytes)
	go copyConn(remoteConn, localConn, &t.totalWriteBytes)
	go func() {
		wg.Wait()
		log.Printf("tunnel closed : %s %s", localConn.RemoteAddr(), remoteConn.RemoteAddr())
		_ = localConn.Close()
		_ = remoteConn.Close()
	}()
}

func (t *TunnelConnection) copyAndMeasureThroughput(writer, reader net.Conn, recorder *int64) {
	var err error
	written := int64(0)
	for {
		buffer := pool.Get().([]byte)
		bytesRead, readErr := reader.Read(buffer)
		if bytesRead > 0 {
			bytesWritten, writeErr := writer.Write(buffer[0:bytesRead])
			if bytesWritten > 0 {
				written += int64(bytesWritten)
			}
			if writeErr != nil {
				err = writeErr
				break
			}
			if bytesRead != bytesWritten {
				err = io.ErrShortWrite
				break
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			err = readErr
			break
		}
		atomic.AddInt64(recorder, written)
		written = 0
	}
	if err != nil {
		log.Println("Tunneling connection produced an error", err)
		return
	}
	atomic.AddInt64(recorder, written)
}

func (t *TunnelConnection) GetReadBytes() int64 {
	return atomic.LoadInt64(&t.totalReadBytes)
}

func (t *TunnelConnection) GetWriteBytes() int64 {
	return atomic.LoadInt64(&t.totalWriteBytes)
}
