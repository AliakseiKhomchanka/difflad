// NOTES:
// 1 cell in OpenPLC IDE is 10 by 10 pixels

package svg

import (
	"encoding/xml"
	"fmt"
	elements "openplc-render/elements"
	"strconv"
)

const CELL_SIZE int = 10

var diff_color = map[elements.Diff]string{
	elements.DiffAdded:     "green",
	elements.DiffDeleted:   "red",
	elements.DiffUnchanged: "white",
}

var stroke_width = map[elements.Diff]int{
	elements.DiffAdded:     5,
	elements.DiffDeleted:   1,
	elements.DiffUnchanged: 1,
}

var stroke_dasharray = map[elements.Diff]string{
	elements.DiffAdded:     "",
	elements.DiffDeleted:   "4",
	elements.DiffUnchanged: "",
}

var text_decoration = map[elements.Diff]string{
	elements.DiffAdded:     "",
	elements.DiffDeleted:   "line-through",
	elements.DiffUnchanged: "",
}

var font_weight = map[elements.Diff]string{
	elements.DiffAdded:     "bold",
	elements.DiffDeleted:   "",
	elements.DiffUnchanged: "",
}

type SVGFile struct {
	XMLName  xml.Name  `xml:"svg"`
	ViewBox  string    `xml:"viewBox,attr"`
	Xmlns    string    `xml:"xmlns,attr"`
	Elements []Element `xml:",any"`
}

// Background rect
type Background struct {
	XMLName xml.Name `xml:"rect"`
	Width   int      `xml:"width,attr"`
	Height  int      `xml:"height,attr"`
	Fill    string   `xml:"fill,attr,omitempty"`
}

type Rect struct {
	XMLName         xml.Name `xml:"rect"`
	Width           int      `xml:"width,attr"`
	Height          int      `xml:"height,attr"`
	X               int      `xml:"x,attr"`
	Y               int      `xml:"y,attr"`
	Fill            string   `xml:"fill,attr,omitempty"`
	FillOpacity     float32  `xml:"fill-opacity,attr,omitempty"`
	Stroke          string   `xml:"stroke,attr,omitempty"`
	StrokeWidth     int      `xml:"stroke-width,attr,omitempty"`
	StrokeDasharray string   `xml:"stroke-dasharray,attr,omitempty"`
}

type Line struct {
	XMLName         xml.Name `xml:"line"`
	X1              int      `xml:"x1,attr"`
	Y1              int      `xml:"y1,attr"`
	X2              int      `xml:"x2,attr"`
	Y2              int      `xml:"y2,attr"`
	Stroke          string   `xml:"stroke,attr,omitempty"`
	StrokeWidth     int      `xml:"stroke-width,attr,omitempty"`
	StrokeDasharray string   `xml:"stroke-dasharray,attr,omitempty"`
}

type Polyline struct {
	XMLName         xml.Name `xml:"polyline"`
	Points          string   `xml:"points,attr"`
	Stroke          string   `xml:"stroke,attr,omitempty"`
	StrokeWidth     int      `xml:"stroke-width,attr,omitempty"`
	Fill            string   `xml:"fill,attr,omitempty"`
	FillOpacity     float32  `xml:"fill-opacity,attr,omitempty"`
	StrokeDasharray string   `xml:"stroke-dasharray,attr,omitempty"`
}

type Group struct {
	XMLName  xml.Name   `xml:"g"`
	Line     []Line     `xml:"line,omitempty"`
	Rect     []Rect     `xml:"rect,omitempty"`
	Text     []Text     `xml:"text,omitempty"`
	Path     []Path     `xml:"path,omitempty"`
	Polyline []Polyline `xml:"polyline,omitempty"`
}

type Text struct {
	XMLName        xml.Name `xml:"text"`
	X              int      `xml:"x,attr"`
	Y              int      `xml:"y,attr"`
	TextAnchor     string   `xml:"text-anchor,attr"`
	LengthAdjust   string   `xml:"lengthAdjust,attr"`
	TextLength     string   `xml:"textLength,attr"`
	TextDecoration string   `xml:"text-decoration,attr,omitempty"`
	FontFamily     string   `xml:"font-family,attr,omitempty"`
	FontStyle      string   `xml:"font-style,attr,omitempty"`
	FontSize       string   `xml:"font-size,attr,omitempty"`
	FontWeight     string   `xml:"font-weight,attr,omitempty"`
	Fill           string   `xml:"fill,attr,omitempty"`
	FillOpacity    float32  `xml:"fill-opacity,attr,omitempty"`
	Content        string   `xml:",chardata"`
}

