package tray

import (
	"context"
	"fmt"
	"github.com/getlantern/systray"
	"github.com/sirupsen/logrus"
	"ocg-ssh-tunnel/tunnel"
)

type TunnelMenuItem struct {
	item      *systray.MenuItem
	name      string
	state     *tunnel.StateMachine
	lastState int
}

func NewTunnelMenuItem(
	item *systray.MenuItem,
	name string,
	username string,
	endpoint tunnel.Endpoint,
	remote tunnel.Endpoint,
	server tunnel.Endpoint,
) *TunnelMenuItem {
	t := &TunnelMenuItem{
		item: item,
		name: name,
		//state:     state,
		lastState: tunnel.StopStatus,
	}
	hook := func(status int) {
		t.lastState = status
		switch status {
		case tunnel.StartingStatus:
			item.SetTitle(fmt.Sprintf("Starting %s ...", name))
		case tunnel.RunningStatus:
			item.SetTitle(fmt.Sprintf("Stop %s", name))
		case tunnel.StopStatus:
			item.SetTitle(fmt.Sprintf("Start %s", name))
		}
	}

	state := tunnel.NewStateMachine(username, endpoint, remote, server, hook)
	t.state = state

	return t
}

func (t *TunnelMenuItem) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			t.state.Stop()
			logrus.Info("Close tunnel")
			return
		case <-t.item.ClickedCh:
			logrus.Info("Switch")
			t.state.Switch(ctx)
		}
	}
}
