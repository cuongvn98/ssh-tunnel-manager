package main

import (
	"os"
	"os/signal"
	"syscall"

	"ssh-tunnel/config"
	"ssh-tunnel/ssh"
)

func main() {
	conf, err := config.LoadConfig("config.yaml")
	if err != nil {
		panic(err)
	}

	clean, err := ssh.NewServers(conf.Servers)
	if err != nil {
		panic(err)
	}

	defer clean()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}