type Path struct {
	XMLName         xml.Name `xml:"path"`
	D               string   `xml:"d,attr"`
	Stroke          string   `xml:"stroke,attr,omitempty"`
	StrokeWidth     int      `xml:"stroke-width,attr,omitempty"`
	StrokeDasharray string   `xml:"stroke-dasharray,attr,omitempty"`
	Fill            string   `xml:"fill,attr,omitempty"`
	FillOpacity     float32  `xml:"fill-opacity,attr,omitempty"`
}

type Element interface{}

func renderContact(elem *elements.Element) Group {
	line_1 := Line{
		X1:              elem.Position.X,
		Y1:              elem.Position.Y,
		X2:              elem.Position.X,
		Y2:              elem.Position.Y + elem.Height,
		Stroke:          diff_color[elem.Diff],
		StrokeWidth:     stroke_width[elem.Diff],
		StrokeDasharray: stroke_dasharray[elem.Diff],
	}
	line_2 := Line{
		X1:              elem.Position.X + elem.Width,
		Y1:              elem.Position.Y,
		X2:              elem.Position.X + elem.Width,
		Y2:              elem.Position.Y + elem.Height,
		Stroke:          diff_color[elem.Diff],
		StrokeWidth:     stroke_width[elem.Diff],
		StrokeDasharray: stroke_dasharray[elem.Diff],
	}
	text := Text{
		X:              elem.Position.X + (elem.Width / 2),
		Y:              elem.Position.Y - CELL_SIZE/2,
		Content:        elem.TopLabel.Value,
		TextAnchor:     "middle",
		FontFamily:     "arial",
		FontSize:       strconv.Itoa(CELL_SIZE + CELL_SIZE/4),
		Fill:           diff_color[elem.TopLabel.Diff],
		TextDecoration: text_decoration[elem.TopLabel.Diff],
		FontWeight:     font_weight[elem.TopLabel.Diff],
	}
	inner_text := Text{
		X:              elem.Position.X + (elem.Width / 2),
		Y:              elem.Position.Y + elem.Height/2 + elem.Height/4,
		Content:        elem.ElementText.Value,
		TextAnchor:     "middle",
		FontFamily:     "arial",
		Fill:           diff_color[elem.ElementText.Diff],
		TextDecoration: text_decoration[elem.ElementText.Diff],
		FontWeight:     font_weight[elem.ElementText.Diff],
	}
	return Group{
		Line: []Line{line_1, line_2},
		Text: []Text{text, inner_text},
	}
}

func renderCoil(elem *elements.Element) Group {
	curve_left_d := fmt.Sprintf("M %d %d Q %d %d %d %d", elem.Position.X+CELL_SIZE/2, elem.Position.Y, elem.Position.X-CELL_SIZE/2, elem.Position.Y+elem.Height/2, elem.Position.X+CELL_SIZE/2, elem.Position.Y+elem.Height)
	curve_left := Path{
		D:               curve_left_d,
		Stroke:          diff_color[elem.Diff],
		StrokeWidth:     stroke_width[elem.Diff],
		StrokeDasharray: stroke_dasharray[elem.Diff],
		Fill:            "transparent",
	}
	curve_right_d := fmt.Sprintf("M %d %d Q %d %d %d %d", elem.Position.X+elem.Width-CELL_SIZE/2, elem.Position.Y, elem.Position.X+elem.Width+CELL_SIZE/2, elem.Position.Y+elem.Height/2, elem.Position.X+elem.Width-CELL_SIZE/2, elem.Position.Y+elem.Height)
	curve_right := Path{
		D:               curve_right_d,
		Stroke:          diff_color[elem.Diff],
		StrokeWidth:     stroke_width[elem.Diff],
		StrokeDasharray: stroke_dasharray[elem.Diff],
		Fill:            "transparent",
	}
	text := Text{
		X:              elem.Position.X + (elem.Width / 2),
		Y:              elem.Position.Y - CELL_SIZE/2,
		Content:        elem.TopLabel.Value,
		TextAnchor:     "middle",
		FontFamily:     "arial",
		FontSize:       strconv.Itoa(CELL_SIZE + CELL_SIZE/4),
		Fill:           diff_color[elem.TopLabel.Diff],
		TextDecoration: text_decoration[elem.TopLabel.Diff],
		FontWeight:     font_weight[elem.TopLabel.Diff],
	}
	inner_text := Text{
		X:              elem.Position.X + (elem.Width / 2),
		Y:              elem.Position.Y + elem.Height/2 + elem.Height/4,
		Content:        elem.ElementText.Value,
		TextAnchor:     "middle",
		FontFamily:     "arial",
		Fill:           diff_color[elem.ElementText.Diff],
		TextDecoration: text_decoration[elem.ElementText.Diff],
		FontWeight:     font_weight[elem.ElementText.Diff],
	}
	return Group{
		Path: []Path{curve_left, curve_right},
		Text: []Text{text, inner_text},
	}
}

