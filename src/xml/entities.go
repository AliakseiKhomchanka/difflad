// TODO: Collapse some attributes like variables of blocks for easier structure?

package openplc_xml

import (
	"encoding/xml"
	"fmt"
)

type Project struct {
	XMLName       xml.Name      `xml:"http://www.plcopen.org/xml/tc6_0201 project"`
	FileHeader    FileHeader    `xml:"fileHeader"`
	ContentHeader ContentHeader `xml:"contentHeader"`
	Types         Types         `xml:"types"`
	Instances     Instances     `xml:"instances"`
}

type FileHeader struct {
	CompanyName      string `xml:"companyName,attr"`
	ProductName      string `xml:"productName,attr"`
	ProductVersion   string `xml:"productVersion,attr"`
	CreationDateTime string `xml:"creationDateTime,attr"`
}

type ContentHeader struct {
	Name                 string         `xml:"name,attr"`
	ModificationDateTime string         `xml:"modificationDateTime,attr"`
	CoordinateInfo       CoordinateInfo `xml:"coordinateInfo"`
}

type CoordinateInfo struct {
	FBD Scaling `xml:"fbd>scaling"`
	LD  Scaling `xml:"ld>scaling"`
	SFC Scaling `xml:"sfc>scaling"`
}

type Scaling struct {
	X int `xml:"x,attr"`
	Y int `xml:"y,attr"`
}

type Types struct {
	POUs POUs `xml:"pous"`
}

type POUs struct {
	POU []POU `xml:"pou"`
}

type POU struct {
	Name      string    `xml:"name,attr"`
	POUType   string    `xml:"pouType,attr"`
	Interface Interface `xml:"interface"`
	Body      Body      `xml:"body"`
}

type Interface struct {
	LocalVars LocalVars `xml:"localVars"`
}

type LocalVars struct {
	Variable []Variable `xml:"variable"`
}

type BlockVariables struct {
	Variable []BlockVariable `xml:"variable"`
}

type Variable struct {
	Name string  `xml:"name,attr"`
	Type VarType `xml:"type"`
}

type BlockVariable struct {
	FormalParameter    string            `xml:"formalParameter,attr,omitempty"`
	ConnectionPointIn  []ConnectionPoint `xml:"connectionPointIn,omitempty"`
	ConnectionPointOut []ConnectionPoint `xml:"connectionPointOut,omitempty"`
}

type VarType struct {
	Derived Derived `xml:"derived"`
}

type Derived struct {
	Name string `xml:"name,attr"`
}

type Body struct {
	LD LD `xml:"LD"`
}

type LD struct {
	LeftPowerRail  []*Primitive `xml:"leftPowerRail"`
	Contact        []*Primitive `xml:"contact"`
	Coil           []*Primitive `xml:"coil"`
	RightPowerRail []*Primitive `xml:"rightPowerRail"`
	Connector      []*Primitive `xml:"connector"`
	Continuation   []*Primitive `xml:"continuation"`
	InOutVariable  []*Primitive `xml:"inOutVariable"`
	InVariable     []*Primitive `xml:"inVariable"`
	OutVariable    []*Primitive `xml:"outVariable"`
	Block          []*Block     `xml:"block"`
}

type ConnectionPoint struct {
	FormalParameter string       `xml:"formalParameter,attr,omitempty"`
	RelPosition     Position     `xml:"relPosition"`
	Connection      []Connection `xml:"connection"`
}

type Connection struct {
	RefLocalId      string     `xml:"refLocalId,attr"`
	FormalParameter string     `xml:"formalParameter,attr,omitempty"`
	Position        []Position `xml:"position"`
}

