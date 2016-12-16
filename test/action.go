package test

import (
	"log"
	"math/rand"
	"monidroid/config"
	"strconv"
	"strings"
	"time"
)

type Action struct {
	content string
	Q       float64
	usable  bool
}

//Create an action
func NewAction(content string) *Action {
	return &Action{content, 1, true}
}

func (this *Action) adjustQ(q float64) {
	this.Q = q
}

func (this *Action) getQ() float64 {
	return this.Q
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
func (this *ActionSet) GetAction(index int) *Action {
	if len(this.queue) <= index {
		return nil
	}
	return this.queue[index]
}

//Add an action in set.
//If this is an new action, return true. Otherwise, return false.
func (this *ActionSet) AddAction(action *Action) bool {
	content := action.getContent()
	iterms := strings.Fields(content)
	if len(iterms) >= 3 && iterms[0] == "tap" {
		x, _ := strconv.Atoi(iterms[1])
		y, _ := strconv.Atoi(iterms[2])
		if x > gX || y > gY {
			return false
		}
	}
	index, exist := this.set[action.content]
	if exist {
		me := this.queue[index]
		me.usable = true
		return false
	}

	this.set[action.content] = len(this.queue)
	this.queue = append(this.queue, action)
	return true
}

//Adjust reward of an action with SARSA
func (this *ActionSet) AdjustQ(index, index2 int, reward float64) {
	index = (index + len(this.queue)) % len(this.queue)
	index2 = (index2 + len(this.queue)) % len(this.queue)

	action := this.queue[index]
	action2 := this.queue[index2]
	action.Q = action.Q + config.GetAlpha()*(reward+config.GetGamma()*action2.Q-action.Q)
}

//Get the action with the maximal reward
func (this *ActionSet) GetMaxQAction() (*Action, int) {
	if len(this.queue) <= 0 {
		log.Println("This action set has no action!")
		return nil, 0
	}

	//find the candidates
	var indexSet []int = make([]int, 0)
	var maxReward float64 = 0
	firstTime := true
	for index, action := range this.queue {
		if action.usable {
			re := action.getQ()
			if firstTime {
				firstTime = false
				indexSet = append(indexSet, index)
				maxReward = re
			} else if re == maxReward {
				indexSet = append(indexSet, index)
			} else if re > maxReward {
				indexSet = make([]int, 0)
				indexSet = append(indexSet, index)
				maxReward = re
			}
		}
		action.usable = false
	}

	//select an action from candidates
	if len(indexSet) <= 0 {
		for index, action := range this.queue {
			re := action.getQ()
			if re == maxReward {
				indexSet = append(indexSet, index)
			} else if re > maxReward {
				indexSet = make([]int, 0)
				indexSet = append(indexSet, index)
				maxReward = re
			}
		}
	}

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
		this.queue[0].usable = false
		return this.queue[0], 0
	}

	index := rand.Intn(l)

	return this.queue[index], index
}

//Get an action based on epsilon-greedy algorithm
func (this *ActionSet) GetEpGreAction() (*Action, int) {
	x := rand.Float64()
	if x < config.GetEpsilon() {
		return this.GetRandomAction()
	} else {
		return this.GetMaxQAction()
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
