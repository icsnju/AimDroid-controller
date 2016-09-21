package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Configure struct {
	PackageName  string
	MainActivity string
	SDKPath      string
}

var configPath = "configure.json"
var configuration *Configure = &Configure{}

func initConfig() {
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalln(err)
	}

	err = json.Unmarshal(content, configuration)
	if err != nil {
		log.Fatalln(err)
	}

	if len(configuration.MainActivity) <= 0 || len(configuration.PackageName) <= 0 || len(configuration.SDKPath) <= 0 {
		log.Fatalln("Configuration error:", configuration.MainActivity, configuration.PackageName, configuration.SDKPath)
	}
}

func getPackageName() string {
	return configuration.PackageName
}

func getMainActivity() string {
	return configuration.MainActivity
}

func getSDKPath() string {
	return configuration.SDKPath
}
