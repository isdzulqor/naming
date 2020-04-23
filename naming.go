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
	formulaOption            = "formula"
	ignoreOption             = "ignore"
	pathOption               = "path"
	rollbackOption           = "rollback"
	startOption              = "start"
	isFileOnlyOption         = "fileOnly"
	isFolderOnlyOption       = "folderOnly"
	isListAllOption          = "listAll"
	isSkipErrorOption        = "skipError"
	formattedIncrementOption = "formatIncrement"

	namingTypeFile   = "file_only"
	namingTypeFolder = "folder_only"
	namingTypeAll    = "file_and_folder"

	formulaIncrement          = "{increment}"
	formulaCurrent            = "{current}"
	defaultIncrement          = 1
	defaultFormattedIncrement = 4
)

var formulas = []string{formulaIncrement, formulaCurrent}

func main() {
	fmt.Println("Naming your files & folders quickly\n")
	var increment int

	path, formula, namingType, isListAll, isRollback, isSkipError, ignoreList, startIncrement, formattedIncrement, err := readArgs()
	if err != nil {
		log.Fatal(err)
	}
	if isListAll != nil && *isListAll {
		fmt.Println("List All files & folders in tree")
		if err = list(path); err != nil {
			log.Fatal(err)
		}
		return
	}

	if isRollback != nil && *isRollback {
		fmt.Println("Rollback files")
		increment, err = execRollback(formula, path, startIncrement, namingType, ignoreList)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Successfully rollback %d items for %s", increment-1, namingType)
		os.Exit(0)
		return
	}

	increment, err = exec(formula, path, startIncrement, namingType, ignoreList, *isSkipError, formattedIncrement)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Successfully naming %d items for %s", increment-1, namingType)
	os.Exit(0)
}

func exec(formula, path string, increment int, namingType string, ignoreList []string, isSkipError bool, formattedIncrement int) (int, error) {
	files, folders, err := readAll(path)
	if err != nil {
		return 0, err
	}
	prefixPath := path + "/"

	if namingType == namingTypeFile || namingType == namingTypeAll {
		if increment, err = naming(files, increment, formula, path, ignoreList, isSkipError, formattedIncrement); err != nil {
			return 0, err
		}
	}

	if namingType == namingTypeFolder || namingType == namingTypeAll {
		if increment, err = naming(folders, increment, formula, path, ignoreList, isSkipError, formattedIncrement); err != nil {
			return 0, err
		}
		if _, folders, err = readAll(path); err != nil {
			return 0, err
		}
	}

	for _, folder := range folders {
		if increment, err = exec(formula, prefixPath+folder, increment, namingType, ignoreList, isSkipError, formattedIncrement); err != nil {
			return 0, err
		}
	}

	return increment, nil
}

func execRollback(formula, path string, increment int, namingType string, ignoreList []string) (int, error) {
	files, folders, err := readAll(path)
	if err != nil {
		return 0, err
	}
	prefixPath := path + "/"

	if namingType == namingTypeFile || namingType == namingTypeAll {
		if increment, err = rollback(files, increment, formula, path, ignoreList); err != nil {
			return 0, err
		}
	}

	if namingType == namingTypeFolder || namingType == namingTypeAll {
		if increment, err = rollback(folders, increment, formula, path, ignoreList); err != nil {
			return 0, err
		}
		if _, folders, err = readAll(path); err != nil {
			return 0, err
		}
	}

	for _, folder := range folders {
		if increment, err = execRollback(formula, prefixPath+folder, increment, namingType, ignoreList); err != nil {
			return 0, err
		}
	}

	return increment, nil
}

func list(path string) (err error) {
	var files, folders []string
	if files, folders, err = readAll(path); err != nil {
		return
	}
	prefixPath := path + "/"
	for _, file := range files {
		fmt.Println(prefixPath + file)
	}
	for _, folder := range folders {
		if err = list(prefixPath + folder); err != nil {
			return
		}
	}
	return
}

func naming(src []string, increment int, formula, path string, ignoreList []string, isSkipError bool, formattedIncrement int) (lastIncrement int, err error) {
	prefixPath := path + "/"
	for _, s := range src {
		destName := readFormula(formula, s, increment, formattedIncrement)
		if okToRename(s, ignoreList) {
			if err = rename(prefixPath+s, prefixPath+destName); err == nil {
				increment++
			}
			if err != nil && !isSkipError {
				return
			}
		}
	}
	lastIncrement = increment
	return
}

func rollback(src []string, increment int, formula, path string, ignoreList []string) (lastIncrement int, err error) {
	prefixPath := path + "/"
	for _, s := range src {
		destName := readFormulaRollback(formula, s, increment)
		if okToRename(s, ignoreList) {
			if err = rename(prefixPath+s, prefixPath+destName); err != nil {
				return
			}
			increment++
		}
	}
	lastIncrement = increment
	return
}

