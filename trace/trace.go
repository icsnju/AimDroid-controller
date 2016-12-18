package trace

import (
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	TRACE_STOP = 0
	TRACE_DUMP = 1
	TRACE_PULL = 2
)

var PREFIX string = "mini_trace_"
var POSTFIX []string = []string{"_config.in", "_data.bin", "_coverage.dat", "_info.log", "package.txt"}

//Push configure file to the device
func PushConfig(name string) {
	uidStr := IdGetter(name)
	uid, err := strconv.Atoi(uidStr)
	if err != nil {
		log.Fatalln(err)
	}
	uidShortStr := strconv.Itoa(uid - 10000)
	targetUser := "u0_a" + uidShortStr
	targetGroup := targetUser

	pre := PREFIX + uidStr

	for i, post := range POSTFIX {
		var targetConfigFileName string = ""
		if i == (len(POSTFIX) - 1) {
			targetConfigFileName = post
		} else {
			targetConfigFileName = pre + post
		}
		// detect if file exists
		var _, err = os.Stat(targetConfigFileName)

		var file *os.File
		// create file if not exists
		if os.IsNotExist(err) {
			file, err = os.Create(targetConfigFileName)
			if err != nil {
				log.Fatalln(err)
			}

		}

		if i == (len(POSTFIX) - 1) {
			file.WriteString(name + "\n")
			file.Sync()
		}
		file.Close()

		// push file to device
		cmd := adb + " push " + targetConfigFileName + " /sdcard/"
		execmd(cmd)
		cmd = adb + " shell su -c mv /sdcard/" + targetConfigFileName + " /data/"
		execmd(cmd)
		cmd = adb + " shell su -c chown " + targetUser + ":" + targetGroup + " /data/" + targetConfigFileName
		execmd(cmd)
		err = os.Remove(targetConfigFileName)
		if err != nil {
			log.Fatalln(err)
		}
	}

	cmd := adb + " shell su -c setenforce 0"
	execmd(cmd)
	//delete old files
	cmd = adb + " shell rm -rf /sdcard/coverage"
	execmd(cmd)
	cmd = adb + " shell mkdir /sdcard/coverage"
	execmd(cmd)
	log.Println("Pushing configure file is completed.")
}

//Pull coverage file from the device
func CopyCoverage(name string, index string) {
	uidStr := IdGetter(name)
	fileName := PREFIX + uidStr + POSTFIX[2]
	cmd := adb + " shell su -c cp /data/" + fileName + " /sdcard/coverage/" + index + "_" + fileName
	execmd(cmd)
	log.Println("Copy coverage file is completed.")
}

func DownloadCoverage(name string) {
	cmd := adb + " pull /sdcard/coverage ."
	execmd(cmd)

	outpath := path.Join("out", name)
	//mv coverage file to out file
	if _, err := os.Stat(outpath); os.IsNotExist(err) {
		os.MkdirAll(outpath, os.ModePerm)
	}
	cmd = "mv -f coverage " + outpath
	execmd(cmd)
	cmd = adb + " shell rm -rf /sdcard/coverage"
	execmd(cmd)

	log.Println("Downloading coverage file is completed.")
}

//Send signal to the device and let it dump the coverage
func DumpCoverage(name string, start time.Time, goon chan int) bool {
	cmd := adb + " shell ps | grep " + name
	out := execmd(cmd)
	log.Println("Start to dump the coverage file in the device.")
	lines := strings.Split(out, "\n")
	for index, line := range lines {
		if len(line) > 0 {
			iterms := strings.Fields(line)
			if len(iterms) >= 9 {
				pid := iterms[1]
				cmd = adb + " shell su -c kill -USR2 " + pid
				execmd(cmd)
				log.Println("Start to dump the coverage file of process:", pid)
				dumpok := <-goon
				if dumpok == TRACE_PULL {
					dur := int(time.Now().Sub(start).Seconds()) + index
					durI := strconv.Itoa(int(dur))
					log.Println("Start to copy the coverage file in time:", durI)
					CopyCoverage(name, durI)
				}
			}
		}
	}

	return false
}

func StartTrace(pckname string, start time.Time, goon, goback chan int) {

	loop := true
	for loop {
		//5 minutes timer

		select {
		case v := <-goon:
			if v == TRACE_STOP {
				loop = false
			} else if v == TRACE_DUMP {
				//dump and copy
				DumpCoverage(pckname, start, goon)
				goback <- TRACE_DUMP
			} else if v == TRACE_PULL {
				dur := time.Now().Sub(start).Seconds()
				durI := strconv.Itoa(int(dur))
				CopyCoverage(pckname, durI)
			}
		}
	}

	DownloadCoverage(pckname)
	goback <- TRACE_STOP
	log.Println("Minitracing is stopping...")
}
