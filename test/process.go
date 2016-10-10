package test

import (
	"log"
	"monidroid/android"
	"monidroid/config"
	"monidroid/util"
	"net"
	"time"
)

const (
	PACKAGE_NAME_KEY = "pkgname"
	BLOCK_KEY        = "block"
	TARGET_KEY       = "target"
	INTENT_KEY       = "intent"

	TRUE  = "true"
	FALSE = "false"

	APE_TREE = "tree"
	APE_SIZE = "size"
)

var gActivityQueue *ActivityQueue = nil
var ape *net.TCPConn = nil
var guider *net.TCPConn = nil
var mTest *Test = nil

//Start test
func Start(a, g *net.TCPConn) {
	//TODO:send pkgname
	//	_, err = service.Write([]byte(PACKAGE_NAME_KEY + "@" + getPackageName() + "\n"))
	//	util.FatalCheck(err)
	log.Println("Start test..")

	//create activity queue
	gActivityQueue = NewQueue()

	//init connection
	ape = a
	guider = g

	//start logcat
	go startObserver()

	//init key
	setKey(FALSE, "", "")

	//It is first time to launch this app
	android.LaunchApp(config.GetPackageName(), config.GetMainActivity())
	time.Sleep(time.Millisecond * 3000)

	//Get currently focused activity
	root := android.GetCurrentActivity()
	//Create the first activity
	gActivityQueue.Enqueue(root, "")

	//Get an activity to test
	for !gActivityQueue.IsEmpty() {

		//Step1: get an activity to start
		act := gActivityQueue.Dequeue()
		name, intent := act.Get()

		//Test record the testing state
		mTest = NewTest()
		mTest.Act = act

		if len(intent) <= 0 {
			//It's the first time to start this application
			//set the key
			setKey(TRUE, name, intent)
		} else {
			startThisActivity(name, intent)
		}
		log.Println("1. Start activity to generate actions..", act.name)

		if !currentActIsRight(name) {
			//In a wrong activity
			log.Println("You are in a wrong activity", android.GetCurrentActivity())
			continue
		}
		//Step2: generate initial actions set
		mTest.Cache.clear()
		sendCommandToApe(APE_TREE)
		time.Sleep(time.Millisecond * 1000)
		mTest.Cache.filterAction(mTest.ActSet)

		if mTest.ActSet.GetCount() <= 0 {
			log.Println("2.No acion is found!")
			continue
		}
		log.Println("2. Initial actions count:", mTest.ActSet.GetCount(), ", start to test activity..")

		//Step3: send event
		log.Println("3. Start send action")
		sequence := NewActionSequence()
		for i := 0; i < 2*mTest.ActSet.GetCount(); i++ {
			//clear log
			mTest.Cache.clear()
			//get an action
			action, index := mTest.ActSet.GetEpGreAction()
			if action == nil {
				log.Fatalln("No action found!")
				break
			}
			//send action
			sendActionToApe(action)
			time.Sleep(time.Millisecond * 500)

			//if it go out of the target activity
			isOut := !currentActIsRight(name)

			//get result
			rs := mTest.Cache.filterResult()

			//if nothing change
			if rs.GetKind() == R_CHANGE && !isOut {
				sendCommandToApe(APE_TREE)
				time.Sleep(time.Millisecond * 500)
				count := mTest.Cache.filterAction(mTest.ActSet)
				if count > 3 {
					cr, ok := rs.(*CommonResult)
					if ok {
						cr.SetKind(R_CHANGE)
						rs = cr
					}
				}
			}
			//Adjust reward of this action
			log.Println("Result: ", rs.ToString(), rs.GetKind())
			Reward(mTest.ActSet, index, rs)
			sequence.add(index, rs)

			if isOut {
				//Save this sequence
				mTest.SequenceArray = append(mTest.SequenceArray, sequence)
				sequence = NewActionSequence()
				//Restart this activity
				ok := startThisActivity(name, intent)
				if !ok {
					break
				}
			}
		} //finish send action

		if sequence.getCount() > 0 {
			mTest.SequenceArray = append(mTest.SequenceArray, sequence)
		}

		//log.Println("[Monkey]", out)
		if mTest != nil {
			mTest.Save("out/" + config.GetPackageName())
		}
	}

	//clear the key
	setKey(FALSE, "", "")

	//stop application
	android.KillApp(config.GetPackageName())

	log.Println(gActivityQueue.ToString())
}

//Start an activity
func startThisActivity(name, intent string) bool {
	//kill this app started last time
	android.KillApp(config.GetPackageName())
	//reset the key
	setKey(FALSE, name, intent)
	//launch app
	android.LaunchApp(config.GetPackageName(), config.GetMainActivity())
	time.Sleep(time.Millisecond * 3000)
	//set the key
	setKey(TRUE, name, intent)
	return currentActIsRight(name)
}

//If this current focused activity is right
func currentActIsRight(name string) bool {
	cn := android.GetCurrentActivity()
	count := 0
	for name != cn {
		count++
		if count > MAX_TRY {
			break
		}
		time.Sleep(time.Millisecond * 1000)
		cn = android.GetCurrentActivity()
	}
	return name == cn
}

//Set the key of guider
func setKey(block, target, intent string) {
	//set key
	guider.SetWriteDeadline(time.Now().Add(time.Minute))
	_, err := guider.Write([]byte(BLOCK_KEY + "@" + block + "\n"))
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

func sendCommandToApe(cmd string) {
	ape.SetWriteDeadline(time.Now().Add(time.Minute))
	_, err := ape.Write([]byte(cmd + "\n"))
	util.FatalCheck(err)
}

func sendActionToApe(a *Action) {
	log.Println("Send action: ", a.getContent())
	ape.SetWriteDeadline(time.Now().Add(time.Minute))
	_, err := ape.Write([]byte(a.getContent() + "\n"))
	util.FatalCheck(err)
}

func startObserver() {
	log.Println("Start observer..")

	read, err := android.StartLogcat()
	util.FatalCheck(err)
	for {
		content, _, err := read.ReadLine()
		if err != nil {
			log.Println(err)
			break
		}
		if len(content) > 0 {
			//TODO mTest is nil
			if mTest != nil {
				mTest.Cache.add(string(content))
			}
		}
	}
}

//Adjust reward of this action
func Reward(set *ActionSet, index int, result Result) {
	kind := result.GetKind()

	switch kind {
	case R_ACTIVITY:
		actRs, ok := result.(*ActivityResult)
		if !ok {
			log.Fatalln("Activity result err")
		}

		ok = gActivityQueue.Enqueue(actRs.GetContent())
		if ok {
			//It is a new activity
			set.AdjustReward(index, 1, 1)

			//Reward my siblings
			set.AdjustReward(index+1, 1, 0)
			set.AdjustReward(index-1, 1, 0)
		} else {
			//It is a old activity
			set.AdjustReward(index, 0, 1)
		}
	case R_FINISH:
		//I don't want to finish
		set.AdjustReward(index, -1, 1)
	case R_NOCHANGE:
		set.AdjustReward(index, 0, 1)
	case R_ERR:
		set.AdjustReward(index, 1, 1)
	case R_CHANGE:
		set.AdjustReward(index, 1, 1)
	default:
		log.Fatalln("Result is unknown, err")
	}
}
