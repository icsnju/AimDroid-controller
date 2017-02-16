package test

import (
	"bufio"
	"log"
	"math/rand"
	"monidroid/android"
	"monidroid/config"
	"monidroid/trace"
	"monidroid/util"
	"net"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	PACKAGE_NAME_KEY = "pkgname"
	BLOCK_KEY        = "block"
	TARGET_KEY       = "target"
	INTENT_KEY       = "intent"
	CHILD_KEY        = "child"

	TRUE  = "true"
	FALSE = "false"

	APE_TREE   = "tree"
	APE_LAUNCH = "launch"
	APE_SIZE   = "size"
)

var gActivityQueue *ActivityQueue = nil
var ape *net.TCPConn = nil
var guider *net.TCPConn = nil
var crashReader *bufio.Reader = nil
var mTest *Test = nil
var gLogCache *LogCache = nil
var gX int = 800
var gY int = 1280

var traceChan chan int = make(chan int)
var traceBackChan chan int = make(chan int)

var eventCount int = 0

//Start test
func Start(a, g *net.TCPConn, cr *bufio.Reader) {
	//TODO:send pkgname
	//	_, err = service.Write([]byte(PACKAGE_NAME_KEY + "@" + getPackageName() + "\n"))
	//	util.FatalCheck(err)
	startTime := time.Now()
	finishTime := startTime.Add(time.Duration(config.GetTime()) * time.Second)
	log.Println("Start test..", config.GetPackageName())

	//create activity queue
	gActivityQueue = NewQueue()
	gLogCache = NewLogCache()
	//init connection
	ape = a
	guider = g
	crashReader = cr

	//start logcat
	go startObserver()
	//start get crash from ape
	go startCrashReader()
	//start get coverage report
	go trace.StartTrace(config.GetPackageName(), startTime, traceChan, traceBackChan)

	//init key
	setKey(FALSE, "", "", "")

	//init x y
	sendCommandToApe(APE_SIZE)

	//It is first time to launch this app
	//	if config.GetClearData() {
	//		android.ClearApp(config.GetPackageName())
	//	} else {
	//		android.KillApp(config.GetPackageName())
	//	}
	android.KillApp(config.GetPackageName())

	android.LaunchApp(config.GetPackageName(), config.GetMainActivity())
	//sendCommandToApe(APE_LAUNCH + " " + config.GetPackageName() + " " + config.GetMainActivity())
	time.Sleep(time.Second * 10)

	//Get currently focused activity
	root := android.GetCurrentActivity(config.GetMainActivity())
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

		gLogCache.clearAll()

		//Step1.2: start this activity
		killStartThisActivity(activity, mTest.HaveCrash)
		log.Println("1. Start activity", activity.name)

		if !currentActIsRight(name, MAX_TRY) {
			//In a wrong activity
			rs := gLogCache.filterResult()
			log.Println("Cannot start activity. You are in a wrong activity", android.GetCurrentActivity(""), rs.GetKind())

			cr, ok := rs.(*CrashResult)
			if ok {
				cout := path.Join("out", config.GetPackageName(), name, "Crash")
				cr.Save(cout)
				gActivityQueue.AddCrash(name+"@launch activity", -1)
			}
			mTest.HaveCrash = true
			gActivityQueue.EnOldQueue(activity)
			continue
		}
		//Step2: generate initial actions set
		gLogCache.clearA()
		sendCommandToApe(APE_TREE)
		time.Sleep(time.Millisecond * 1000)
		gLogCache.filterAction(mTest.ActSet)

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

		log.Println("2. Initial actions count:", mTest.ActSet.GetCount(), ", start to test activity..")

		//Step3: send event
		log.Println("3. Start an action sequence..")
		times := 0
		sequence := NewActionSequence()

		var index int = 0
		_, index = mTest.ActSet.GetEpGreAction()
		//Create an action sequence
		for i := 0; ; i++ {
			if i > 2*mTest.ActSet.GetCount() && i > config.GetMinSeqLen() {
				break
			}

			//get an action
			action := mTest.ActSet.GetAction(index)
			if action == nil {
				log.Fatalln("No action found!")
				break
			}
			//send action
			//gLogCache.clearRC()
			eventCount += action.getEventCount()
			sendActionToApe(action)
			log.Println("["+strconv.Itoa(i)+"] Send action:", action.content, action.getQ())
			//time.Sleep(time.Millisecond * 500)

			//if it go out of the target activity
			isOut := !currentActIsRight(name, MAX_TRY)

			//Get current view
			if !isOut {
				gLogCache.clearA()
				sendCommandToApe(APE_TREE)
				time.Sleep(time.Millisecond * 1000)
				gLogCache.filterAction(mTest.ActSet)
				ifAddMenuKey := rand.Intn(10)
				if ifAddMenuKey <= 3 {
					tb = NewAction("key down 82")
					mTest.ActSet.AddAction(tb)
				}
			}

			//get result
			rs := gLogCache.filterResult()

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

			if isOut && rs.GetKind() <= R_FINISH {
				rs.SetKind(R_FINISH)
			}

			//Step4. Adjust Q value of this action
			_, index2 := mTest.ActSet.GetEpGreAction()
			feedback := Reward(mTest.ActSet, index, index2, rs, name, int64(time.Now().Sub(startTime).Seconds()))
			//If you can find something new, we will loop again
			times += feedback
			log.Println("Reward of this action: ", rs.GetKind(), feedback)

			sequence.add(index, rs)
			mTest.addEdge(rs, sequence.count)
			index = index2
			//Testing is out of this activity, so restart it
			//This aciton sequence it too long, let's start a new sequence
			if isOut || sequence.count > config.GetMaxSeqLen() {
				break
			}

		} //finish an action sequence

		if sequence.getCount() > 0 {
			mTest.SequenceArray = append(mTest.SequenceArray, sequence)
		}

		if times > 0 {
			//				ok := killStartThisActivity(activity, mTest.HaveCrash)
			//				if !ok {
			//					break
			//				}
			//				log.Println("4. I want to test this Activity again:", activity.GetName())
			//Test this activity again
			gActivityQueue.EnqueueAgain(activity)
		} else if mTest != nil {
			gActivityQueue.EnOldQueue(activity)
		}
	}

	//clear the key
	setKey(FALSE, "", "", "")

	//stop trace
	traceChan <- trace.TRACE_DUMP
	<-traceBackChan
	traceChan <- trace.TRACE_STOP
	<-traceBackChan

	//stop application
	//	if config.GetClearData() {
	//		android.ClearApp(config.GetPackageName())
	//	} else {
	//		android.KillProApp(config.GetPackageName())
	//	}
	android.KillApp(config.GetPackageName())

	log.Println(gActivityQueue.ToString())
	gActivityQueue.Save("out/"+config.GetPackageName(), eventCount)
}

