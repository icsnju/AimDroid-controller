package android

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
