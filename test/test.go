package test

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

const MAX_TRY = 10

type Test struct {
	Act           *Activity
	ActSet        *ActionSet
	SequenceArray []*ActionSequence
	Find          map[string]*AAEdge
	HaveCrash     bool
}

func NewTest() *Test {
	t := new(Test)
	t.Act = nil
	t.ActSet = NewActionSet()
	t.SequenceArray = make([]*ActionSequence, 0)
	t.Find = make(map[string]*AAEdge)
	t.HaveCrash = false
	return t
}

//Save test into a file
func (this *Test) Save(out string) {
	mDir := path.Join(out, this.Act.GetName())
	if _, err := os.Stat(mDir); os.IsNotExist(err) {
		os.MkdirAll(mDir, os.ModePerm)
	}

	//save activity
	actFile := path.Join(mDir, "activity.txt")
	fs, err := os.OpenFile(actFile, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Fatalln("Save activity:", err)
	}

	name, intent := this.Act.Get()
	fs.WriteString(name + "\n")
	fs.WriteString(intent + "\n")
	fs.WriteString(this.Act.GetParent() + "\n")
	fs.Close()

	//save actions
	actionFile := path.Join(mDir, "actions.txt")
	fs, err = os.OpenFile(actionFile, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Fatalln("Save Actions:", err)
	}
	queue := this.ActSet.queue
	for _, action := range queue {
		fs.WriteString(action.getContent() + "\t" + strconv.FormatFloat(action.Q, 'f', 4, 64) + "\n")
	}
	fs.Close()

	//save sequences
	seqsDir := path.Join(mDir, "Sequences")
	if _, err := os.Stat(seqsDir); os.IsNotExist(err) {
		os.MkdirAll(seqsDir, os.ModePerm)
	}

	crashDir := path.Join(mDir, "Crash")
	if _, err := os.Stat(crashDir); os.IsNotExist(err) {
		os.MkdirAll(crashDir, os.ModePerm)
	}

	for i, seq := range this.SequenceArray {
		seqFile := path.Join(seqsDir, strconv.Itoa(i)+".txt")
		fs, err = os.OpenFile(seqFile, os.O_CREATE|os.O_RDWR, os.ModePerm)
		if err != nil {
			log.Fatalln("Save Sequence:", err)
		}
		for j, ai := range seq.sequence {
			action := queue[ai]
			fs.WriteString(action.getContent())
			rs, ex := seq.tag[j]
			if ex {
				fs.WriteString("@" + rs.ToString())
				cr, ok := rs.(*CrashResult)
				if ok {
					cr.Save(crashDir)
				}
			}
			fs.WriteString("\n")
		}
		fs.Close()
	}

	//save edges
	edgeFile := path.Join(mDir, "edge.txt")
	fs, err = os.OpenFile(edgeFile, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Println("Save edge:", err)
	}

	for _, e := range this.Find {
		content, err := json.Marshal(e)
		if err == nil {
			fs.Write(content)
			fs.WriteString("\n")
		} else {
			log.Println(err)
		}
	}
	fs.Close()
}

//Put device log into this cache
type LogCache struct {
	rlogs []string //result
	clogs []string //crash
	alogs []string //actions
	rlock *sync.Mutex
	clock *sync.Mutex
	alock *sync.Mutex
}

func NewLogCache() *LogCache {
	return &LogCache{make([]string, 0), make([]string, 0), make([]string, 0), new(sync.Mutex), new(sync.Mutex), new(sync.Mutex)}
}

func (this *LogCache) addR(line string) {
	this.rlock.Lock()
	defer this.rlock.Unlock()
	this.rlogs = append(this.rlogs, line)
}

func (this *LogCache) clearR() {
	this.rlock.Lock()
	defer this.rlock.Unlock()
	this.rlogs = make([]string, 0)
}

func (this *LogCache) addC(line string) {
	this.clogs = append(this.clogs, line)
}

