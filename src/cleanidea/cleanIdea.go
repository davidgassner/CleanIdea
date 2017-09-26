package main

import (
	"fmt"
	"os"
	"path/filepath"
	"fileutil"
	"io"
	"io/ioutil"
	"strings"
)

var directories, files map[string]bool
var dirsToDelete []string
var filesToDelete []string

var projectName, newProjectName string

func main() {

	initLists()

	// Get the original project name; if not provided, exit
	if len(os.Args) < 2 {
		fmt.Println("Provide project/subdirectory name")
		os.Exit(1)
	}
	projectName = os.Args[1]
	fmt.Println("Processing project:", projectName)

	//Clean up the existing project by walking the directory structure
	//Walk directory structure and collect dirs and files to delete
	err := filepath.Walk(projectName, visitDirectory)
	printError(err)
	
	// Delete dirs and files
	for _, dirName := range dirsToDelete {
		fmt.Println("Deleting dir:", dirName)
		os.RemoveAll(dirName)
	}
	for _, fileName := range filesToDelete {
		fmt.Println("Deleting file:", fileName)
		os.Remove(fileName)
	}
	
	// If a target dir was passed, copy everything to the new dir
	if len(os.Args) == 3 {
		newProjectName = os.Args[2]
		fmt.Println("New project:", newProjectName)
		err := fileutil.CopyDir(projectName, newProjectName)
		printError(err)
		
		// Rename IML file in project root
		imlFile := newProjectName + "/" + projectName + ".iml"
		imlFileNew := newProjectName + "/" + newProjectName + ".iml"
		if _, err := os.Stat(imlFile); err == nil {
			fmt.Println("Renaming IML file")
			os.Rename(imlFile, imlFileNew)
		}
		
		// Rewrite ".name" and "modules.xml" files in ".idea" folder
		rewriteNameFile()
		rewriteModulesFile()
	}
}

// Lists of directories and files to delete
func initLists() {
	directories = make(map[string]bool)
	dirsToDelete = make([]string, 0)
	directories["build"] = true
	directories[".gradle"] = true
	directories[".git"] = true
	directories["libraries"] = true

	files = make(map[string]bool)
	filesToDelete = make([]string, 0)
	files["gradle.xml"] = true
	files["workspace.xml"] = true
	files["local.properties"] = true
	files[".DS_Store"] = true
	files["thumbs.db"] = true
}

// Called during cleanup process for existing project
// 	for each directory and each file
func visitDirectory(path string, f os.FileInfo, err error) error {
	
	if f.IsDir() {
		dirName := filepath.Base(path)
		_, ok := directories[dirName]
		if ok {
			dirsToDelete = append(dirsToDelete, path)
		}
	} else {
		fileName := filepath.Base(path)
		_, ok := files[fileName]
		if ok {
			filesToDelete = append(filesToDelete, path)
		}
	}
	return nil
}

// General error handling function
func printError(err error) bool {
	if err != nil {
		fmt.Println("Error:", err)
		return true
	} else {
		return false
	}
}

// Rewrite the .name file in .idea with the new project name
func rewriteNameFile() {
	ideaNameFile := newProjectName + "/.idea/.name"
	fmt.Println("Rewriting", ideaNameFile)
	file, _ := os.Create(ideaNameFile)
	defer file.Close()
	_, err := io.WriteString(file, newProjectName)
	printError(err)
}

func rewriteModulesFile() {
	ideaModulesFile := newProjectName + "/.idea/modules.xml"
	fmt.Println("Rewriting", ideaModulesFile)
	fmt.Println("Replacing", projectName, "with", newProjectName)
	
	// Get current file content
	content, err := ioutil.ReadFile(ideaModulesFile)
	printError(err)
	xml := string(content)
	
	// Rewrite file with new project name
	newXml := strings.Replace(xml, projectName, newProjectName, 2)
	file, _ := os.Create(ideaModulesFile)
	defer file.Close()
	_, err = io.WriteString(file, newXml)
	printError(err)
	
}
