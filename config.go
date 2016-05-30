package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

var config struct {
	TFL struct {
		AppID  string
		AppKey string
	}
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
