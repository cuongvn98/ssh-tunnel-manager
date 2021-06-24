package tray

import (
	"context"
	"github.com/getlantern/systray"
	"ocg-ssh-tunnel/tunnel"
)

type MenuGroupItem struct {
	items     []*TunnelMenuItem
	item      *systray.MenuItem
	lastState int
}

func NewMenuGroupItem(item *systray.MenuItem) *MenuGroupItem {
	return &MenuGroupItem{item: item, items: make([]*TunnelMenuItem, 0), lastState: tunnel.StopStatus}
}

func (mgi *MenuGroupItem) Add(m *TunnelMenuItem) {
	mgi.items = append(mgi.items, m)
}

func (mgi *MenuGroupItem) Start(ctx context.Context) {
	mgi.item.SetTitle("Switch All")

	for _, item := range mgi.items {
		go item.Start(ctx)
	}

	for {
		select {
		case <-ctx.Done():
			break
		case <-mgi.item.ClickedCh:
			mgi.switchState(ctx)
		}
	}
}
func (mgi *MenuGroupItem) switchState(ctx context.Context) {
	if mgi.nextState() == tunnel.StopStatus {
		for _, item := range mgi.items {
			if item.lastState != tunnel.StopStatus {
				item.state.Switch(ctx)
			}
		}
	} else {
		for _, item := range mgi.items {
			if item.lastState == tunnel.StopStatus {
				item.state.Switch(ctx)
			}
		}
	}
	mgi.lastState = mgi.nextState()
}
func (mgi *MenuGroupItem) nextState() int {
	if mgi.lastState == tunnel.StopStatus {
		return tunnel.RunningStatus
	}
	return tunnel.StopStatus
}
