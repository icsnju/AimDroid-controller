package test

import (
	"monidroid/util"
	"os"
	"path"
	"strconv"
)

const (
	R_NOCHANGE = iota
	R_CHANGE   = iota
	R_FINISH   = iota
	R_ACTIVITY = iota
	R_CRASH    = iota
)

const (
	LOG_CRASH      = "crash"
	LOG_CRASH_END  = "crashend"
	LOG_START      = "start"
	LOG_FINISH     = "finish"
	LOG_CHANGE     = "change"
	LOG_ACTION     = "action"
	LOG_ACTION_END = "end"
	LOG_SIZE       = "size"
)

var crashIndex int = 0

type Result interface {
	GetKind() int
	SetKind(int)
	ToString() string
}

type CommonResult struct {
	kind int
}

func (this *CommonResult) GetKind() int {
	return this.kind
}

func (this *CommonResult) ToString() string {
	return "common result"
}

func (this *CommonResult) SetKind(k int) {
	this.kind = k
}

type ActivityResult struct {
	CommonResult
	name   string
	intent string
}

func (this *ActivityResult) ToString() string {
	return "activity@" + this.name + "@" + this.intent + "@"
}

func (this *ActivityResult) GetContent() (string, string) {
	return this.name, this.intent
}

type CrashResult struct {
	CommonResult
	content string
	index   int
}

func NewCrashResult() *CrashResult {
	c := new(CrashResult)
	c.kind = R_CRASH
	c.content = ""
	c.index = crashIndex
	crashIndex++
	return c
}

func (this *CrashResult) GetContent() string {
	return this.content
}

func (this *CrashResult) ToString() string {
	return "crash@" + strconv.Itoa(this.index) + "@"
}

func (this *CrashResult) AddLine(line string) {
	this.content += line + "\n"
}

func (this *CrashResult) Save(out string) {
	if _, err := os.Stat(out); os.IsNotExist(err) {
		os.MkdirAll(out, os.ModePerm)
	}

	crashFile := path.Join(out, strconv.Itoa(this.index)+".txt")
	fs, err := os.OpenFile(crashFile, os.O_CREATE|os.O_RDWR, os.ModePerm)
	util.FatalCheck(err)
	fs.WriteString(this.content)
	fs.Close()
}
