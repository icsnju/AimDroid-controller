package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Configure struct {
	PackageName  string
	MainActivity string
	SDKPath      string
	Epsilon      float64
	Alpha        float64
	Gamma        float64
	MaxSeqLen    int
	MinSeqLen    int
	Time         int
	ClearData    bool
}

var configPath = "configure.json"
var configuration *Configure = &Configure{"", "", "", 0.2, 0.5, 0.5, 50, 10, 1800, false}

func InitConfig() {
	log.Println("Init configuration..")
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

func GetPackageName() string {
	return configuration.PackageName
}

func SetPackageName(pkg string) {
	configuration.PackageName = pkg
}

func GetMainActivity() string {
	return configuration.MainActivity
}

func SetMainActivity(act string) {
	configuration.MainActivity = act
}

func GetSDKPath() string {
	return configuration.SDKPath
}

func GetEpsilon() float64 {
	return configuration.Epsilon
}

func GetAlpha() float64 {
	return configuration.Alpha
}

func GetGamma() float64 {
	return configuration.Gamma
}

func GetMaxSeqLen() int {
	return configuration.MaxSeqLen
}

func GetMinSeqLen() int {
	return configuration.MinSeqLen
}

func GetTime() int {
	return configuration.Time
}

func GetClearData() bool {
	return configuration.ClearData
}
