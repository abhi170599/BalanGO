/*
 Configuration file parser
*/

package ConfigParser

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Config struct {
	Instances Servers `xml:"Servers"`
}

type Server struct {
	Url string `xml:"address,attr"`
}

type Servers struct {
	Instances []Server `xml:"Server"`
	Port      int      `xml:"Port"`
	Mode      string   `xml:"Mode"`
}

/*
function to parse the file
*/
func ParseFile(filePath string) (string, string, int, error) {

	//open xml config file
	xmlFile, err := os.Open(filePath)
	if err != nil {
		return "", "", 0, err
	}
	fmt.Println("Reading Configuration...")
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	//initialize the variables
	var config Config

	xml.Unmarshal(byteValue, &config)

	serverList := make([]string, 0)
	for _, s := range config.Instances.Instances {
		serverList = append(serverList, s.Url)
	}
	servers := strings.Join(serverList, ",")

	return servers, config.Instances.Mode, config.Instances.Port, nil
}
