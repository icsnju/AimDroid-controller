package test

const (
	R_NOCHANGE = iota
	R_CHANGE   = iota
	R_FINISH   = iota
	R_ACTIVITY = iota
	R_ERR      = iota
)

const (
	LOG_START  = "start"
	LOG_FINISH = "finish"
	LOG_ACTION = "action"
	LOG_SIZE   = "size"

	LEVEL_ZERO   = 0
	LEVEL_FINISH = 1
	LEVEL_START  = 2
)

type Result interface {
	GetKind() int
}

type CommonResult struct {
	kind int
}

func (cr CommonResult) GetKind() int {
	return cr.kind
}

func (this *CommonResult) SetKind(k int) {
	this.kind = k
}

type ActivityResult struct {
	CommonResult
	name   string
	intent string
}

func (this *ActivityResult) GetContent() (string, string) {
	return this.name, this.intent
}

type ErrResult struct {
	CommonResult
	content string
}

func (this *ErrResult) GetContent() string {
	return this.content
}
