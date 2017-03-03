package main

import (
	"bufio"
	"log"
	"monidroid/android"
	"monidroid/config"
	"monidroid/test"
	"monidroid/trace"
	"net"
	"os"
	"time"
)

const (
	GUIDER           = "127.0.0.1:8123"
	APE              = "127.0.0.1:8025"
	MY_GUIDER_PORT   = "8123"
	YOUR_GUIDER_PORT = "50122"
	MY_APE_PORT      = "8025"
	YOUR_APE_PORT    = "8025"

	GUIDER_PACKAGE_NAME = "com.tianchi.monidroid"
	GUIDER_MAIN_NAME    = "com.tianchi.monidroid.MainActivity"
)

func main() {

	t1 := time.Now()
	//init configuration
	config.InitConfig()
	if len(os.Args) >= 3 {
		config.SetPackageName(os.Args[1])
		config.SetMainActivity(os.Args[2])
	}
	android.InitADB(config.GetSDKPath())

	//push config to the device
	trace.InitADB(config.GetSDKPath())
	trace.PushConfig(config.GetPackageName())

	//start ape server
	apeIn, apeOut := startApeServer()
	defer closeApe(apeIn)

	//start guider server
	guider := startGuiderServer()
	defer closeGuider(guider)

	//start test
	test.Start(apeIn, guider, apeOut)

	t2 := time.Now()
	log.Println(t2.Sub(t1).Seconds())

}

//Start ape server
func startApeServer() (*net.TCPConn, *bufio.Reader) {
	log.Println("Start ape server..")
	android.KillApe()
	//Adb forward tcp
	err := android.Forward(MY_APE_PORT, YOUR_APE_PORT)
	if err != nil {
		log.Fatalln("startApeServer adb forward:", err)
	}
	//Start Ape server
	apeOut, err := android.StartApe(YOUR_APE_PORT)
	if err != nil {
		log.Fatalln("startApeServer StartApe:", err)
	}

	time.Sleep(time.Second * 3)
	apeIn := connectToServer(APE)
	return apeIn, apeOut
}

//Start guider server
func startGuiderServer() *net.TCPConn {
	//Adb forward tcp
	err := android.Forward(MY_GUIDER_PORT, YOUR_GUIDER_PORT)
	if err != nil {
		log.Fatalln("startGuiderServer:", err)
	}

	log.Println("Start guider server..")
	android.ClearApp(GUIDER_PACKAGE_NAME)

	//start guider service in mobile
	err = android.LaunchApp(GUIDER_PACKAGE_NAME, GUIDER_MAIN_NAME)
	if err != nil {
		log.Fatalln("startGuiderServer:", err)
	}
	time.Sleep(time.Millisecond * 1000)

	//setup socket connection
	tcpAddr, err := net.ResolveTCPAddr("tcp4", GUIDER)
	if err != nil {
		log.Fatalln("startGuiderServer:", err)
	}

	service, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatalln("startGuiderServer:", err)
	}
	return service
}

//Connect to Server
func connectToServer(name string) *net.TCPConn {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", name)
	if err != nil {
		log.Fatalln("connectToServer:", err)
	}
	server, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatalln("connectToServer:", err)
	}
	return server
}

func closeApe(ape *net.TCPConn) {
	ape.Write([]byte("quit\n"))
	ape.Close()
}

func closeGuider(guider *net.TCPConn) {
	keys := "stop"
	guider.SetWriteDeadline(time.Now().Add(time.Minute))
	_, err := guider.Write([]byte(keys))
	log.Println("Stop guider..")
	if err != nil {
		log.Fatalln("closeGuider:", err)
	}
	guider.Close()
	//stop guider service
	android.ClearApp(GUIDER_PACKAGE_NAME)
}
