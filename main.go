package main

import (
	"log"
	"monidroid/android"
	"monidroid/util"
	"os"
)

var PKG_PATH = "pkg.txt"
var TARGET_PATH = "target.txt"
var DEVICE_PKG_PATH = "/sdcard/pkg.txt"
var DEVICE_TARGET_PATH = "/sdcard/target.txt"

var activityQueue *android.ActivityQueue

func main() {
	//create activity queue
	activityQueue = android.NewQueue()
	//init configuration
	initConfig()

	//create package information file
	//	file, err := os.OpenFile(PKG_PATH, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	//	util.FatalCheck(err)

	//	_, err = file.WriteString(getPackageName())
	//	util.FatalCheck(err)
	//	file.Close()

	//	//push apk file to device
	//	android.PushFile(getSDKPath(), PKG_PATH, DEVICE_PKG_PATH)

	//start logcat
	go android.StartLogcat(getSDKPath(), activityQueue)

	isFirst := true
	for !activityQueue.IsEmpty() || isFirst {
		android.KillApp(getSDKPath(), getPackageName())
		if !isFirst {
			//create package information file
			file, err := os.OpenFile(TARGET_PATH, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
			util.FatalCheck(err)
			activity := activityQueue.Dequeue()
			name, intent := activity.Get()
			log.Println("[Test]", name)
			_, err = file.WriteString(name + "\n")
			util.FatalCheck(err)
			_, err = file.WriteString(intent)
			util.FatalCheck(err)

			file.Sync()
			file.Close()
			android.PushFile(getSDKPath(), TARGET_PATH, DEVICE_TARGET_PATH)
		} else {
			isFirst = false
		}
		//android.LaunchApp(getSDKPath(), getPackageName(), getMainActivity())
		//time.Sleep(time.Millisecond * 1000)

		//TODO: start send event
		android.StartMonkey(getSDKPath(), getPackageName())
		//log.Println("[Monkey]", out)
	}

	android.KillApp(getSDKPath(), getPackageName())
	//remove target file
	android.RemoveFile(getSDKPath(), DEVICE_TARGET_PATH)

	log.Println(activityQueue.ToString())
}
