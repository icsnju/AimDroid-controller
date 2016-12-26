package trace

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"
)

type Packages struct {
	XMLName xml.Name  `xml:"packages"`
	Pcks    []Package `xml:"package"`
}

type Package struct {
	XMLName      xml.Name `xml:"package"`
	Name         string   `xml:"name,attr"`
	UserId       string   `xml:"userId,attr"`
	SharedUserId string   `xml:"sharedUserId,attr"`
}

//func main() {
//	args := os.Args
//	if len(args) <= 1 {
//		XmlPuller()
//		return
//	}

//	name := args[1]
//	IdGetter(name)
//}

var id string = ""

func XmlPuller() {
	//copy packages.xml file to sdcard
	cmd := adb + " shell su -c cp /data/system/packages.xml /sdcard/"
	execmd(cmd)
	//pull packages.xml to local
	cmd = adb + " pull /sdcard/packages.xml ."
	execmd(cmd)
	log.Println("Pulling packages.xml is completed.")
}

func IdGetter(name string) string {
	if len(id) > 0 {
		return id
	}

	if _, err := os.Stat(XML_PATH); os.IsNotExist(err) {
		XmlPuller()
	}

	//parser the xml file
	packFile, err := os.Open("packages.xml")
	if err != nil {
		log.Fatalln(err)
	}
	defer packFile.Close()

	data, err := ioutil.ReadAll(packFile)
	if err != nil {
		log.Fatalln(err)
	}

	packages := Packages{}
	err = xml.Unmarshal(data, &packages)
	if err != nil {
		log.Fatalln(err)
	}

	for _, pck := range packages.Pcks {
		if pck.Name == name {
			id = pck.UserId
			if len(id) <= 0 {
				id = pck.SharedUserId
			}
			log.Println(pck.Name, id)
			break
		}
	}

	if len(id) <= 0 {
		log.Fatalln("Cannot find this package in packages.xml...", name)
	}
	return id
}
