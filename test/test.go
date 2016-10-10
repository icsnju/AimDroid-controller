package test

import (
	"log"
	"monidroid/util"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

type Test struct {
	Act           *Activity
	ActSet        *ActionSet
	SequenceArray []*ActionSequence
	Cache         *LogCache
}

func NewTest() *Test {
	t := new(Test)
	t.Act = nil
	t.ActSet = NewActionSet()
	t.Cache = NewLogCache()
	t.SequenceArray = make([]*ActionSequence, 0)
	return t
}

//Save test into a file
func (this *Test) Save(out string) {
	mDir := path.Join(out, this.Act.GetName())
	if _, err := os.Stat(mDir); os.IsNotExist(err) {
		os.MkdirAll(mDir, os.ModeDir)
	}

	//save activity
	actFile := path.Join(mDir, "activity.txt")
	fs, err := os.OpenFile(actFile, os.O_CREATE|os.O_RDWR, os.ModePerm)
	util.FatalCheck(err)

	name, intent := this.Act.Get()
	fs.WriteString(name + "\n")
	fs.WriteString(intent + "\n")
	fs.Close()

	//save actions
	actionFile := path.Join(mDir, "actions.txt")
	fs, err = os.OpenFile(actionFile, os.O_CREATE|os.O_RDWR, os.ModePerm)
	util.FatalCheck(err)
	queue := this.ActSet.queue
	for _, action := range queue {
		fs.WriteString(action.getContent() + "\n")
	}
	fs.Close()

	seqsDir := path.Join(mDir, "Sequences")
	if _, err := os.Stat(seqsDir); os.IsNotExist(err) {
		os.MkdirAll(mDir, os.ModeDir)
	}

	//save sequences
	for i, seq := range this.SequenceArray {
		seqFile := path.Join(seqsDir, strconv.Itoa(i)+".txt")
		fs, err = os.OpenFile(seqFile, os.O_CREATE|os.O_RDWR, os.ModePerm)
		util.FatalCheck(err)
		for j, ai := range seq.sequence {
			action := queue[ai]
			fs.WriteString(action.getContent())
			rs, ex := seq.tag[j]
			if ex {
				fs.WriteString(" " + rs.ToString() + "\n")
			}
		}
		fs.Close()
	}

}

//Put device log into this cache
type LogCache struct {
	logs []string
	lock *sync.Mutex
}

func NewLogCache() *LogCache {
	return &LogCache{make([]string, 0), new(sync.Mutex)}
}

func (this *LogCache) add(line string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.logs = append(this.logs, line)
}

func (this *LogCache) clear() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.logs = make([]string, 0)
}

//Filter out actions from the logs
func (this *LogCache) filterAction(set *ActionSet) int {
	count := 0
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, line := range this.logs {
		iterms := strings.Split(string(line), "@")
		if len(iterms) >= 3 && iterms[1] == LOG_ACTION {
			a := NewAction(iterms[2])
			ok := set.AddAction(a)
			if ok {
				count++
			}
		}
	}
	//clear the log cache
	this.logs = make([]string, 0)
	return count
}

//Filter out size from the logs
func (this *LogCache) filterSize() (int, int) {
	this.lock.Lock()
	defer this.lock.Unlock()
	x := 0
	y := 0
	for _, line := range this.logs {
		iterms := strings.Split(string(line), "@")
		if len(iterms) >= 4 && iterms[1] == LOG_SIZE {
			xs := iterms[2]
			ys := iterms[3]
			var err error
			x, err = strconv.Atoi(xs)
			if err != nil {
				log.Println("Size x err:", err)
				continue
			}
			y, err = strconv.Atoi(ys)
			if err != nil {
				log.Println("Size y err:", err)
				continue
			}
			break
		}
	}
	//clear the log cache
	this.logs = make([]string, 0)
	return x, y
}

//Filter out the results from the logs
func (this *LogCache) filterResult() Result {
	this.lock.Lock()
	defer this.lock.Unlock()
	var result Result
	level := LEVEL_ZERO

	for _, line := range this.logs {
		iterms := strings.Split(string(line), "@")
		if len(iterms) >= 2 {
			switch iterms[1] {
			case LOG_START:
				if len(iterms) >= 4 && LEVEL_START >= level {
					result = &ActivityResult{CommonResult{R_ACTIVITY}, iterms[2], iterms[3]}
					level = LEVEL_START
				}
			case LOG_FINISH:
				if LEVEL_FINISH >= level {
					result = &CommonResult{R_FINISH}
					level = LEVEL_FINISH
				}
			default:
				log.Println("Unknown result:", line)
			}
		} else {
			log.Println("Unknown result:", line)
		}
	}
	//clear the log cache
	this.logs = make([]string, 0)
	if level == LEVEL_ZERO {
		result = &CommonResult{R_NOCHANGE}
	}
	return result
}
