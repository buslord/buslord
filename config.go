package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

var config struct {
	Cache struct {
		Enabled         bool     `yaml:"enabled"`
		MemcacheServers []string `yaml:"memcache-servers"`
	} `yaml:"cache"`
	TFL struct {
		AppID  string `yaml:"app-id"`
		AppKey string `yaml:"app-key"`
	} `yaml:"tfl"`
}

func init() {
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatal("Error reading config.yaml")
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatal("Error parsing config.yaml")
	}
}