func renderConnectorOrContinuation(elem *elements.Element) Group {
	group := Group{}
	box := Rect{
		Width:           elem.Width,
		Height:          elem.Height,
		X:               elem.Position.X,
		Y:               elem.Position.Y,
		Fill:            "transparent",
		Stroke:          diff_color[elem.Diff],
		StrokeWidth:     stroke_width[elem.Diff],
		StrokeDasharray: stroke_dasharray[elem.Diff],
	}
	group.Rect = append(group.Rect, box)
	// Render arrow lines on both sides
	// Left
	points := ""
	points += fmt.Sprintf("%d,%d ", elem.Position.X, elem.Position.Y)
	points += fmt.Sprintf("%d,%d ", elem.Position.X+elem.Height/2, elem.Position.Y+elem.Height/2)
	points += fmt.Sprintf("%d,%d ", elem.Position.X, elem.Position.Y+elem.Height)
	group.Polyline = append(group.Polyline, Polyline{
		Points:          points,
		Stroke:          diff_color[elem.Diff],
		StrokeWidth:     1,
		StrokeDasharray: stroke_dasharray[elem.Diff],
		Fill:            "transparent",
	})
	//Right
	points = ""
	points += fmt.Sprintf("%d,%d ", elem.Position.X+elem.Width-elem.Height/2, elem.Position.Y)
	points += fmt.Sprintf("%d,%d ", elem.Position.X+elem.Width, elem.Position.Y+elem.Height/2)
	points += fmt.Sprintf("%d,%d ", elem.Position.X+elem.Width-elem.Height/2, elem.Position.Y+elem.Height)
	group.Polyline = append(group.Polyline, Polyline{
		Points:          points,
		Stroke:          diff_color[elem.Diff],
		StrokeWidth:     1,
		StrokeDasharray: stroke_dasharray[elem.Diff],
		Fill:            "transparent",
	})
	// Element text
	elem_text := Text{
		X:          elem.Position.X + (elem.Width / 2),
		Y:          elem.Position.Y + CELL_SIZE*2,
		Content:    elem.ElementText.Value,
		TextAnchor: "middle",
		FontFamily: "arial",
		FontSize:   strconv.Itoa(CELL_SIZE + CELL_SIZE/4),
		Fill:       diff_color[elem.ElementText.Diff],
	}
	group.Text = append(group.Text, elem_text)
	return group
}

func renderVariable(elem *elements.Element) Group {
	group := Group{}
	box := Rect{
		Width:           elem.Width,
		Height:          elem.Height,
		X:               elem.Position.X,
		Y:               elem.Position.Y,
		Fill:            "transparent",
		Stroke:          diff_color[elem.Diff],
		StrokeWidth:     stroke_width[elem.Diff],
		StrokeDasharray: stroke_dasharray[elem.Diff],
	}
	group.Rect = append(group.Rect, box)
	// Element text
	elem_text := Text{
		X:          elem.Position.X + (elem.Width / 2),
		Y:          elem.Position.Y + CELL_SIZE*2,
		Content:    elem.ElementText.Value,
		TextAnchor: "middle",
		FontFamily: "arial",
		FontSize:   strconv.Itoa(CELL_SIZE + CELL_SIZE/4),
		Fill:       diff_color[elem.ElementText.Diff],
	}
	group.Text = append(group.Text, elem_text)
	return group
}

