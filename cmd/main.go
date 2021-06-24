package main

import (
	"github.com/getlantern/systray"
	"ocg-ssh-tunnel/tray"
)

func main() {

	systray.Run(tray.OnReady, tray.OnQuit)

	//tun := tunnel.NewSSHTunnel("cuongvu", tunnel.Endpoint{
	//	Host: "localhost",
	//	Port: 3306,
	//}, tunnel.Endpoint{
	//	Host: "localhost",
	//	Port: 3306,
	//}, tunnel.Endpoint{
	//	Host: "54.213.148.49",
	//	Port: 16889,
	//})
	//
	//ctx := context.Background()
	//ctx, cancel := context.WithCancel(ctx)
	//
	//err := tun.Start(ctx)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//c := make(chan os.Signal)
	//signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	//<-c
	//cancel()
	//time.Sleep(1000)

}
