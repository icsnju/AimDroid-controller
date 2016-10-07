package test

import (
	"log"
	"monidroid/android"
	"monidroid/util"
	"net"
	"strings"
	"time"
)

const (
	PACKAGE_NAME_KEY = "pkgname"
	BLOCK_KEY        = "block"
	TARGET_KEY       = "target"
	INTENT_KEY       = "intent"
)

var activityQueue *android.ActivityQueue
var ape *net.TCPConn
var guider *net.TCPConn

//Start test
func Start(a, g *net.TCPConn) {
	//TODO:send pkgname
	//	_, err = service.Write([]byte(PACKAGE_NAME_KEY + "@" + getPackageName() + "\n"))
	//	util.FatalCheck(err)

	//create activity queue
	activityQueue = android.NewQueue()

	//init connection
	ape = a
	guider = g

	//start logcat
	go startObserver()

	//init key
	setKey("false", "", "")

	//It is first time to launch this app
	android.LaunchApp(getSDKPath(), getPackageName(), getMainActivity())
	time.Sleep(time.Millisecond * 1000)

	//Get currently focused activity
	root := activityQueue.GetFocusedActivity()
	activityQueue.AddActivityInSet(root)

	//set the key
	service.SetWriteDeadline(time.Now().Add(time.Minute))
	_, err = service.Write([]byte(BLOCK_KEY + "@true\n"))
	util.FatalCheck(err)
	_, err = service.Write([]byte(TARGET_KEY + "@" + root + "\n"))
	util.FatalCheck(err)

	//start test the root activity
	log.Println("[Test]", root)
	android.StartMonkey(getSDKPath(), getPackageName())

	//	for !activityQueue.IsEmpty() {
	//		//kill this app started last time
	//		android.KillApp(getSDKPath(), getPackageName())

	//		//get an activity to start
	//		act := activityQueue.Dequeue()
	//		name, intent := act.Get()

	//		//init key
	//		service.SetWriteDeadline(time.Now().Add(time.Minute))
	//		_, err = service.Write([]byte(BLOCK_KEY + "@false\n"))
	//		util.FatalCheck(err)
	//		_, err = service.Write([]byte(TARGET_KEY + "@" + name + "\n"))
	//		util.FatalCheck(err)
	//		_, err = service.Write([]byte(INTENT_KEY + "@" + intent + "\n"))
	//		util.FatalCheck(err)
	//		//launch app
	//		android.LaunchApp(getSDKPath(), getPackageName(), getMainActivity())
	//		time.Sleep(time.Millisecond * 1000)

	//		service.SetWriteDeadline(time.Now().Add(time.Minute))
	//		_, err = service.Write([]byte(BLOCK_KEY + "@true\n"))
	//		util.FatalCheck(err)

	//		//TODO: start send event
	//		log.Println("[Test]", name)
	//		android.StartMonkey(getSDKPath(), getPackageName())
	//		//log.Println("[Monkey]", out)
	//	}

	//clear the key
	//	service.SetWriteDeadline(time.Now().Add(time.Minute))
	//	_, err = service.Write([]byte(BLOCK_KEY + "@false\n"))
	//	util.FatalCheck(err)
	//	_, err = service.Write([]byte(TARGET_KEY + "\n"))
	//	util.FatalCheck(err)
	//	_, err = service.Write([]byte(INTENT_KEY + "\n"))
	//	util.FatalCheck(err)

	//	//close socket connection
	//	service.Close()
	//	//stop application
	//	android.KillApp(getSDKPath(), getPackageName())
	//	//stop guider service
	//	android.KillApp(getSDKPath(), GUIDER_PACKAGE_NAME)

	//	log.Println(activityQueue.ToString())
}

func setKey(block, target, intent string) {
	//set key
	guider.SetWriteDeadline(time.Now().Add(time.Minute))
	_, err = guider.Write([]byte(BLOCK_KEY + "@" + block + "\n"))
	util.FatalCheck(err)
	if len(target) <= 0 {
		_, err = guider.Write([]byte(TARGET_KEY + "\n"))
	} else {
		_, err = guider.Write([]byte(TARGET_KEY + "@" + target + "\n"))
	}
	util.FatalCheck(err)
	if len(intent) <= 0 {
		_, err = guider.Write([]byte(INTENT_KEY + "\n"))
	} else {
		_, err = guider.Write([]byte(INTENT_KEY + "@" + intent + "\n"))
	}
	util.FatalCheck(err)
}

func startObserver() {
	read, err := android.StartLogcat()
	util.FatalCheck(err)

	for {
		content, _, err := read.ReadLine()
		if err != nil {
			log.Println(err)
			break
		}
		if len(content) > 0 {
			iterms := strings.Split(string(content), "@")
			if len(iterms) >= 2 {
				switch iterms[1] {
				case LOG_START:
					if len(iterms) >= 4 {
						activityQueue.Enqueue(iterms[2], iterms[3])
					}
				case LOG_CREATE:
					activityQueue.SetFocusedActivity(iterms[2])
				case LOG_FINISH:
				default:
					//log.Println(content)
				}
			}
		}
	}
}
