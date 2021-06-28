package tray

import (
	"context"
	"github.com/getlantern/systray"
	"ocg-ssh-tunnel/config"
	"ocg-ssh-tunnel/icon"
	"ocg-ssh-tunnel/tunnel"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var ctx context.Context
var cancel func()

func init() {
	ctx, cancel = context.WithCancel(context.Background())
}

func OnReady(groups []config.Group) func() {
	return func() {
		systray.SetIcon(icon.Data)
		systray.SetTooltip("SSH tunnel manager")

		for _, groupData := range groups {
			dev := systray.AddMenuItem(groupData.Name, "")
			systray.AddSeparator()
			group := NewMenuGroupItem(dev.AddSubMenuItem("", ""))
			for _, service := range groupData.Services {

				item := dev.AddSubMenuItem(service.Name, "")

				tunItem := NewTunnelMenuItem(
					item,
					service.Name,
					service.Username,
					tunnel.TransfromEndpointFromConfig(service.Local),
					tunnel.TransfromEndpointFromConfig(service.Remote),
					tunnel.TransfromEndpointFromConfig(service.Server),
				)
				group.Add(tunItem)
			}
			go group.Start(ctx)
		}

		mQuit := systray.AddMenuItem("Quit", "")

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

		for {
			select {
			case <-mQuit.ClickedCh:
				systray.Quit()
			case <-sig:
				systray.Quit()
			}
		}
	}
}

func OnQuit() {
	cancel()
	time.Sleep(2 * time.Second)
}
