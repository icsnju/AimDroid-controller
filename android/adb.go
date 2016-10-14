package android

import (
	"bufio"
	"log"
	"monidroid/util"
	"os"
	"path"
	"strings"
)

var adb string = "adb"

//Launch an application
func LaunchApp(pck, act string) error {
	cmd := adb + " shell am start -n " + pck + "/" + act
	_, err := util.ExeCmd(cmd)
	return err
}

//Kill an application
func KillApp(pck string) error {
	cmd := adb + " shell am force-stop " + pck
	_, err := util.ExeCmd(cmd)
	return err
}

func ClearApp(pck string) error {
	cmd := adb + " shell pm clear " + pck
	_, err := util.ExeCmd(cmd)
	return err
}

func InitADB(sdk string) {
	adb = path.Join(sdk, "platform-tools/adb")
}

//start logcat
func StartLogcat() (*bufio.Reader, error) {

	_, err := util.ExeCmd(adb + " logcat -c")
	if err != nil {
		return nil, err
	}

	cmd := util.CreateCmd(adb + " logcat Monitor_Log:V *:S")

	// Create stdout, stderr streams of type io.Reader
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	// Start command
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	read := bufio.NewReader(stdout)
	return read, nil
}

//push file to the device
func PushFile(src, dst string) error {
	if _, err := os.Stat(src); err == nil {
		cmd := adb + " push " + src + " " + dst
		_, err = util.ExeCmd(cmd)
		return err
	} else {
		return err
	}
}

//remove file from the device
func RemoveFile(dst string) error {
	cmd := adb + " shell rm " + dst
	_, err := util.ExeCmd(cmd)
	return err
}

func StartMonkey(pkg string) (string, error) {
	cmd := adb + " shell monkey -p " + pkg + " --throttle 500 -v 1000"
	//cmd := GetADBPath(sdk) + " shell monkey --pct-touch 80 --pct-trackball 20 --throttle 300 --uiautomator -v 1000"
	//cmd := GetADBPath(sdk) + " shell monkey --throttle 300 --uiautomator-dfs -v 100"

	out, err := util.ExeCmd(cmd)
	return out, err
	//time.Sleep(time.Millisecond * 10000)
	//return "", nil
}

func StartApe(port string) error {
	cmd := adb + " shell monkey --ignore-crashes --port " + port
	out, err := util.ExeCmd(cmd)
	log.Println(out)
	return err
}

//adb forward
func Forward(pcPort, mobilePort string) error {
	cmd := adb + " forward tcp:" + pcPort + " tcp:" + mobilePort
	_, err := util.ExeCmd(cmd)
	return err
}

//get current focused activity
func GetCurrentActivity() string {
	name := ""
	cmd := adb + " shell dumpsys activity activities | grep mFocusedActivity"
	out, err := util.ExeCmd(cmd)
	if err != nil {
		log.Println("Cannot get current activity!", err)
		return name
	}
	iterms := strings.Split(out, " ")
	for i, iterm := range iterms {
		if iterm == "u0" {
			i++
			if i < len(iterms) {
				name = iterms[i]
			}
			break
		}
	}
	//log.Println(name)
	if len(name) > 0 {
		names := strings.Split(name, "/")
		if len(names) == 2 {
			name = names[1]
		}
	}
	return name
}
