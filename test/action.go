package test

import (
	"log"
	"math/rand"
	"monidroid/config"
	"time"
)

type Action struct {
	content string
	reward  float64
	count   int
}

//Create an action
func NewAction(content string) *Action {
	return &Action{content, 1, 1}
}

func (this *Action) adjustReward(r float64, c int) {
	if c < 0 {
		c = 0
	}
	this.count = this.count + c
	this.reward += r
}

func (this *Action) getAveReward() float64 {
	return this.reward / float64(this.count)
}

func (this *Action) getContent() string {
	return this.content
}

//Action set
type ActionSet struct {
	queue []*Action
	set   map[string]int
}

func NewActionSet() *ActionSet {
	as := new(ActionSet)
	as.queue = make([]*Action, 0)
	as.set = make(map[string]int)
	return as
}

//Get the count of actions
func (this *ActionSet) GetCount() int {
	return len(this.queue)
}

//Add an action in set.
//If this is an new action, return true. Otherwise, return false.
func (this *ActionSet) AddAction(action *Action) bool {
	_, exist := this.set[action.content]
	if exist {
		return false
	}

	this.set[action.content] = len(this.queue)
	this.queue = append(this.queue, action)
	return true
}

//Adjust reward of an action
func (this *ActionSet) AdjustReward(index int, reward float64, count int) {
	index = (index + len(this.queue)) % len(this.queue)
	action := this.queue[index]
	action.adjustReward(reward, count)
}

//Get the action with the maximal reward
func (this *ActionSet) GetMaxRewardAction() (*Action, int) {
	if len(this.queue) <= 0 {
		log.Println("This action set has no action!")
		return nil, 0
	}

	//find the candidates
	var indexSet []int = make([]int, 0)
	var maxReward float64 = this.queue[0].getAveReward()
	for index, action := range this.queue {
		re := action.getAveReward()
		if re == maxReward {
			indexSet = append(indexSet, index)
		} else if re > maxReward {
			indexSet = make([]int, 0)
			indexSet = append(indexSet, index)
			maxReward = re
		}
	}

	//select an action from candidates
	if len(indexSet) <= 0 {
		log.Println("This action set has no action!")
		return nil, 0
	}

	if len(indexSet) == 1 {
		action := this.queue[indexSet[0]]
		return action, indexSet[0]
	}

	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(len(indexSet))
	index := indexSet[i]
	return this.queue[index], index
}

//Get an action from the set randomly
func (this *ActionSet) GetRandomAction() (*Action, int) {
	l := len(this.queue)
	if l <= 0 {
		log.Println("This action set has no action!")
		return nil, 0
	}

	if l == 1 {
		return this.queue[0], 0
	}

	index := rand.Intn(l)
	return this.queue[index], index
}

//Get an action based on epsilon-greedy algorithm
func (this *ActionSet) GetEpGreAction() (*Action, int) {
	x := rand.Float32()
	if x < config.GetEpsilon() {
		return this.GetRandomAction()
	} else {
		return this.GetMaxRewardAction()
	}
}

type ActionSequence struct {
	sequence []int
	tag      map[int]Result
	count    int
}

func NewActionSequence() *ActionSequence {
	return &ActionSequence{make([]int, 0), make(map[int]Result), 0}
}

func (this *ActionSequence) add(index int, result Result) {
	this.sequence = append(this.sequence, index)
	kind := result.GetKind()

	if kind >= R_ACTIVITY {
		this.tag[this.count] = result
	}
	this.count++
}

func (this *ActionSequence) getCount() int {
	return len(this.sequence)
}
