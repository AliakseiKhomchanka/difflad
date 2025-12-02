package parser

import (
	"encoding/xml"
	"fmt"
	"os"

	elements "openplc-render/elements"
	plcxml "openplc-render/xml"
)

func Parse(filepath string) (*elements.POU, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error: could not read xml file from the specified path: %s", filepath)
	}
	var project plcxml.Project
	err = xml.Unmarshal(data, &project)
	if err != nil {
		return nil, fmt.Errorf("error: could not unmarshal XML data from: %s", filepath)
	}
	return nil, nil
}
