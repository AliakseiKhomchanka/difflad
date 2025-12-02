// TODO: Add more geometric features for visually impaired

package elements

import (
	"fmt"
	plcxml "openplc-render/xml"
	"os"
)

// Consts

type Diff int

const (
	DiffUnchanged Diff = iota
	DiffDeleted
	DiffAdded
)

// Used to represent fields that can have a diff
type MutableString struct {
	Value string
	Diff  Diff
}

type Position struct {
	X int
	Y int
}

// Since connections are tracked separately, they include detailed information
// on pins on both sides for the ease of diffing.
// We wouldn't need to be this verbose for just rendering, but for diffing
// it becomes essential since we need to track not only whether we're connected to the same element,
// but also to the same pin in it (and we can't rely on coordinates here since they can change)
type Connection struct {
	TargetRef   string
	TargetLabel string
	TargetPin   int
	Points      []*Position
	Diff        Diff
}

type Pin struct {
	Position    Position // Topographical coordinates
	Order       int      // In the order of inputs/outputs, i.e 1st, 2nd pin and so on
	Label       MutableString
	Connections []*Connection
}

type Element struct {
	UID         string
	Type        string
	Position    Position
	Width       int
	Height      int
	ElementText MutableString // Like a slash for a negated contact, for example
	TopLabel    MutableString
	BottomLabel MutableString
	BlockLabel  MutableString
	Inputs      []*Pin
	Outputs     []*Pin
	Diff        Diff
}

type POU struct {
	Name     string
	Comment  string
	Elements map[string]*Element
}

// TODO: Add description and comment processing
func (p *POU) Parse(pou plcxml.POU) error {
	p.Name = pou.Name
	p.parseElements(pou)
	return nil
}

func (p *POU) parseElements(pou plcxml.POU) error {
	p.Name = pou.Name
	p.Elements = make(map[string]*Element)
	// Step 1: parse primitives first
	primitives := pou.Body.LD.GatherAllPrimitives()
	for _, prim := range primitives {
		new_prim, err := initPrimitiveFromXML(*prim)
		if err != nil {
			fmt.Printf("Error parsing primitive: %v\n", err)
			os.Exit(1)
		}
		p.Elements[new_prim.UID] = new_prim
	}
	// Step 2: parse blocks
	blocks := pou.Body.LD.GatherAllBlocks()
	for _, block := range blocks {
		new_block, err := initBlockFromXML(*block)
		if err != nil {
			fmt.Printf("Error parsing primitive: %v\n", err)
			os.Exit(1)
		}
		p.Elements[new_block.UID] = new_block
	}
	return nil
}

// Not necessarily the cleanest way at scale, but for a small case
// like this I find it better for readability, otherwise use some
// additional type-text mapping structure at package level
func getPrimitiveText(prim plcxml.Primitive) MutableString {
	text := ""
	switch prim.ElemType {
	case "contact":
		if prim.Negated {
			text = "/"
			break
		}
		if prim.Edge == "rising" {
			text = "P"
		}
		if prim.Edge == "falling" {
			text = "N"
		}
	case "coil":
		if prim.Negated {
			text = "/"
		}
		if prim.Edge == "rising" {
			text = "P"
		}
		if prim.Storage == "set" {
			text = "S"
		}
		if prim.Storage == "reset" {
			text = "R"
		}
		if prim.Edge == "falling" {
			text = "N"
		}
	case "connector", "continuation":
		text = prim.Name
	case "inOutVariable", "inVariable", "outVariable":
		text = prim.Expression
	}
	return MutableString{
		Value: text,
	}
}

func initPrimitiveFromXML(prim plcxml.Primitive) (*Element, error) {
	new_prim := Element{}
	// Process basic fields
	new_prim.UID = prim.LocalId
	new_prim.Type = prim.ElemType
	new_prim.Position = Position(prim.Position)
	new_prim.Width = prim.Width
	new_prim.Height = prim.Height
	new_prim.ElementText = getPrimitiveText(prim)
	if prim.Variable != "" {
		new_prim.TopLabel = MutableString{
			Value: prim.Variable,
		}
	}
	// Process inputs
	for pin_index, in_pin := range prim.ConnectionPointIn {
		pin_position := Position(in_pin.RelPosition)
		pin_order := pin_index
		pin_connections := []*Connection{}
		// Process connections
		for _, conn := range in_pin.Connection {
			conn_target_ref := conn.RefLocalId
			conn_target_label := conn.FormalParameter
			conn_target_pin := 0
			conn_points := []*Position{}
			for _, pos := range conn.Position {
				parsed_position := Position(pos)
				conn_points = append(conn_points, &parsed_position)
			}
			pin_connections = append(pin_connections, &Connection{
				TargetRef:   conn_target_ref,
				TargetLabel: conn_target_label,
				TargetPin:   conn_target_pin,
				Points:      conn_points,
			})
		}
		new_prim.Inputs = append(new_prim.Inputs, &Pin{
			Position:    pin_position,
			Order:       pin_order,
			Connections: pin_connections,
		})
	}
	// Process outputs
	for pin_index, out_pin := range prim.ConnectionPointOut {
		pin_position := Position(out_pin.RelPosition)
		pin_order := pin_index
		pin_connections := []*Connection{}
		// Process connections
		for _, conn := range out_pin.Connection {
			conn_target_ref := conn.RefLocalId
			conn_target_label := conn.FormalParameter
			conn_target_pin := 0
			conn_points := []*Position{}
			for _, pos := range conn.Position {
				parsed_position := Position(pos)
				conn_points = append(conn_points, &parsed_position)
			}
			pin_connections = append(pin_connections, &Connection{
				TargetRef:   conn_target_ref,
				TargetLabel: conn_target_label,
				TargetPin:   conn_target_pin,
				Points:      conn_points,
			})
		}
		new_prim.Outputs = append(new_prim.Outputs, &Pin{
			Position:    pin_position,
			Order:       pin_order,
			Connections: pin_connections,
		})
	}
	return &new_prim, nil
}