func (this *LogCache) clearC() {
	this.clock.Lock()
	defer this.clock.Unlock()
	this.clogs = make([]string, 0)
}

func (this *LogCache) clearRC() {
	this.clearR()
	this.clearC()
}

func (this *LogCache) addA(line string) {
	this.alogs = append(this.alogs, line)
}

func (this *LogCache) clearA() {
	this.alock.Lock()
	defer this.alock.Unlock()
	this.alogs = make([]string, 0)
}

func (this *LogCache) clearAll() {
	this.clearRC()
	this.clearA()
}

//Filter out actions from the logs
func (this *LogCache) filterAction(set *ActionSet) int {
	count := 0

	this.alock.Lock()

	for _, line := range this.alogs {
		iterms := strings.Split(line, "@")
		if len(iterms) >= 3 && iterms[1] == LOG_ACTION {
			if LOG_ACTION_END == iterms[2] {
				break
			}
			a := NewAction(iterms[2])
			ok := set.AddAction(a)
			if ok {
				count++
			}
		}
	}
	//clear the log cache
	this.alogs = make([]string, 0)

	this.alock.Unlock()

	return count
}

//Filter out size from the logs
//func (this *LogCache) filterSize() (int, int) {
//	this.lock.Lock()
//	defer this.lock.Unlock()
//	x := 0
//	y := 0
//	for _, line := range this.logs {
//		iterms := strings.Split(string(line), "@")
//		if len(iterms) >= 4 && iterms[1] == LOG_SIZE {
//			xs := iterms[2]
//			ys := iterms[3]
//			var err error
//			x, err = strconv.Atoi(xs)
//			if err != nil {
//				log.Println("Size x err:", err)
//				continue
//			}
//			y, err = strconv.Atoi(ys)
//			if err != nil {
//				log.Println("Size y err:", err)
//				continue
//			}
//			break
//		}
//	}
//	//clear the log cache
//	this.logs = make([]string, 0)
//	return x, y
//}

//Filter out the results from the logs
func (this *LogCache) filterResult() Result {

	var result Result
	level := R_NOCHANGE

	//find normal results
	this.rlock.Lock()
	for _, line := range this.rlogs {
		iterms := strings.Split(string(line), "@")
		if len(iterms) >= 2 {
			switch iterms[1] {
			case LOG_START:
				if len(iterms) >= 4 && R_ACTIVITY >= level {
					result = &ActivityResult{CommonResult{R_ACTIVITY}, iterms[2], iterms[3]}
					level = R_ACTIVITY
				}
			case LOG_FINISH:
				if R_FINISH >= level {
					result = &CommonResult{R_FINISH}
					level = R_FINISH
				}
			case LOG_CHANGE:
				if R_CHANGE >= level {
					result = &CommonResult{R_CHANGE}
					level = R_CHANGE
				}
			default:
				log.Println("Unknown result:", line)
			}

		} else {
			log.Println("Unknown result:", line)
		}
	}
	//clear the rlog cache
	this.rlogs = make([]string, 0)
	this.rlock.Unlock()

	//find crash
	this.clock.Lock()
	if len(this.clogs) > 0 {
		level = R_CRASH
		var rr *CrashResult = NewCrashResult()
		result = rr
		l := len(this.clogs)
		for j := 0; j < l; j++ {
			rr.AddLine(this.clogs[j])
		}
		this.clogs = make([]string, 0)
	}
	this.clock.Unlock()

	if level == R_NOCHANGE {
		result = &CommonResult{R_NOCHANGE}
	}
	return result
}

func (this *Test) addEdge(rs Result, step int) {
	if rs.GetKind() == R_ACTIVITY {
		act, ok := rs.(*ActivityResult)
		if ok {
			name, _ := act.GetContent()
			_, ex := this.Find[name]
			if !ex {
				ne := new(AAEdge)
				ne.SeqIndex = len(this.SequenceArray)
				ne.StepLen = step
				ne.ToActivity = name
				this.Find[name] = ne
			}
		}
	}
}
