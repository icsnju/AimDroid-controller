package test

import (
	"log"
	"minitrace/trace"
	"strconv"
	"time"
)

const (
	TRACE_STOP = 0
	TRACE_DUMP = 1
	TRACE_PULL = 2
)

//var index int = 1

//func StartTrace(pckname string, start time.Time, stop chan int) {

//	loop := true
//	for loop {
//		//5 minutes timer
//		now := time.Now()

//		select {
//		case <-stop:
//			loop = false
//		case <-time.After((start.Add(time.Duration(index) * 5 * time.Minute)).Sub(now)):
//		}
//		trace.DumpCoverage(pckname)
//		if loop {
//			time.Sleep(2 * time.Minute)
//		} else {
//			time.Sleep(20 * time.Second)
//		}
//		trace.PullCoverage(pckname, index)
//		index++
//	}
//	stop <- 0
//}

func StartTrace(pckname string, start time.Time, goon, goback chan int) {

	loop := true
	needback := false
	for loop {
		//5 minutes timer

		select {
		case v := <-goon:
			if v == TRACE_STOP {
				loop = false
			} else if v == TRACE_DUMP {
				if trace.DumpCoverage(pckname) {
					needback = true
				} else {
					goback <- TRACE_DUMP
				}
			} else if v == TRACE_PULL {
				if needback {
					goback <- TRACE_PULL
					needback = false
				}
				dur := time.Now().Sub(start).Seconds()
				durI := strconv.Itoa(int(dur))
				trace.CopyCoverage(pckname, durI)
			}
		}
	}

	trace.DownloadCoverage(pckname)
	goback <- TRACE_PULL
	log.Println("Minitracing is stopping...")
}