func killStartThisActivity(act *Activity, haveCrash bool) bool {
	//dump coverage
	traceChan <- trace.TRACE_DUMP
	<-traceBackChan

	android.KillApp(config.GetPackageName())
	time.Sleep(time.Millisecond * 500)

	return startThisActivity(act, haveCrash)
}

//Start an activity
func startThisActivity(act *Activity, haveCrash bool) bool {
	//kill this app started last time
	//	if config.GetClearData() {
	//		android.ClearApp(config.GetPackageName())
	//	} else {
	//		android.KillApp(config.GetPackageName())
	//	}
	if haveCrash && len(act.GetParent()) > 0 {
		return startThisActivityFromParent(act.GetParent(), act.GetName())
	} else {
		return startThisActivityDirectly(act.Get())
	}
}

func startThisActivityDirectly(name, intent string) bool {
	//reset the key
	setKey(FALSE, name, intent, "")
	//launch app
	android.LaunchApp(config.GetPackageName(), config.GetMainActivity())
	//sendCommandToApe(APE_LAUNCH + " " + config.GetPackageName() + " " + config.GetMainActivity())

	time.Sleep(time.Millisecond * 1000)
	ok := currentActIsRight(name, MAX_TRY)
	//set the key
	setKey(TRUE, name, intent, "")
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
	pname, pin := parentTest.Act.Get()
	setKey(TRUE, pname, pin, me)
	//replay this sequence
	for _, ai := range seq.sequence {
		action := parentTest.ActSet.queue[ai]
		sendCommandToApe(action.content)
		time.Sleep(1 * time.Second)
		find := currentActIsRight(me, 1)
		if find {
			break
		}
	}

	ok = currentActIsRight(me, MAX_TRY)
	//set the key
	setKey(TRUE, me, intent, "")
	return ok
}

