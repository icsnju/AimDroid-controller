package android

import (
	"bufio"
	"monidroid/util"
	"os"
	"path"
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
func StartLogcat(sdk string) (*bufio.Reader, error) {

	_, err := util.ExeCmd(GetADBPath(sdk) + " logcat -c")
	if err != nil {
		return nil, err
	}

	cmd := util.CreateCmd(GetADBPath(sdk) + " logcat Monitor_Log:V *:E")

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
	cmd := GetADBPath(sdk) + " shell monkey --pct-touch 100 --throttle 300 -v 500"
	//cmd := GetADBPath(sdk) + " shell monkey --pct-touch 80 --pct-trackball 20 --throttle 300 --uiautomator -v 1000"
	//cmd := GetADBPath(sdk) + " shell monkey --throttle 300 --uiautomator-dfs -v 100"

	out, err := util.ExeCmd(cmd)
	return out, err
	//time.Sleep(time.Millisecond * 10000)
	//return "", nil
}

func StartApe(sdk, port string) {
	cmd := GetADBPath(sdk) + " shell monkey --port " + port
	_, err := util.ExeCmd(cmd)
	util.FatalCheck(err)
}

//adb forward
func Forward(sdk, pcPort, mobilePort string) error {
	cmd := GetADBPath(sdk) + " forward tcp:" + pcPort + " tcp:" + mobilePort
	_, err := util.ExeCmd(cmd)
	return err
}
