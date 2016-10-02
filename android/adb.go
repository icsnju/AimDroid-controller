package android

import (
	"bufio"
	"log"
	"monidroid/util"
	"os"
	"path"
	"strings"
)

const (
	LOG_START  = "start"
	LOG_CREATE = "create"
	LOG_FINISH = "finish"
)

//Launch an application
func LaunchApp(sdk, pck, act string) error {
	adb := GetADBPath(sdk)
	cmd := adb + " shell am start -n " + pck + "/" + act
	_, err := util.ExeCmd(cmd)
	return err
}

//Kill an application
func KillApp(sdk, pck string) error {
	adb := GetADBPath(sdk)
	cmd := adb + " shell am force-stop " + pck
	_, err := util.ExeCmd(cmd)
	return err
}

func GetADBPath(sdk string) string {
	return path.Join(sdk, "platform-tools/adb")
}

//start logcat
func StartLogcat(sdk string, queue *ActivityQueue) {

	_, err := util.ExeCmd(GetADBPath(sdk) + " logcat -c")
	if err != nil {
		log.Println(err)
		return
	}

	cmd := util.CreateCmd(GetADBPath(sdk) + " logcat Monitor_Log:V *:E")

	// Create stdout, stderr streams of type io.Reader
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		return
	}
	// Start command
	err = cmd.Start()
	if err != nil {
		log.Println(err)
		return
	}

	read := bufio.NewReader(stdout)
	for {
		content, _, err := read.ReadLine()
		if err != nil {
			log.Println(err)
			break
		}
		if len(content) > 0 {
			iterms := strings.Split(string(content), "#")
			if len(iterms) >= 2 {
				switch iterms[1] {
				case LOG_START:
					if len(iterms) >= 4 {
						queue.Enqueue(iterms[2], iterms[3])
					}
				case LOG_CREATE:
					queue.SetFocusedActivity(iterms[2])
				case LOG_FINISH:
				default:
					//log.Println(content)
				}
			}
		}
	}
}

//push file to the device
func PushFile(sdk, src, dst string) error {
	if _, err := os.Stat(sdk); err == nil {
		cmd := GetADBPath(sdk) + " push " + src + " " + dst
		_, err = util.ExeCmd(cmd)
		return err
	} else {
		return err
	}
}

//remove file from the device
func RemoveFile(sdk, dst string) error {
	cmd := GetADBPath(sdk) + " shell rm " + dst
	_, err := util.ExeCmd(cmd)
	return err
}

func StartMonkey(sdk, pkg string) (string, error) {
	cmd := GetADBPath(sdk) + " shell monkey --pct-touch 80 --pct-trackball 20 --throttle 300 -v 500"
	out, err := util.ExeCmd(cmd)
	return out, err
}

//adb forward
func Forward(sdk, pcPort, mobilePort string) error {
	cmd := GetADBPath(sdk) + " forward tcp:" + pcPort + " tcp:" + mobilePort
	_, err := util.ExeCmd(cmd)
	return err
}
