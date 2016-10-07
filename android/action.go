package android

import (
	"log"
	"math/rand"
	"time"
)

type Action struct {
	content string
	reward  float32
	count   int
}

type ActionSet struct {
	queue []*Action
	set   map[string]int
}

func (this *Action) adjustReward(r float32, c int) {
	sum := float32(this.count) * this.reward
	sum += r
	this.count = this.count + c
	this.reward = sum / float32(this.count)
}

func (this *Action) getContent() string {
	return this.content
}

//Add an action in set.
//If this is an new action, return true. Otherwise, return false.
func (this *ActionSet) AddAction(action *Action) bool {
	_, exist := this.set[action.content]
	if exist {
		return false
	}

	this.set[action.content] = 1
	this.queue = append(this.queue, action)
	return true
}

//Get the action with the maximal reward
func (this *ActionSet) GetMaxRewardAction() *Action {
	if len(this.queue) <= 0 {
		log.Println("This action set has no action!")
		return nil
	}

	//find the candidates
	var indexSet []int = make([]int, 0)
	var maxReward float32 = this.queue[0].reward
	for index, action := range this.queue {
		re := action.reward
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
		return nil
	}

	if len(indexSet) == 1 {
		action := this.queue[indexSet[0]]
		return action
	}

	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(indexSet))
	return this.queue[index]
}

//Get an action from the set randomly
func (this *ActionSet) GetRandomAction() *Action {
	l := len(this.queue)
	if l <= 0 {
		log.Println("This action set has no action!")
		return nil
	}

	if l == 1 {
		return this.queue[0]
	}

	index := rand.Intn(l)
	return this.queue[index]
}
