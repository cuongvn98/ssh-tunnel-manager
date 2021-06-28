package main

import (
	"fmt"
	"github.com/getlantern/systray"
	"github.com/spf13/viper"
	"ocg-ssh-tunnel/config"
	"ocg-ssh-tunnel/tray"
)

func main() {

	viper.SetConfigName("config")           // name of config file (without extension)
	viper.SetConfigType("yaml")             // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/ocgtunnel/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.ocgtunnel") // call multiple times to add many search paths
	viper.AddConfigPath(".")                // optionally look for config in the working directory
	err := viper.ReadInConfig()             // Find and read the config file
	if err != nil {                         // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}
	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("Can't load config file: %w \n", err))
	}

	systray.Run(tray.OnReady(cfg.Groups), tray.OnQuit)

}
