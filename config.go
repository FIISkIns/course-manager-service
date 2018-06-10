package main

import "github.com/kelseyhightower/envconfig"

type ConfigurationSpec struct {
	Port        int    `default:"8001"`
	DatabaseUrl string `default:"Gafi:bagpicioarele@tcp(127.0.0.1:3306)/courses"`
}

var config ConfigurationSpec

func initConfig() {
	envconfig.MustProcess("manager", &config)
}
