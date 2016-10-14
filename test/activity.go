package test

import (
	"monidroid/util"
	"os"
	"path"
	"strconv"
	"sync"
)

type Activity struct {
	name   string
	intent string
}

//Set the Activity
func (this *Activity) Set(n, i string) {
	this.name = n
	this.intent = i
}

//Get the Activity
func (this *Activity) Get() (string, string) {
	return this.name, this.intent
}

//Get the Activity name
func (this *Activity) GetName() string {
	return this.name
}

//Activity Queue
type ActivityQueue struct {
	queue []*Activity
	set   map[string]int
	lock  *sync.Mutex
}

func NewQueue() *ActivityQueue {
	return &ActivityQueue{make([]*Activity, 0), make(map[string]int), new(sync.Mutex)}
}

func (this *ActivityQueue) Enqueue(name, intent string) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	_, ex := this.set[name]
	if !ex {
		this.set[name] = 0
		a := &Activity{}
		a.Set(name, intent)
		this.queue = append(this.queue, a)
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

func (this *ActivityQueue) AddActivityInSet(name string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.set[name] = 0
}

//Save queue in file
func (this *ActivityQueue) Save(out string) {
	if _, err := os.Stat(out); os.IsNotExist(err) {
		os.MkdirAll(out, os.ModePerm)
	}

	queueFile := path.Join(out, "queue.txt")
	fs, err := os.OpenFile(queueFile, os.O_CREATE|os.O_RDWR, os.ModePerm)
	util.FatalCheck(err)
	fs.WriteString("Find activity " + strconv.Itoa(len(this.set)) + ":\n")
	for act, _ := range this.set {
		fs.WriteString(act + "\n")
	}
	fs.Close()
}

//Edge between two activities
type AAEdge struct {
	ToActivity string
	StepLen    int
	SeqIndex   int
}