// Element is a generic struct for LD components
type Primitive struct {
	ElemType           string
	LocalId            string            `xml:"localId,attr"`
	Name               string            `xml:"name,attr,omitempty"`
	Expression         string            `xml:"expression,omitempty"`
	Position           Position          `xml:"position"`
	ConnectionPointIn  []ConnectionPoint `xml:"connectionPointIn,omitempty"`
	ConnectionPointOut []ConnectionPoint `xml:"connectionPointOut,omitempty"`
	Variable           string            `xml:"variable"`
	Negated            bool              `xml:"negated,attr"`
	Storage            string            `xml:"storage,attr,omitempty"`
	Edge               string            `xml:"edge,attr,omitempty"`
	Width              int               `xml:"width,attr"`
	Height             int               `xml:"height,attr"`
}

type Block struct {
	ElemType        string
	LocalId         string         `xml:"localId,attr"`
	TypeName        string         `xml:"typeName,attr"`
	InstanceName    string         `xml:"instanceName,attr"`
	Width           int            `xml:"width,attr"`
	Height          int            `xml:"height,attr"`
	Position        Position       `xml:"position"`
	InputVariables  BlockVariables `xml:"inputVariables"`
	InOutVariables  BlockVariables `xml:"inOutVariables"`
	OutputVariables BlockVariables `xml:"outputVariables"`
}

type Position struct {
	X int `xml:"x,attr"`
	Y int `xml:"y,attr"`
}

type Instances struct {
	Configurations Configurations `xml:"configurations"`
}

type Configurations struct {
	Configuration []Configuration `xml:"configuration"`
}

type Configuration struct {
	Name     string     `xml:"name,attr"`
	Resource []Resource `xml:"resource"`
}

type Resource struct {
	Name string `xml:"name,attr"`
	Task []Task `xml:"task"`
}

type Task struct {
	Name        string      `xml:"name,attr"`
	Priority    int         `xml:"priority,attr"`
	Interval    string      `xml:"interval,attr"`
	POUInstance POUInstance `xml:"pouInstance"`
}

type POUInstance struct {
	Name     string `xml:"name,attr"`
	TypeName string `xml:"typeName,attr"`
}

func (project *Project) GetPouByName(name string) (pou POU, err error) {
	for _, pou := range project.Types.POUs.POU {
		if pou.Name == name {
			return pou, nil
		}
	}
	return POU{}, fmt.Errorf("no POU with name %s available", name)
}

func (ld *LD) ensurePrimitiveTypeLabels() {
	for _, prim := range ld.Contact {
		prim.ElemType = "contact"
	}
	for _, prim := range ld.Coil {
		prim.ElemType = "coil"
	}
	for _, prim := range ld.LeftPowerRail {
		prim.ElemType = "leftPowerRail"
	}
	for _, prim := range ld.RightPowerRail {
		prim.ElemType = "rightPowerRail"
	}
	for _, prim := range ld.Connector {
		prim.ElemType = "connector"
	}
	for _, prim := range ld.Continuation {
		prim.ElemType = "continuation"
	}
	for _, prim := range ld.InOutVariable {
		prim.ElemType = "inOutVariable"
	}
	for _, prim := range ld.InVariable {
		prim.ElemType = "inVariable"
	}
	for _, prim := range ld.OutVariable {
		prim.ElemType = "outVariable"
	}
}

func (ld *LD) ensureBlockTypeLabels() {
	for _, block := range ld.Block {
		block.ElemType = "block"
	}
}

// TODO: Optimize, make it more generalized, don't hardcode field names
func (ld *LD) GatherAllPrimitives() []*Primitive {
	var all []*Primitive
	ld.ensurePrimitiveTypeLabels()
	ld.ensureBlockTypeLabels()
	all = append(all, ld.LeftPowerRail...)
	all = append(all, ld.Contact...)
	all = append(all, ld.Coil...)
	all = append(all, ld.RightPowerRail...)
	all = append(all, ld.Connector...)
	all = append(all, ld.Continuation...)
	all = append(all, ld.InOutVariable...)
	all = append(all, ld.InVariable...)
	all = append(all, ld.OutVariable...)
	return all
}

func (ld LD) GatherAllBlocks() []*Block {
	return ld.Block
}
