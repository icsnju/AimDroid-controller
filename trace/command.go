package trace

import (
	"log"
	"os/exec"
	"path"
	"strings"
)

var SDK_PATH string = "/Users/Tianchi/Tool/sdk"
var adb string = path.Join(SDK_PATH, "platform-tools/adb")
var XML_PATH string = "packages.xml"

func InitADB(sdk string) {
	SDK_PATH = sdk
	adb = path.Join(sdk, "platform-tools/adb")
}

func execmd(cmd string) string {
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	out, err := exec.Command(head, parts...).Output()
	if err != nil {
		log.Fatalln(err)
	} else if len(out) > 0 {
		//log.Println(string(out))
		return string(out)
	}
	return ""
}
