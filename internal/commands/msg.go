package commands

import (
	"encoding/xml"
	"fmt"
	"os"
)

func Msg(inXmlPath string) error {
	xmlFile, err := os.Open(inXmlPath)
	if err != nil {
		return fmt.Errorf("fatal: %s not found", inXmlPath)
	}
	defer xmlFile.Close()

	decoder := xml.NewDecoder(xmlFile)
	for {
		t, err := decoder.Token()
		if err != nil {
			break
		}
		if se, ok := t.(xml.StartElement); ok && se.Name.Local == "wit" {
			for _, attr := range se.Attr {
				if attr.Name.Local == "message" {
					fmt.Println(attr.Value)
					return nil
				}
			}
			break
		}
	}
	fmt.Println("(No snapshot message found)")
	return nil
}
