package config

import (
	"flag"
	"fmt"
	"os"
)

var Conf = &MsConfig{
	logger: msgo.Default(),
}

type MsConfig struct {
	logger *msgo.Logger
	Log    map[string]any
	Pool   map[string]any
}

func init() {
	loadToml()
}

func loadToml() {
	configFile := flag.String("conf", "conf/app.toml", "app config file")
	flag.Parse()
	if _, err := os.Stat(*configFile); err != nil {
		fmt.Println("conf/app.toml not load, not exist")
		return
	}

	_, err := toml.DecodeFile(*configFile, Conf)
	if err != nil {
		fmt.Println("conf/app.toml decode fail")
		return
	}

}
