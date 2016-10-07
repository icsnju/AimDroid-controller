package main

import (
	"monidroid/android"
	"monidroid/config"
	"monidroid/util"
	"net"
	"time"
)

const (
	GUIDER           = "127.0.0.1:8024"
	APE              = "127.0.0.1:8024"
	MY_GUIDER_PORT   = "8024"
	YOUR_GUIDER_PORT = "1909"
	MY_APE_PORT      = "8025"
	YOUR_APE_PORT    = "8025"

	GUIDER_PACKAGE_NAME = "com.tianchi.monidroid"
	GUIDER_MAIN_NAME    = "com.tianchi.monidroid.MainActivity"
)

func main() {

	//init configuration
	config.InitConfig()

	//start ape server
	ape := startApeServer()

	//start guider server
	guider := startGuiderServer()

	//start test
	test.startTest(ape, guider)
}

//Start ape server
func startApeServer() *net.TCPConn {
	//Adb forward tcp
	err := android.Forward(config.GetSDKPath(), MY_APE_PORT, YOUR_APE_PORT)
	util.FatalCheck(err)

	//Start Ape server
	go android.StartApe(config.GetSDKPath(), YOUR_APE_PORT)
	ape := connectToServer(APE)
	return ape
}

//Start guider server
func startGuiderServer() *net.TCPConn {
	//Adb forward tcp
	err := android.Forward(config.GetSDKPath(), MY_GUIDER_PORT, YOUR_GUIDER_PORT)
	util.FatalCheck(err)

	//start guider service in mobile
	err = android.LaunchApp(config.GetSDKPath(), GUIDER_PACKAGE_NAME, GUIDER_MAIN_NAME)
	util.FatalCheck(err)
	time.Sleep(time.Millisecond * 1000)

	//setup socket connection
	tcpAddr, err := net.ResolveTCPAddr("tcp4", GUIDER)
	util.FatalCheck(err)

	service, err := net.DialTCP("tcp", nil, tcpAddr)
	util.FatalCheck(err)
	return service
}

//Connect to Server
func connectToServer(name string) *net.TCPConn {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", name)
	util.FatalCheck(err)
	server, err := net.DialTCP("tcp", nil, tcpAddr)
	util.FatalCheck(err)
	return server
}
