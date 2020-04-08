package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	formulaOption      = "formula"
	pathOption         = "path"
	isFileOnlyOption   = "fileOnly"
	isFolderOnlyOption = "folderOnly"

	namingTypeFile   = "file_only"
	namingTypeFolder = "folder_only"
	namingTypeAll    = "file_and_folder"

	formulaIncrement = "{increment}"
	formulaCurrent   = "{current}"

	defaultIncrement = 1
)

var formulas = []string{formulaIncrement, formulaCurrent}

func main() {
	fmt.Println("Naming your files & folders quickly\n")

	path, formula, namingType, err := readArgs()
	if err != nil {
		log.Fatal(err)
	}
	if err := exec(formula, path, defaultIncrement, namingType); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully naming files & folders")
}

func exec(formula, path string, increment int, namingType string) error {
	files, folders, err := readAll(path)
	if err != nil {
		return err
	}
	prefixPath := path + "/"

	if namingType == namingTypeFile || namingType == namingTypeAll {
		if increment, err = naming(files, increment, formula, path); err != nil {
			return err
		}
	}

	if namingType == namingTypeFolder || namingType == namingTypeAll {
		if increment, err = naming(folders, increment, formula, path); err != nil {
			return err
		}
		_, folders, err = readAll(path)
	}

	for _, folder := range folders {
		if err = exec(formula, prefixPath+folder, increment+1, namingType); err != nil {
			return err
		}
	}

	return nil
}

func naming(src []string, increment int, formula, path string) (lastIncrement int, err error) {
	prefixPath := path + "/"
	for _, s := range src {
		increment++
		destName := readFormula(formula, s, increment)
		if err = rename(prefixPath+s, prefixPath+destName); err != nil {
			return
		}
	}
	lastIncrement = increment
	return
}

func readArgs() (pathDir, formulaNaming, namingType string, err error) {
	formula := flag.String(formulaOption, "", `[mandatory] formula for naming files/folders. i.e: --formula="finance{increment} - {currentName} | output: finance1 - How to manage money"`)
	path := flag.String(pathOption, "./", "[optional] read path")
	isFileOnly := flag.Bool(isFileOnlyOption, false, "[optional] naming for files only")
	isFolderOnly := flag.Bool(isFolderOnlyOption, false, "[optional] naming for folders only")
	flag.Parse()
	if formula == nil || *formula == "" {
		err = fmt.Errorf("%s is Mandatory", formulaOption)
		return
	}
	formulaNaming = *formula
	pathDir = *path
	namingType = getNamingType(*isFileOnly, *isFolderOnly)
	return
}

func getNamingType(isFileOnly, isFolderOnly bool) string {
	if isFileOnly {
		return namingTypeFile
	}
	if isFolderOnly {
		return namingTypeFile
	}
	return namingTypeAll
}

func readAll(pathDir string) (fileNames, folderNames []string, err error) {
	var (
		files []os.FileInfo
		stat  os.FileInfo
	)
	if stat, err = os.Stat(pathDir); err != nil || !stat.IsDir() {
		err = fmt.Errorf("Wrong path - %v", err)
		return
	}

	if files, err = ioutil.ReadDir(pathDir); err != nil {
		return
	}
	for _, f := range files {
		if f.IsDir() {
			folderNames = append(folderNames, f.Name())
			continue
		}
		fileNames = append(fileNames, f.Name())
	}
	return
}

func rename(src, dst string) error {
	err := os.Rename(src, dst)
	if err != nil {
		return err
	}
	return nil
}

func readFormula(in, currentName string, index int) string {
	if strings.Contains(in, formulaIncrement) {
		in = strings.ReplaceAll(in, formulaIncrement, fmt.Sprint(index))
	}
	if strings.Contains(in, formulaCurrent) {
		in = strings.ReplaceAll(in, formulaCurrent, currentName)
	}
	return in
}
