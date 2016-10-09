package main

import (
	"log"
	"monidroid/android"
	"monidroid/config"
	"monidroid/test"
	"monidroid/util"
	"net"
	"time"
)

const (
	GUIDER           = "127.0.0.1:8024"
	APE              = "127.0.0.1:8025"
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
	android.InitADB(config.GetSDKPath())

	//start ape server
	ape := startApeServer()
	defer closeApe(ape)

	//start guider server
	guider := startGuiderServer()
	defer closeGuider(guider)

	//start test
	test.Start(ape, guider)
}

//Start ape server
func startApeServer() *net.TCPConn {
	log.Println("Start ape server..")
	//Adb forward tcp
	err := android.Forward(MY_APE_PORT, YOUR_APE_PORT)
	util.FatalCheck(err)

	//Start Ape server
	go android.StartApe(YOUR_APE_PORT)
	time.Sleep(time.Second * 2)
	ape := connectToServer(APE)
	return ape
}

//Start guider server
func startGuiderServer() *net.TCPConn {
	log.Println("Start guider server..")
	//Adb forward tcp
	err := android.Forward(MY_GUIDER_PORT, YOUR_GUIDER_PORT)
	util.FatalCheck(err)

	//start guider service in mobile
	err = android.LaunchApp(GUIDER_PACKAGE_NAME, GUIDER_MAIN_NAME)
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

func closeApe(ape *net.TCPConn) {
	ape.Write([]byte("quit\n"))
	ape.Close()
}

func closeGuider(guider *net.TCPConn) {
	guider.Close()
	//stop guider service
	android.KillApp(GUIDER_PACKAGE_NAME)
}
