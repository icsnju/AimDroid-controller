package util

import (
	"log"
	"os/exec"
	"strings"
)

//Execute command line
func ExeCmd(cmd string) (string, error) {
	//log.Println("command is ", cmd)
	// splitting head => g++ parts => rest of the command
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	out, err := exec.Command(head, parts...).Output()
	return string(out), err
}

//Execute cmdline
func CreateCmd(cmd string) *exec.Cmd {
	log.Println("command is ", cmd)
	// splitting head => g++ parts => rest of the command
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	command := exec.Command(head, parts...)
	return command
}

func FatalCheck(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
