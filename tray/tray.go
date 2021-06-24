package tray

import (
	"context"
	"github.com/getlantern/systray"
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

func OnReady() {
	systray.SetIcon(icon.Data)
	systray.SetTooltip("OCG tunnel manager")
	dev := systray.AddMenuItem("Dev Tunnel", "Start dev tunnel")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit example tray application")

	group := NewMenuGroupItem(dev.AddSubMenuItem("", ""))
	systray.AddSeparator()
	{

		item := dev.AddSubMenuItem("Start MySQL", "MySQL")
		//elastic := dev.AddSubMenuItem("ElasticSearch", "Start ElasticSearch")

		tunItem := NewTunnelMenuItem(item, "MySQL", "cuongvu", tunnel.Endpoint{
			Host: "localhost",
			Port: 3306,
		}, tunnel.Endpoint{
			Host: "localhost",
			Port: 3306,
		}, tunnel.Endpoint{
			Host: "54.213.148.49",
			Port: 16889,
		})
		group.Add(tunItem)
		//go tunItem.Start(ctx)
	}
	{
		item := dev.AddSubMenuItem("Start Redis", "Redis")
		tunItem := NewTunnelMenuItem(item, "Redis", "cuongvu", tunnel.Endpoint{
			Host: "localhost",
			Port: 6379,
		}, tunnel.Endpoint{
			Host: "localhost",
			Port: 6379,
		}, tunnel.Endpoint{
			Host: "54.213.148.49",
			Port: 16889,
		})
		group.Add(tunItem)
		//go tunItem.Start(ctx)
	}
	{
		item := dev.AddSubMenuItem("Start RabbitMQ", "RabbitMQ")
		tunItem := NewTunnelMenuItem(item, "RabbitMQ", "cuongvu", tunnel.Endpoint{
			Host: "localhost",
			Port: 5672,
		}, tunnel.Endpoint{
			Host: "localhost",
			Port: 5672,
		}, tunnel.Endpoint{
			Host: "54.213.148.49",
			Port: 16889,
		})
		group.Add(tunItem)
		//go tunItem.Start(ctx)
	}
	go group.Start(ctx)

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

func OnQuit() {
	cancel()
	time.Sleep(2 * time.Second)
}