func readArgs() (pathDir, formulaNaming, namingType string, isListAll, isRollback, isSkipError *bool, ignoreList []string, startIncrement, formattedIncrement int, err error) {
	formula := flag.String(formulaOption, "", `[mandatory] formula for naming files/folders. i.e: --formula="A{slash}B{slash}finance{increment} - {current} | output: A/B/finance1 - How to manage money"`)
	path := flag.String(pathOption, ".", "[optional] read path")
	start := flag.Int(startOption, defaultIncrement, "[optional] start increment value")
	formattedIncr := flag.Int(formattedIncrementOption, defaultFormattedIncrement, "[optional] increment format. i.e: 4. it will print increment like this format 0001, 0002, 0003")
	isRollback = flag.Bool(rollbackOption, false, "[optional] for rollback only")
	isFileOnly := flag.Bool(isFileOnlyOption, false, "[optional] naming for files only")
	isFolderOnly := flag.Bool(isFolderOnlyOption, false, "[optional] naming for folders only")
	isListAll = flag.Bool(isListAllOption, false, "[optional] for listing all files and folders")
	isSkipError = flag.Bool(isSkipErrorOption, false, "[optional] ignore error whenever it's failed to rename one of the file")
	ignores := flag.String(ignoreOption, "", `[optional] ignore list, comma-separated. i.e --ignore="Desktop.ini,*.exe" | * means all`)

	flag.Parse()
	ignoreList = strings.Split(*ignores, ",")

	if len(ignoreList) == 0 && *ignores != "" {
		ignoreList = append(ignoreList, *ignores)
	}
	pathDir = *path
	namingType = getNamingType(*isFileOnly, *isFolderOnly)

	if isListAll != nil && *isListAll {
		return
	}

	if formula == nil || *formula == "" {
		err = fmt.Errorf("%s is Mandatory", formulaOption)
		return
	}
	formulaNaming = *formula
	startIncrement = *start
	formattedIncrement = *formattedIncr
	return
}

func getNamingType(isFileOnly, isFolderOnly bool) string {
	if isFileOnly {
		return namingTypeFile
	}
	if isFolderOnly {
		return namingTypeFolder
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

// format i.e: 4
func getFormattedInt(format int, val int) string {
	stringFormat := "%" + fmt.Sprint(format) + "d"
	out := fmt.Sprintf(stringFormat, val)
	out = strings.ReplaceAll(out, " ", "0")
	return out
}

func readFormula(in, currentName string, index, formattedIncrement int) string {
	if strings.Contains(in, formulaIncrement) {
		in = strings.ReplaceAll(in, formulaIncrement, getFormattedInt(formattedIncrement, index))
	}
	if strings.Contains(in, formulaCurrent) {
		in = strings.ReplaceAll(in, formulaCurrent, currentName)
	}
	return in
}

func readFormulaRollback(formula, currentName string, index int) string {
	increment := getIncrementString(formula, currentName)
	formula = strings.ReplaceAll(formula, formulaIncrement, increment)
	realFilename := getRealFilename(formula, currentName)
	return realFilename
}

func getIncrementString(formula, filename string) string {
	if !strings.Contains(formula, formulaIncrement) {
		return ""
	}
	formula = strings.Replace(formula, formulaCurrent, "", -1)
	return stringBetween(filename, stringBefore(formula, formulaIncrement), stringAfter(formula, formulaIncrement))
}

func getRealFilename(formula, filename string) string {
	if !strings.Contains(formula, formulaCurrent) {
		return ""
	}
	formula = strings.Replace(formula, formulaIncrement, "", -1)
	before := stringBefore(formula, formulaCurrent)
	after := stringAfter(formula, formulaCurrent)
	filename = strings.Replace(filename, before, "", 1)
	filename = strings.Replace(filename, after, "", 1)
	return filename
}

func stringSliceContains(in []string, v string) bool {
	for _, s := range in {
		if s == v {
			return true
		}
	}
	return false
}

func okToRename(v string, ignoreList []string) bool {
	if stringSliceContains(ignoreList, v) {
		return false
	}
	for _, ignore := range ignoreList {
		if strings.Contains(ignore, "*.") {
			splitted := strings.Split(ignore, "*.")
			ignoreExtension := splitted[len(splitted)-1]
			if extension, err := getFileExtension(v); err == nil {
				if extension == ignoreExtension {
					return false
				}
			}
		}
	}
	return true
}

func getFileExtension(fileName string) (string, error) {
	splitted := strings.Split(fileName, ".")
	if len(splitted) > 0 {
		extension := splitted[len(splitted)-1]
		return extension, nil
	}
	return "", fmt.Errorf("No file extension!")
}

func stringBetween(value string, a string, b string) string {
	if !strings.Contains(value, a) || !strings.Contains(value, b) {
		return ""
	}
	value = strings.Replace(value, a, "", 1)
	return stringBefore(value, b)
}

func stringAfter(value string, a string) string {
	pos := strings.LastIndex(value, a)
	if pos == -1 {
		return ""
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return ""
	}
	return value[adjustedPos:len(value)]
}

func stringBefore(value string, a string) (out string) {
	slices := strings.Split(value, a)
	if len(slices) < 2 {
		return
	}
	return slices[0]
}