func renderBlock(elem *elements.Element) Group {
	group := Group{}
	box := Rect{
		Width:           elem.Width,
		Height:          elem.Height,
		X:               elem.Position.X,
		Y:               elem.Position.Y,
		Fill:            "transparent",
		Stroke:          diff_color[elem.Diff],
		StrokeWidth:     stroke_width[elem.Diff],
		StrokeDasharray: stroke_dasharray[elem.Diff],
	}
	group.Rect = append(group.Rect, box)
	box_type_text := Text{
		X:          elem.Position.X + (elem.Width / 2),
		Y:          elem.Position.Y + CELL_SIZE + CELL_SIZE/2,
		Content:    elem.BlockLabel.Value,
		TextAnchor: "middle",
		FontFamily: "arial",
		FontSize:   strconv.Itoa(CELL_SIZE + CELL_SIZE/4),
		Fill:       diff_color[elem.BlockLabel.Diff],
	}
	group.Text = append(group.Text, box_type_text)
	top_text := Text{
		X:          elem.Position.X + (elem.Width / 2),
		Y:          elem.Position.Y - CELL_SIZE/2 - CELL_SIZE/4,
		Content:    elem.TopLabel.Value,
		TextAnchor: "middle",
		FontFamily: "arial",
		FontSize:   strconv.Itoa(CELL_SIZE + CELL_SIZE/4),
		Fill:       diff_color[elem.TopLabel.Diff],
	}
	group.Text = append(group.Text, top_text)
	// Input pins
	for _, pin := range elem.Inputs {
		pin_text := Text{
			X:          elem.Position.X + pin.Position.X + CELL_SIZE/2,
			Y:          elem.Position.Y + pin.Position.Y + CELL_SIZE/2,
			Content:    pin.Label.Value,
			TextAnchor: "left",
			FontFamily: "arial",
			FontSize:   strconv.Itoa(CELL_SIZE + CELL_SIZE/4),
			Fill:       diff_color[pin.Label.Diff],
		}
		group.Text = append(group.Text, pin_text)
	}
	// Output pins
	for _, pin := range elem.Outputs {
		pin_text := Text{
			X:          elem.Position.X + pin.Position.X - CELL_SIZE/2,
			Y:          elem.Position.Y + pin.Position.Y + CELL_SIZE/2,
			Content:    pin.Label.Value,
			TextAnchor: "end",
			FontFamily: "arial",
			FontSize:   strconv.Itoa(CELL_SIZE + CELL_SIZE/4),
			Fill:       diff_color[pin.Label.Diff],
		}
		group.Text = append(group.Text, pin_text)
	}
	return group
}

func renderLeftPowerRail(elem *elements.Element) Group {
	group := Group{}
	line := Line{
		X1:          elem.Position.X,
		Y1:          elem.Position.Y,
		X2:          elem.Position.X,
		Y2:          elem.Position.Y + elem.Height,
		Stroke:      diff_color[elem.Diff],
		StrokeWidth: 3,
	}
	group.Line = append(group.Line, line)
	// Add little stubs for output pins
	for _, pin := range elem.Outputs {
		pin_line := Line{
			X1:     elem.Position.X,
			Y1:     elem.Position.Y + pin.Position.Y,
			X2:     elem.Position.X + pin.Position.X,
			Y2:     elem.Position.Y + pin.Position.Y,
			Stroke: diff_color[pin.Label.Diff],
		}
		group.Line = append(group.Line, pin_line)
	}
	return group
}

// Pin coordinates are slightly different for the right rail in the XML file, shifted right by one cell width
func renderRightPowerRail(elem *elements.Element) Group {
	group := Group{}
	line := Line{
		X1:          elem.Position.X,
		Y1:          elem.Position.Y,
		X2:          elem.Position.X,
		Y2:          elem.Position.Y + elem.Height,
		Stroke:      diff_color[elem.Diff],
		StrokeWidth: 3,
	}
	group.Line = append(group.Line, line)
	// Add little stubs for input pins
	for _, pin := range elem.Inputs {
		pin_line := Line{
			X1:     elem.Position.X,
			Y1:     elem.Position.Y + pin.Position.Y,
			X2:     elem.Position.X - CELL_SIZE,
			Y2:     elem.Position.Y + pin.Position.Y,
			Stroke: diff_color[pin.Label.Diff],
		}
		group.Line = append(group.Line, pin_line)
	}
	return group
}