func initBlockFromXML(block plcxml.Block) (*Element, error) {
	new_block := Element{}
	new_block.UID = block.LocalId
	new_block.Type = block.ElemType
	new_block.Position = Position(block.Position)
	new_block.Width = block.Width
	new_block.Height = block.Height
	new_block.TopLabel = MutableString{
		Value: block.InstanceName,
	}
	new_block.BlockLabel = MutableString{
		Value: block.TypeName,
	}
	// Handle inputs
	for pin_index, variable := range block.InputVariables.Variable {
		pin_position := Position(variable.ConnectionPointIn[0].RelPosition)
		pin_order := pin_index
		pin_label := MutableString{
			Value: variable.FormalParameter,
		}
		pin_connections := []*Connection{}
		for _, conn := range variable.ConnectionPointIn[0].Connection {
			conn_target_ref := conn.RefLocalId
			conn_target_label := conn.FormalParameter
			conn_target_pin := 0
			conn_points := []*Position{}
			for _, pos := range conn.Position {
				parsed_position := Position(pos)
				conn_points = append(conn_points, &parsed_position)
			}
			pin_connections = append(pin_connections, &Connection{
				TargetRef:   conn_target_ref,
				TargetLabel: conn_target_label,
				TargetPin:   conn_target_pin,
				Points:      conn_points,
			})
		}
		new_block.Inputs = append(new_block.Inputs, &Pin{
			Position:    pin_position,
			Order:       pin_order,
			Label:       pin_label,
			Connections: pin_connections,
		})
	}
	// Handle outputs
	for pin_index, variable := range block.OutputVariables.Variable {
		pin_position := Position(variable.ConnectionPointOut[0].RelPosition)
		pin_order := pin_index
		pin_label := MutableString{
			Value: variable.FormalParameter,
		}
		pin_connections := []*Connection{}
		for _, conn := range variable.ConnectionPointOut[0].Connection {
			conn_target_ref := conn.RefLocalId
			conn_target_label := conn.FormalParameter
			conn_target_pin := 0
			conn_points := []*Position{}
			for _, pos := range conn.Position {
				parsed_position := Position(pos)
				conn_points = append(conn_points, &parsed_position)
			}
			pin_connections = append(pin_connections, &Connection{
				TargetRef:   conn_target_ref,
				TargetLabel: conn_target_label,
				TargetPin:   conn_target_pin,
				Points:      conn_points,
			})
		}
		new_block.Outputs = append(new_block.Outputs, &Pin{
			Position:    pin_position,
			Order:       pin_order,
			Label:       pin_label,
			Connections: pin_connections,
		})
	}

	return &new_block, nil
}

// Diffing logic

