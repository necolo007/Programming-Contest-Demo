package config

import (
	"Programming-Demo/pkg/fs"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"os"
)

var serveConfig *GlobalConfig

func LoadConfig(configYml string) {
	println("configYml file path: ", configYml)
	if !fs.FileExist(configYml) {
		println("cannot find config file")
		os.Exit(1)
	}
	serveConfig = new(GlobalConfig)
	viper.SetConfigFile(configYml)
	err := viper.ReadInConfig()
	if err != nil {
		println("Config Read failed: " + err.Error())
		os.Exit(1)
	}
	err = viper.Unmarshal(serveConfig)
	if err != nil {
		println("Config Unmarshal failed: " + err.Error())
		os.Exit(1)
	}
	viper.OnConfigChange(func(e fsnotify.Event) {
		println("Config fileHandle changed: ", e.Name)
		_ = viper.ReadInConfig()
		err = viper.Unmarshal(serveConfig)
		if err != nil {
			println("New Config fileHandle Parse Failed: ", e.Name)
			return
		}
	})
	viper.WatchConfig()
}

func GetConfig() *GlobalConfig {
	return serveConfig
}
