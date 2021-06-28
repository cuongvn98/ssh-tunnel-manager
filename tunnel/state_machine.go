package tunnel

import (
	"context"
	"github.com/sirupsen/logrus"
)

type StateMachine struct {
	username string
	local    Endpoint
	remote   Endpoint
	server   Endpoint
	hook     func(int, string)
	tunnel   *SSHtunnel
}

func NewStateMachine(username string, local Endpoint, remote Endpoint, server Endpoint, hook func(int, string)) *StateMachine {
	return &StateMachine{username: username, local: local, remote: remote, server: server, hook: hook}
}

func (s *StateMachine) Switch(ctx context.Context) {
	if s.tunnel == nil {
		s.tunnel = NewSSHTunnel(s.username, s.local, s.remote, s.server)
		s.tunnel.Subscribe(s.hook)
	}

	switch s.tunnel.Status {
	case RunningStatus:
		s.Stop()
	case StopStatus:
		go func() {
			if err := s.tunnel.Start(ctx); err != nil {
				logrus.Infof("Failed while starting tunnel: %s", err)
			}
			s.Stop()
		}()
	}
}

func (s *StateMachine) Stop() {
	if s.tunnel != nil {
		s.tunnel.Stop()
		s.tunnel = nil
	}
}