//If this current focused activity is right
func currentActIsRight(name string, try int) bool {
	cn := android.GetCurrentActivity(name)
	count := 0

	for !(strings.Contains(name, cn) || strings.Contains(cn, name)) {
		count++
		if count > try {
			break
		}
		time.Sleep(time.Millisecond * 1000)
		cn = android.GetCurrentActivity(name)
	}
	return (strings.Contains(name, cn) || strings.Contains(cn, name))
}

//Set the key of guider
func setKey(block, target, intent, child string) {
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

	if len(child) <= 0 {
		keys += CHILD_KEY + "\n"
	} else {
		keys += CHILD_KEY + "@" + child + "\n"
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

	if gLogCache == nil {
		log.Fatalln("gLogCache is nil.")
	}

	read, err := android.StartLogcat()
	util.FatalCheck(err)

	isLock := false
	deadTime := time.Now()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			if isLock && time.Now().After(deadTime) {
				isLock = false
				gLogCache.alock.Unlock()
			}
		}
	}()

	for {
		content, _, err := read.ReadLine()
		if err != nil {
			log.Println("Observer", err)
			break
		}
		if len(content) > 0 {
			line := string(content)
			iterms := strings.Split(string(line), "@")
			if len(iterms) >= 2 && (iterms[1] == LOG_START || iterms[1] == LOG_FINISH || iterms[1] == LOG_CHANGE) {
				gLogCache.addR(line)
			} else if len(iterms) >= 3 && iterms[1] == LOG_ACTION {
				if !isLock {
					gLogCache.alock.Lock()
					deadTime = time.Now().Add(MAX_TRY * time.Second)
					isLock = true
				}

				gLogCache.addA(line)

				if iterms[2] == LOG_ACTION_END {
					isLock = false
					gLogCache.alock.Unlock()
				}
			} else if len(iterms) >= 4 && iterms[1] == LOG_SIZE {
				gX, err = strconv.Atoi(iterms[2])
				if err != nil {
					log.Println("X and Y", err)
				}
				gY, err = strconv.Atoi(iterms[3])
				if err != nil {
					log.Println("X and Y", err)
				}
				log.Println("X and Y:", gX, gY)
			} else if len(iterms) >= 3 && iterms[1] == LOG_MINITRACE && iterms[2] == LOG_SUCCEED {
				traceChan <- trace.TRACE_PULL
				log.Println("Dump is finished. Start to Pull....")
			}

		}
	}
}

func startCrashReader() {
	if gLogCache == nil {
		log.Fatalln("gLogCache is nil.")
	}

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
				gLogCache.addC("@" + line + "@")
				crashStart = true
			} else if crashStart {
				if line == LOG_CRASH_END {
					gLogCache.addC("@" + line + "@")
					crashStart = false
				} else {
					gLogCache.addC(line)
				}
			}
		}
	}
}

//Adjust reward of this action
func Reward(set *ActionSet, index, index2 int, result Result, parent string, time int64) int {
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
			//Reward
			set.AdjustQ(index, index2, 2)
			feedback = 1
		} else {
			//It is a old activity
			set.AdjustQ(index, index2, 0)
		}
	case R_FINISH:
		//I don't want to finish
		set.AdjustQ(index, index2, -1)
	case R_NOCHANGE:
		set.AdjustQ(index, index2, 0)
	case R_CRASH:
		set.AdjustQ(index, index2, 1)
		feedback = 1
	case R_CHANGE:
		set.AdjustQ(index, index2, 0)
	default:
		log.Fatalln("Result is unknown, err")
	}
	return feedback
}
