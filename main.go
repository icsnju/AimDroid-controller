package main

import (
	"log"
	"monidroid/android"
	"monidroid/util"
	"net"
	"time"
)

var activityQueue *android.ActivityQueue

const (
	SERVER              = "127.0.0.1:8024"
	MY_PORT             = "8024"
	YOUR_PORT           = "1909"
	GUIDER_PACKAGE_NAME = "com.tianchi.monidroid"
	GUIDER_MAIN_NAME    = "com.tianchi.monidroid.MainActivity"
	PACKAGE_NAME_KEY    = "pkgname"
	BLOCK_KEY           = "block"
	TARGET_KEY          = "target"
	INTENT_KEY          = "intent"
)

func main() {
	//create activity queue
	activityQueue = android.NewQueue()
	//init configuration
	initConfig()

	//adb forward
	err := android.Forward(getSDKPath(), MY_PORT, YOUR_PORT)
	util.FatalCheck(err)

	//start guider service in mobile
	err = android.LaunchApp(getSDKPath(), GUIDER_PACKAGE_NAME, GUIDER_MAIN_NAME)
	util.FatalCheck(err)
	time.Sleep(time.Millisecond * 1000)

	//setup socket connection
	tcpAddr, err := net.ResolveTCPAddr("tcp4", SERVER)
	util.FatalCheck(err)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	util.FatalCheck(err)

	//send pkgname
	_, err = conn.Write([]byte(PACKAGE_NAME_KEY + "#" + getPackageName() + "\n"))

	util.FatalCheck(err)
	//init other key
	conn.SetWriteDeadline(time.Now().Add(time.Minute))
	_, err = conn.Write([]byte(BLOCK_KEY + "#false\n"))
	util.FatalCheck(err)
	_, err = conn.Write([]byte(TARGET_KEY + "\n"))
	util.FatalCheck(err)
	_, err = conn.Write([]byte(INTENT_KEY + "\n"))
	util.FatalCheck(err)

	//start logcat
	go android.StartLogcat(getSDKPath(), activityQueue)

	//It is first time to launch this app
	android.LaunchApp(getSDKPath(), getPackageName(), getMainActivity())
	time.Sleep(time.Millisecond * 1000)

	//Get currently focused activity
	root := activityQueue.GetFocusedActivity()
	activityQueue.AddActivityInSet(root)
	conn.SetWriteDeadline(time.Now().Add(time.Minute))
	_, err = conn.Write([]byte(BLOCK_KEY + "#true\n"))
	util.FatalCheck(err)
	_, err = conn.Write([]byte(TARGET_KEY + "#" + root + "\n"))
	util.FatalCheck(err)
	log.Println("[Test]", root)
	android.StartMonkey(getSDKPath(), getPackageName())

	for !activityQueue.IsEmpty() {
		//kill this app started last time
		android.KillApp(getSDKPath(), getPackageName())

		//get an activity to start
		act := activityQueue.Dequeue()
		name, intent := act.Get()

		//init key
		conn.SetWriteDeadline(time.Now().Add(time.Minute))
		_, err = conn.Write([]byte(BLOCK_KEY + "#false\n"))
		util.FatalCheck(err)
		_, err = conn.Write([]byte(TARGET_KEY + "#" + name + "\n"))
		util.FatalCheck(err)
		_, err = conn.Write([]byte(INTENT_KEY + "#" + intent + "\n"))
		util.FatalCheck(err)
		//launch app
		android.LaunchApp(getSDKPath(), getPackageName(), getMainActivity())
		time.Sleep(time.Millisecond * 1000)

		conn.SetWriteDeadline(time.Now().Add(time.Minute))
		_, err = conn.Write([]byte(BLOCK_KEY + "#true\n"))
		util.FatalCheck(err)

		//TODO: start send event
		log.Println("[Test]", name)
		android.StartMonkey(getSDKPath(), getPackageName())
		//log.Println("[Monkey]", out)
		//time.Sleep(time.Millisecond * 10000)

	}
	conn.Close()

	android.KillApp(getSDKPath(), getPackageName())
	android.KillApp(getSDKPath(), GUIDER_PACKAGE_NAME)

	log.Println(activityQueue.ToString())
}
