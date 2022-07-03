package ssh

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/ssh"
	"ssh-tunnel/config"
)

func NewServers(svs []config.Server) (func(), error) {
	err := checkConflict(svs)
	if err != nil {
		return nil, fmt.Errorf("failed to check conflict: %w", err)
	}

	svMap := make(map[string]*ssh.Client, len(svs))

	for i := range svs {
		sv := svs[i]
		cl, err := NewSSHClient(sv)
		if err != nil {
			return nil, fmt.Errorf("failed to create ssh client: %w", err)
		}
		svMap[sv.GetHash()] = cl
	}

	ctx, cancel := context.WithCancel(context.Background())

	for i := range svs {
		for j := range svs[i].Endpoints {
			endpoint := svs[i].Endpoints[j]
			sv := svMap[svs[i].GetHash()]
			_, err := NewTunnel(ctx, endpoint, sv)
			if err != nil {
				log.Printf("failed to create tunnel: %v\n", err)
			}
		}
	}

	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Println(Meter.GetHumanReadablePer10Seconds())
			}
		}
	}()

	return func() {
		cancel()
		for _, sv := range svMap {
			sv.Close()
		}
	}, nil
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
