package ssh

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/crypto/ssh"
	"ssh-tunnel/config"
)

func NewServers(svs []config.Server) (Tunnels, error) {
	err := checkConflict(svs)
	if err != nil {
		return nil, fmt.Errorf("failed to check conflict: %w", err)
	}

	if len(svs) == 0 {
		return nil, errors.New("no server")
	}

	return NewTunnels(svs)
}

func checkConflict(svs []config.Server) error {
	addressMap := make(map[string]string, len(svs))
	for _, s := range svs {
		if hash, ok := addressMap[s.ServerAddress]; ok {
			if hash != s.GetHash() {
				return errors.New("duplicate server address")
			}
		}
		addressMap[s.ServerAddress] = s.GetHash()
	}

	serverMap := make(map[string]struct{}, len(svs))

	for i := range svs {
		for _, e := range svs[i].Endpoints {
			if _, ok := serverMap[e.LocalAddress]; ok {
				return errors.New("duplicate remote address")
			}
			serverMap[e.RemoteAddress] = struct{}{}
		}
	}
	return nil
}

func newSSHClientMap(svs []config.Server) (map[string]*ssh.Client, error) {
	sshClientMap := make(map[string]*ssh.Client, len(svs))

	for i := range svs {
		cl, err := NewSSHClient(svs[i])
		if err != nil {
			// close all clients
			for _, client := range sshClientMap {
				err := client.Close()
				if err != nil {
					log.Printf("failed to close ssh client: %v\n", err)
				}
			}

			return nil, fmt.Errorf("failed to create ssh client: %w", err)
		}
		sshClientMap[svs[i].GetHash()] = cl
	}

	return sshClientMap, nil
}

type Tunnels map[string]*TunnelConnection

func NewTunnels(svs []config.Server) (Tunnels, error) {
	sshClientMap, err := newSSHClientMap(svs)
	if err != nil {
		return nil, fmt.Errorf("failed to create ssh client map: %w", err)
	}

	tunnels := make(map[string]*TunnelConnection)

	for i := range svs {
		for j := range svs[i].Endpoints {
			endpoint := svs[i].Endpoints[j]
			client := sshClientMap[svs[i].GetHash()]
			tunnelConnection, err := NewTunnelConnection(endpoint, client)
			if err != nil {
				return nil, fmt.Errorf("failed to init tunnel connection: %w", err)
			}
			tunnels[endpoint.GetHash()] = tunnelConnection
		}
	}

	return tunnels, nil
}

func (t *Tunnels) Start(ctx context.Context) {
	var wg sync.WaitGroup
	for _, connection := range *t {
		wg.Add(1)
		go func(connection *TunnelConnection) {
			wg.Done()
			connection.Start(ctx)
		}(connection)
	}

	// wg.Add(1)
	//go func() {
	//	wg.Wait()
	//	ticker := time.NewTicker(time.Second * 5)
	//	defer log_terminal.Reset()
	//	for {
	//		select {
	//		case <-ctx.Done():
	//			return
	//		case <-ticker.C:
	//			for _, connection := range *t {
	//				log_terminal.Clear()
	//				totalBytes := connection.GetReadBytes()
	//				log_terminal.Printf(
	//					"%s -> %s: %d bytes\n",
	//					connection.localAddress,
	//					connection.remoteAddress,
	//					totalBytes,
	//				)
	//				log_terminal.Show()
	//			}
	//		}
	//	}
	//}()

	wg.Wait()
}

func (t *Tunnels) Stop() error {
	var result error
	for _, connection := range *t {
		err := connection.Close()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}