func (p *POU) CalculateDiff(new_pou *POU) {
	// Layer 1: diff elements
outer:
	for _, elem := range p.Elements {
		// For each element, check if an element with the same ref exists in the second POU
		for _, elem2 := range new_pou.Elements {
			if elem.UID == elem2.UID {
				// Check if element type is the same
				if elem.Type == elem2.Type {
					if elem.Type == "block" {
						// fmt.Printf("BLOCK TYPE\n")
						// They may both be blocks, but of different types, check that
						if elem.BlockLabel.Value == elem2.BlockLabel.Value {
							// Check top label (instance name in case of blocks)
							if elem.TopLabel.Value != elem2.TopLabel.Value {
								elem.TopLabel.Diff = DiffDeleted
								elem2.TopLabel.Diff = DiffAdded
							}
							elem.connectionsDiff(elem2)
							continue outer
						} else {
							elem.Diff = DiffDeleted
							elem2.Diff = DiffAdded
							// Layer 2: diff connections
							elem.connectionsDiff(elem2)
							continue outer
						}
					}
				}
				if elem.TopLabel.Value != elem2.TopLabel.Value {
					elem.TopLabel.Diff = DiffDeleted
					elem2.TopLabel.Diff = DiffAdded
				}
				// Primitive type may be the same (contact, for example), but the subtype could be negated, for example, affecting the element text
				if elem.ElementText.Value != elem2.ElementText.Value {
					elem.ElementText.Diff = DiffDeleted
					elem2.ElementText.Diff = DiffAdded
				}
				// Layer 2: diff connections
				elem.connectionsDiff(elem2)
				continue outer
			}
		}
		// If not element in the new version matched the UID - it has been deleted
		elem.Diff = DiffDeleted
		elem.markAllConnectionsDeleted()
		elem.markAllLabelsDeleted()
	}
	// Check backwards the same way with flipped logic
	// Elements with matching UIDs have already been handled, now we just need to find ones exclusive to the new version
outer_back:
	for _, elem := range new_pou.Elements {
		// For each element, check if an element with the same ref exists in the second POU
		for _, elem2 := range p.Elements {
			if elem.UID == elem2.UID {
				continue outer_back
			}
		}
		// If no eleemnt in the old version matched the UID - it's an new one
		elem.Diff = DiffAdded
		elem.markAllConnectionsAdded()
		elem.markAllLabelsAdded()
	}
}

func (e *Element) markAllConnectionsDeleted() {
	for _, pin := range e.Inputs {
		pin.Label.Diff = DiffDeleted
		for _, conn := range pin.Connections {
			conn.Diff = DiffDeleted
		}
	}
	for _, pin := range e.Outputs {
		pin.Label.Diff = DiffDeleted
		for _, conn := range pin.Connections {
			conn.Diff = DiffDeleted
		}
	}
}

func (e *Element) markAllConnectionsAdded() {
	for _, pin := range e.Inputs {
		pin.Label.Diff = DiffAdded
		for _, conn := range pin.Connections {
			conn.Diff = DiffAdded
		}
	}
	for _, pin := range e.Outputs {
		pin.Label.Diff = DiffAdded
		for _, conn := range pin.Connections {
			conn.Diff = DiffAdded
		}
	}
}

func (e *Element) markAllLabelsDeleted() {
	e.ElementText.Diff = DiffDeleted
	e.TopLabel.Diff = DiffDeleted
	e.BottomLabel.Diff = DiffDeleted
	e.BlockLabel.Diff = DiffDeleted
}

func (e *Element) markAllLabelsAdded() {
	e.ElementText.Diff = DiffAdded
	e.TopLabel.Diff = DiffAdded
	e.BottomLabel.Diff = DiffAdded
	e.BlockLabel.Diff = DiffAdded
}

// TODO: Consider possible different number of inputs and outputs in different versions
func (e *Element) connectionsDiff(new_elem *Element) {
	// Forward check
	for i, pin := range e.Inputs {
		for _, conn := range pin.Connections {
			matched := false
			if i >= len(new_elem.Inputs) {
				conn.Diff = DiffDeleted
				continue
			}
			for _, conn2 := range new_elem.Inputs[i].Connections {
				if conn.TargetRef == conn2.TargetRef && conn.TargetLabel == conn2.TargetLabel {
					matched = true
				}
			}
			if !matched {
				conn.Diff = DiffDeleted
			}
		}
	}
	for i, pin := range e.Outputs {
		for _, conn := range pin.Connections {
			matched := false
			if i >= len(new_elem.Outputs) {
				conn.Diff = DiffDeleted
				continue
			}
			for _, conn2 := range new_elem.Outputs[i].Connections {
				if conn.TargetRef == conn2.TargetRef && conn.TargetLabel == conn2.TargetLabel {
					matched = true
				}
			}
			if !matched {
				conn.Diff = DiffDeleted
			}
		}
	}
	// Backward check
	for i, pin := range new_elem.Inputs {
		for _, conn := range pin.Connections {
			matched := false
			if i >= len(e.Inputs) {
				conn.Diff = DiffAdded
				continue
			}
			for _, conn2 := range e.Inputs[i].Connections {

				if conn.TargetRef == conn2.TargetRef && conn.TargetLabel == conn2.TargetLabel {
					matched = true
				}
			}
			if !matched {
				conn.Diff = DiffAdded
			}
		}
	}
	for i, pin := range new_elem.Outputs {
		for _, conn := range pin.Connections {
			matched := false
			if i >= len(e.Outputs) {
				conn.Diff = DiffAdded
				continue
			}
			for _, conn2 := range e.Outputs[i].Connections {
				if conn.TargetRef == conn2.TargetRef && conn.TargetLabel == conn2.TargetLabel {
					matched = true
				}
			}
			if !matched {
				conn.Diff = DiffAdded
			}
		}
	}
}
