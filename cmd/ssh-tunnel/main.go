package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"ssh-tunnel/config"
	"ssh-tunnel/ssh"
)

func main() {
	conf, err := config.LoadConfig(filepath.Join(os.Getenv("HOME"), "ssh-tunnel.yaml"))
	if err != nil {
		panic(err)
	}

	if conf.Debug != nil && !*conf.Debug {
		log.SetOutput(io.Discard)
	}

	tunnel, err := ssh.NewServers(conf.Servers)
	if err != nil {
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	go tunnel.Start(ctx)

	log.Println("start ssh tunnel")

	<-ctx.Done()
	log.Println("stop ssh tunnel")
	err = tunnel.Stop()
	if err != nil {
		log.Printf("failed to stop tunnels: %v\n", err)
	}
	time.Sleep(time.Second)
}