func renderConnections(elem *elements.Element) Group {
	group := Group{}
	// Inputs
	for _, pin := range elem.Inputs {
		for _, conn := range pin.Connections {
			points := ""
			for _, point := range conn.Points {
				points += fmt.Sprintf("%d,%d ", point.X, point.Y)
			}
			group.Polyline = append(group.Polyline, Polyline{
				Points:          points,
				Stroke:          diff_color[conn.Diff],
				StrokeWidth:     stroke_width[conn.Diff],
				StrokeDasharray: stroke_dasharray[conn.Diff],
				Fill:            "transparent",
			})
		}
	}
	// Outputs
	for _, pin := range elem.Outputs {
		for _, conn := range pin.Connections {
			points := ""
			for _, point := range conn.Points {
				points += fmt.Sprintf("%d,%d ", point.X, point.Y)
			}
			group.Polyline = append(group.Polyline, Polyline{
				Points:          points,
				Stroke:          diff_color[conn.Diff],
				StrokeWidth:     stroke_width[conn.Diff],
				StrokeDasharray: stroke_dasharray[conn.Diff],
				Fill:            "transparent",
			})
		}
	}
	return group
}

func calculateViewBox(pou elements.POU) (x, y int) {
	maxX := 0
	maxY := 0
	for _, elem := range pou.Elements {
		if elem.Position.X+elem.Width > maxX {
			maxX = elem.Position.X + elem.Width
		}
		if elem.Position.Y+elem.Height > maxY {
			maxY = elem.Position.Y + elem.Height
		}
		for _, pin := range elem.Inputs {
			for _, conn := range pin.Connections {
				for _, point := range conn.Points {
					if point.X > maxX {
						maxX = point.X
					}
					if point.Y > maxY {
						maxY = point.Y
					}
				}
			}
		}
		for _, pin := range elem.Outputs {
			for _, conn := range pin.Connections {
				for _, point := range conn.Points {
					if point.X > maxX {
						maxX = point.X
					}
					if point.Y > maxY {
						maxY = point.Y
					}
				}
			}
		}
	}
	return maxX + 10, maxY + 10
}

func renderBackground(width, height int, style string) Background {
	var fill string
	switch style {
	case "dark":
		fill = "#0d1117"
	case "light":
		fill = "#f6f8fa"
	}
	return Background{
		Width:  width,
		Height: height,
		Fill:   fill,
	}
}

func setStyle(style string) {
	switch style {
	case "dark":
		diff_color[elements.DiffUnchanged] = "white"
	case "light":
		diff_color[elements.DiffUnchanged] = "black"
	}
}

func RenderPOU(pou elements.POU, style string) SVGFile {
	var file SVGFile
	// Init SVG file headers and metadata
	viewX, viewY := calculateViewBox(pou)
	file.ViewBox = fmt.Sprintf("0 0 %d %d", viewX, viewY)
	file.Xmlns = "http://www.w3.org/2000/svg"
	// Add background
	file.Elements = append(file.Elements, renderBackground(viewX, viewY, style))
	setStyle(style)
	// Render elements
	for _, element := range pou.Elements {
		//.Printf("ELEM: %v\n", element)
		switch element.Type {
		case "contact":
			geometry := renderContact(element)
			file.Elements = append(file.Elements, geometry)
		case "coil":
			geometry := renderCoil(element)
			file.Elements = append(file.Elements, geometry)
		case "connector", "continuation":
			geometry := renderConnectorOrContinuation(element)
			file.Elements = append(file.Elements, geometry)
		case "inOutVariable", "inVariable", "outVariable":
			fmt.Printf("RENDERING VARIABLE: %s", element.ElementText.Value)
			geometry := renderVariable(element)
			file.Elements = append(file.Elements, geometry)
		case "block":
			geometry := renderBlock(element)
			file.Elements = append(file.Elements, geometry)
		case "leftPowerRail":
			geometry := renderLeftPowerRail(element)
			file.Elements = append(file.Elements, geometry)
		case "rightPowerRail":
			geometry := renderRightPowerRail(element)
			file.Elements = append(file.Elements, geometry)
		default:
			svg_elem := Rect{
				Width:  element.Width,
				Height: element.Height,
				X:      element.Position.X,
				Y:      element.Position.Y,
				Fill:   "white",
				Stroke: "black",
			}
			file.Elements = append(file.Elements, svg_elem)
		}
		connection_group := renderConnections(element)
		file.Elements = append(file.Elements, connection_group)
	}
	//fmt.Printf("FILE ELEMENTS: %v\n", file.Elements)
	return file
}
