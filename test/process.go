package test

import (
	"bufio"
	"log"
	"monidroid/android"
	"monidroid/config"
	"monidroid/util"
	"net"
	"path"
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
var crashReader *bufio.Reader = nil
var mTest *Test = nil

//Start test
func Start(a, g *net.TCPConn, cr *bufio.Reader) {
	//TODO:send pkgname
	//	_, err = service.Write([]byte(PACKAGE_NAME_KEY + "@" + getPackageName() + "\n"))
	//	util.FatalCheck(err)
	startTime := time.Now()
	finishTime := startTime.Add(time.Duration(config.GetTime()) * time.Second)
	log.Println("Start test..")

	//create activity queue
	gActivityQueue = NewQueue()

	//init connection
	ape = a
	guider = g
	crashReader = cr
	//	setKey(FALSE, "", "")
	//	return

	//start logcat
	go startObserver()
	go startCrashReader()

	//init key
	setKey(FALSE, "", "")

	//It is first time to launch this app
	android.ClearApp(config.GetPackageName())
	time.Sleep(time.Millisecond * 1000)
	android.LaunchApp(config.GetPackageName(), config.GetMainActivity())
	time.Sleep(time.Millisecond * 3000)

	//Get currently focused activity
	root := android.GetCurrentActivity()
	//Create the first activity
	gActivityQueue.Enqueue(root, "", "", int64(time.Now().Sub(startTime).Seconds()))

	//Get an activity to test
	for time.Now().Before(finishTime) {

		//Step1: get an activity to start
		var activity *Activity = nil
		if gActivityQueue.IsEmpty() {
			activity = gActivityQueue.DeOldQueue()
		} else {
			activity = gActivityQueue.Dequeue()
		}
		if activity == nil {
			log.Println("No activity is in the queue!")
			break
		}
		name := activity.GetName()

		//Test record the testing state
		mTest = gActivityQueue.GetTest(name)
		if mTest == nil {
			log.Println("Cannot find test of this activity!")
			continue
		}

		mTest.Cache.clear()

		log.Println("1. Start activity to generate actions..", activity.name)
		//Step1.2: start this activity
		startThisActivity(activity, mTest.HaveCrash)

		if !currentActIsRight(name) {
			//In a wrong activity
			rs := mTest.Cache.filterResult()
			log.Println("Cannot start activity. You are in a wrong activity", android.GetCurrentActivity(), rs.GetKind())

			cr, ok := rs.(*CrashResult)
			if ok {
				cout := path.Join("out", config.GetPackageName(), name, "Crash")
				cr.Save(cout)
				gActivityQueue.AddCrash(name+"@launch activity", -1)
				mTest.HaveCrash = true
			}

			gActivityQueue.EnOldQueue(activity)
			continue
		}
		//Step2: generate initial actions set
		mTest.Cache.clear()
		sendCommandToApe(APE_TREE)
		time.Sleep(time.Millisecond * 1000)
		mTest.Cache.filterAction(mTest.ActSet)

		if mTest.ActSet.GetCount() <= 0 {
			//Add some basic trackball events
			tb := NewAction("trackball 100 0")
			mTest.ActSet.AddAction(tb)
			tb = NewAction("trackball -100 0")
			mTest.ActSet.AddAction(tb)
			tb = NewAction("trackball 0 100")
			mTest.ActSet.AddAction(tb)
			tb = NewAction("trackball 0 -100")
			mTest.ActSet.AddAction(tb)
			tb = NewAction("key down 82")
			mTest.ActSet.AddAction(tb)
		} else {
			tb := NewAction("trackball 0 -100")
			mTest.ActSet.AddAction(tb)
			tb = NewAction("trackball -100 0")
			mTest.ActSet.AddAction(tb)
			tb = NewAction("key down 82")
			mTest.ActSet.AddAction(tb)
		}

		log.Println("2. Initial actions count:", mTest.ActSet.GetCount(), ", start to test activity..")

		//Step3: send event
		log.Println("3. Start send action")
		times := 1
		//Generate some testing sequences
		for times > 0 {
			times = 0
			sequence := NewActionSequence()

			//Create an action sequence
			for i := 0; ; i++ {
				if i > 2*mTest.ActSet.GetCount() && i > config.GetMinSeqLen() {
					break
				}
				//clear log
				mTest.Cache.clear()
				//get an action
				action, index := mTest.ActSet.GetEpGreAction()
				if action == nil {
					log.Fatalln("No action found!")
					break
				}
				//send action
				mTest.Cache.clear()
				sendActionToApe(action)
				log.Println("Send action:", action.content, action.getAveReward())
				time.Sleep(time.Millisecond * 1000)

				//if it go out of the target activity
				isOut := !currentActIsRight(name)

				//get result
				rs := mTest.Cache.filterResult()

				//If it is a crash
				if rs.GetKind() == R_CRASH {
					cr, ok := rs.(*CrashResult)
					if ok {
						ex := gActivityQueue.AddCrash(name+"@"+action.content, len(mTest.SequenceArray))
						//old crash
						if ex {
							cr.SetKind(R_NOCHANGE)
						}
						mTest.HaveCrash = true
					} else {
						cr.SetKind(R_NOCHANGE)
					}
				}

				//if nothing change
				if rs.GetKind() <= R_CHANGE && !isOut {
					mTest.Cache.clear()
					sendCommandToApe(APE_TREE)
					time.Sleep(time.Millisecond * 1000)
					count := mTest.Cache.filterAction(mTest.ActSet)
					//Little views change, so it is unchanged
					cr, ok := rs.(*CommonResult)
					if ok {
						if count < 3 {
							cr.SetKind(R_NOCHANGE)
						} else {
							cr.SetKind(R_CHANGE)
						}
						rs = cr

					}
				}

				//Step4. Adjust reward of this action
				feedback := Reward(mTest.ActSet, index, rs, name, int64(time.Now().Sub(startTime).Seconds()))
				//If you can find something new, we will loop again
				times += feedback
				log.Println("Adjust reward of this action: ", rs.GetKind(), feedback)

				sequence.add(index, rs)
				mTest.addEdge(rs, sequence.count)

				//Testing is out of this activity, so restart it
				//This aciton sequence it too long, let's start a new sequence
				if isOut || sequence.count > config.GetMaxSeqLen() {
					break
				}

			} //finish an action sequence

			if sequence.getCount() > 0 {
				mTest.SequenceArray = append(mTest.SequenceArray, sequence)
			}
			//Restart this activity
			if times > 0 {
				ok := startThisActivity(activity, mTest.HaveCrash)
				if !ok {
					break
				}
			}
		}
		//Step5. Save results of this activity
		//log.Println("[Monkey]", out)
		if mTest != nil {
			gActivityQueue.EnOldQueue(activity)
		}
	}

	//clear the key
	setKey(FALSE, "", "")

	//stop application
	android.ClearApp(config.GetPackageName())

	log.Println(gActivityQueue.ToString())
	gActivityQueue.Save("out/" + config.GetPackageName())
}

//Start an activity
func startThisActivity(act *Activity, haveCrash bool) bool {
	//kill this app started last time
	android.ClearApp(config.GetPackageName())
	time.Sleep(time.Millisecond * 500)
	if haveCrash {
		return startThisActivityFromParent(act.GetParent(), act.GetName())
	} else {
		return startThisActivityDirectly(act.Get())
	}
}

func startThisActivityDirectly(name, intent string) bool {
	//reset the key
	setKey(FALSE, name, intent)
	//launch app
	android.LaunchApp(config.GetPackageName(), config.GetMainActivity())
	time.Sleep(time.Millisecond * 2000)
	ok := currentActIsRight(name)
	//set the key
	setKey(TRUE, name, intent)
	return ok
}

func startThisActivityFromParent(parent, me string) bool {
	//Get parent firstly
	parentTest := gActivityQueue.GetTest(parent)
	ok := startThisActivity(parentTest.Act, parentTest.HaveCrash)
	if !ok {
		return ok
	}

	//find the sequence
	edge, ex := parentTest.Find[me]
	if !ex {
		return ex
	}
	index := edge.SeqIndex
	seq := parentTest.SequenceArray[index]
	intent := ""
	//replay this sequence
	for i, ai := range seq.sequence {
		action := parentTest.ActSet.queue[ai]
		rs, ex := seq.tag[i]
		find := false
		if ex {
			ar, ok := rs.(*ActivityResult)
			if ok && ar.name == me {
				setKey(FALSE, "", "")
				find = true
				intent = ar.intent
			}
		}
		sendCommandToApe(action.content)
		if find {
			break
		}
	}

	ok = currentActIsRight(me)
	//set the key
	setKey(TRUE, me, intent)
	return ok
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
	keys := BLOCK_KEY + "@" + block + "\n"

	if len(target) <= 0 {
		keys += TARGET_KEY + "\n"
	} else {
		keys += TARGET_KEY + "@" + target + "\n"
	}

	if len(intent) <= 0 {
		keys += INTENT_KEY + "\n"
	} else {
		keys += INTENT_KEY + "@" + intent + "\n"
	}

	guider.SetWriteDeadline(time.Now().Add(time.Minute))
	_, err := guider.Write([]byte(keys))
	util.FatalCheck(err)
	time.Sleep(1000 * time.Millisecond)
}

func sendCommandToApe(cmd string) {
	ape.SetWriteDeadline(time.Now().Add(time.Minute))
	_, err := ape.Write([]byte(cmd + "\n"))
	util.FatalCheck(err)
}

func sendActionToApe(a *Action) {
	//log.Println("Send action: ", a.getContent(), a.getAveReward())
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
			log.Println("Observer", err)
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

func startCrashReader() {
	log.Println("Start crash reader..")
	crashStart := false
	for {
		content, _, err := crashReader.ReadLine()
		if err != nil {
			log.Println("CrashReader", err)
			break
		}
		if len(content) > 0 {
			line := string(content)
			if line == LOG_CRASH {
				//TODO mTest is nil
				if mTest != nil {
					mTest.Cache.add("@" + line + "@")
					crashStart = true
				}
			} else if crashStart {
				if line == LOG_CRASH_END {
					mTest.Cache.add("@" + line + "@")
					crashStart = false
				} else {
					mTest.Cache.add(line)
				}
			}
		}
	}
}

//Adjust reward of this action
func Reward(set *ActionSet, index int, result Result, parent string, time int64) int {
	kind := result.GetKind()

	feedback := 0
	switch kind {
	case R_ACTIVITY:
		actRs, ok := result.(*ActivityResult)
		if !ok {
			log.Fatalln("Activity result err")
		}
		name, content := actRs.GetContent()
		ok = gActivityQueue.Enqueue(name, content, parent, time)
		if ok {
			//It is a new activity
			set.AdjustReward(index, 1, 1)

			//Reward my siblings
			set.AdjustReward(index+1, 1, 0)
			set.AdjustReward(index-1, 1, 0)
			feedback = 1
		} else {
			//It is a old activity
			set.AdjustReward(index, -1, 1)
		}
	case R_FINISH:
		//I don't want to finish
		set.AdjustReward(index, -1, 1)
	case R_NOCHANGE:
		set.AdjustReward(index, 0, 1)
	case R_CRASH:
		set.AdjustReward(index, 1, 1)
		feedback = 1
	case R_CHANGE:
		set.AdjustReward(index, 1, 1)
	default:
		log.Fatalln("Result is unknown, err")
	}
	return feedback
}
