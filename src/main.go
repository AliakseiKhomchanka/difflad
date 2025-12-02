package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	elements "openplc-render/elements"
	svg "openplc-render/svg"
	plcxml "openplc-render/xml"
)

type refList []string

func (r *refList) String() string {
	return strings.Join(*r, " ")
}

func (r *refList) Set(value string) error {
	*r = append(*r, value)
	return nil
}

func main() {
	// Get flags
	// File path, required
	filePath := flag.String("file", "", "Path to file inside the git repo")
	// Refs for diffing (or one rep for rendering without diff)
	var refs refList
	flag.Var(&refs, "ref", "One or two commit SHAs (repeatable, e.g. --ref abc --ref def)")
	pouName := flag.String("pou", "", "Which POU to render")
	outputFolder := flag.String("output", "", "Folder for output .svg files, will put them in a system temporary folder otherwise")
	style := flag.String("style", "dark", "Diagram style, \"light\"/\"dark\", dark by default")

	flag.Parse()

	// If file path or pou name are not given - exit
	log.Printf("file path provided: %s", *filePath)
	if *filePath == "" {
		log.Fatal("error: file path not provided")
	}
	if *pouName == "" {
		log.Fatal("error: pou name not provided")
	}

	// Ensure output directory exists or gets created
	if *outputFolder == "" {
		tmp, err := os.MkdirTemp("", "lad_differ-*")
		if err != nil {
			log.Fatalf("failed to create temp directory: %s", err)
		}
		*outputFolder = tmp
	} else {
		if err := os.MkdirAll(*outputFolder, 0755); err != nil {
			log.Fatalf("failed to create output directory: %s", err)
		}
	}
	log.Printf("output folder path: %s", *outputFolder)

	err := renderFiles(*filePath, *pouName, *outputFolder, *style, []string(refs)...)
	if err != nil {
		log.Fatal(err)
	}
}

func getRepoRoot(filePath string) (string, error) {
	log.Printf("file path: %s", filePath)
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = filepath.Dir(filePath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf(
			"git show failed: %w (%s)",
			err,
			strings.TrimSpace(stderr.String()),
		)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func getFileContentsFromGit(filePath, ref string) ([]byte, error) {
	repoPath, err := getRepoRoot(filePath)
	log.Printf("repo path: %s, error: %s", repoPath, err)
	if err != nil {
		return nil, err
	}
	filePath, err = filepath.Rel(repoPath, filePath)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", ref, filePath))
	cmd.Dir = repoPath
	log.Printf("git command dir: %s", cmd.Dir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf(
			"git show failed: %w (%s)",
			err,
			strings.TrimSpace(stderr.String()),
		)
	}
	return stdout.Bytes(), nil
}

func writeOutputFiles(outputFolder string, files []svg.SVGFile) error {
	for i, file := range files {
		path := filepath.Join(outputFolder, fmt.Sprintf("output_%d.svg", i))
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		svgContent, _ := xml.MarshalIndent(file, " ", "  ")
		f.Write(svgContent)
		f.Close()
	}
	return nil
}

func openOutputFolder(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)

	case "darwin":
		cmd = exec.Command("open", path)

	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", path)
	}

	return cmd.Start()
}

func renderFiles(filePath, pouName, outputFolder, style string, refs ...string) error {
	var outFiles []svg.SVGFile // Output .svg files, either one or two
	// If no refs provided - render the file at HEAD
	if len(refs) == 0 {
		refs = append(refs, "HEAD")
	}
	// Parse the first file regardless of whether the second one is provided
	contents1, err := getFileContentsFromGit(filePath, refs[0])
	if err != nil {
		log.Fatalf("error fetching file contents via git: %s", err)
	}
	var parsedPou1 elements.POU
	var project1 plcxml.Project
	err = xml.Unmarshal(contents1, &project1)
	if err != nil {
		log.Fatalf("Error parsing XML: %v\n", err)
	}
	pou, err := project1.GetPouByName(pouName)
	if err != nil {
		log.Fatal(err)
	}
	parsedPou1.Parse(pou)
	// If the second file is provided - parse it too and get the diff
	if len(refs) == 2 {
		contents2, err := getFileContentsFromGit(filePath, refs[1])
		if err != nil {
			log.Fatalf("error fetching file contents via git: %s", err)
		}
		var parsedPou2 elements.POU
		var project2 plcxml.Project
		err = xml.Unmarshal(contents2, &project2)
		if err != nil {
			log.Fatalf("Error parsing XML: %v\n", err)
		}
		pou, err := project2.GetPouByName(pouName)
		if err != nil {
			log.Fatal(err)
		}
		parsedPou2.Parse(pou)
		parsedPou1.CalculateDiff(&parsedPou2)
		outFiles = append(outFiles, svg.RenderPOU(parsedPou1, style))
		outFiles = append(outFiles, svg.RenderPOU(parsedPou2, style))
	} else {
		outFiles = append(outFiles, svg.RenderPOU(parsedPou1, style))
	}
	err = writeOutputFiles(outputFolder, outFiles)
	if err != nil {
		log.Fatal(err)
	}
	err = openOutputFolder(outputFolder)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
