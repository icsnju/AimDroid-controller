package test

import (
	"monidroid/util"
	"os"
	"path"
	"strconv"
	"sync"
)

type Activity struct {
	name     string
	intent   string
	parent   string
	findTime int64
}

//Set the Activity
func (this *Activity) Set(n, i, p string, time int64) {
	this.name = n
	this.intent = i
	this.parent = p
	this.findTime = time
}

//Get the Activity
func (this *Activity) Get() (string, string) {
	return this.name, this.intent
}

//Get the Activity name
func (this *Activity) GetName() string {
	return this.name
}

//Get the Activity name
func (this *Activity) GetParent() string {
	return this.parent
}

//Activity Queue
type ActivityQueue struct {
	queue    []*Activity
	oldQueue []*Activity
	set      map[string]*Test
	crashSet map[string]int
	lock     *sync.Mutex
}

func NewQueue() *ActivityQueue {
	return &ActivityQueue{make([]*Activity, 0), make([]*Activity, 0), make(map[string]*Test), make(map[string]int), new(sync.Mutex)}
}

func (this *ActivityQueue) Enqueue(name, intent, parent string, time int64) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	_, ex := this.set[name]
	if !ex {
		a := &Activity{}
		a.Set(name, intent, parent, time)
		this.queue = append(this.queue, a)
		test := NewTest()
		test.Act = a
		this.set[name] = test
	}
	return !ex
}

func (this *ActivityQueue) Dequeue() *Activity {
	this.lock.Lock()
	defer this.lock.Unlock()

	if len(this.queue) <= 0 {
		return nil
	}
	first := this.queue[0]
	this.queue = this.queue[1:]
	return first
}

func (this *ActivityQueue) EnOldQueue(act *Activity) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.oldQueue = append(this.oldQueue, act)
}

func (this *ActivityQueue) DeOldQueue() *Activity {
	this.lock.Lock()
	defer this.lock.Unlock()
	if len(this.oldQueue) <= 0 {
		return nil
	}
	first := this.oldQueue[0]
	this.oldQueue = this.oldQueue[1:]
	return first
}

func (this *ActivityQueue) GetTest(name string) *Test {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.set[name]
}

func (this *ActivityQueue) IsEmpty() bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	if len(this.queue) == 0 {
		return true
	}
	return false
}

func (this *ActivityQueue) ToString() string {
	this.lock.Lock()
	defer this.lock.Unlock()
	result := "Activities count: "
	l := len(this.set)
	result += strconv.Itoa(l) + "\nActivity names:\n"
	for name, _ := range this.set {
		result += name + "\n"
	}
	return result
}

func (this *ActivityQueue) AddCrash(content string, seqindex int) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	_, ex := this.crashSet[content]
	if !ex {
		this.crashSet[content] = seqindex
	}
	return ex
}

//Save queue in file
func (this *ActivityQueue) Save(out string, eventCount int) {
	if _, err := os.Stat(out); os.IsNotExist(err) {
		os.MkdirAll(out, os.ModePerm)
	}

	queueFile := path.Join(out, "queue.txt")
	fs, err := os.OpenFile(queueFile, os.O_CREATE|os.O_RDWR, os.ModePerm)
	util.FatalCheck(err)
	fs.WriteString("Send events: " + strconv.Itoa(eventCount) + "\n")
	fs.WriteString("Find activities " + strconv.Itoa(len(this.set)) + ":\n")
	for act, test := range this.set {
		fs.WriteString(act + "\t" + strconv.FormatInt(test.Act.findTime, 10) + "\n")
	}
	fs.Close()

	crashFile := path.Join(out, "crash.txt")
	fs, err = os.OpenFile(crashFile, os.O_CREATE|os.O_RDWR, os.ModePerm)
	util.FatalCheck(err)
	for content, seqindex := range this.crashSet {
		fs.WriteString(content + "\t" + strconv.Itoa(seqindex) + "\n")
	}
	fs.Close()

	for _, test := range this.set {
		if test != nil {
			test.Save(out)
		}
	}
}

//Edge between two activities
type AAEdge struct {
	ToActivity string
	StepLen    int
	SeqIndex   int
}
