// Parse files with the comment "//go:generate benchmarkCodeGen"

package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

var (
	magicGoGenerateCommentRegexp = regexp.MustCompile(`go:generate ([0-9A-Za-z_\.]+)`)
)

const (
	goGenerateExpectedValue = "benchmarkCodeGen"
)

func parseGoFile(path string) (*ast.File, error) {
	fset := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	checkErr(err)

	return parsedFile, err
}

// parseGoGenerateComment returns words passed right after "//go:generate"
// comments.
func parseGoGenerateComments(goFile *ast.File) (result []string) {
	commentGroups := goFile.Comments

	for _, commentGroup := range commentGroups {
		for _, comment := range commentGroup.List {
			goGenerateMatches := magicGoGenerateCommentRegexp.FindStringSubmatch(comment.Text)
			if len(goGenerateMatches) < 2 {
				continue
			}
			result = append(result, goGenerateMatches[1:]...)
		}
	}
	return

}

func hasExpectedGoGenerateValue(goFile *ast.File) bool {
	for _, goGenerateArgument := range parseGoGenerateComments(goFile) {
		if goGenerateArgument == goGenerateExpectedValue {
			return true
		}
	}
	return false
}

func tryParseHashMapSourceFile(dirPath, fileName string) (*hashMapSourceFile, error) {
	goFile, err := parseGoFile(path.Join(dirPath, fileName))
	if err != nil {
		return nil, err
	}

	if !hasExpectedGoGenerateValue(goFile) {
		return nil, nil
	}

	return &hashMapSourceFile{PackageName: goFile.Name.Name, Name: fileName}, nil
}

func parseHashMapSourceFiles(dirPath string) (result hashMapSourceFiles) {
	fileName := os.Getenv("GOFILE")
	if fileName != "" {
		hashMapSourceFilePtr, err := tryParseHashMapSourceFile(dirPath, fileName)
		if err != nil {
			log.Println(err)
		}
		if hashMapSourceFilePtr != nil {
			result = append(result, *hashMapSourceFilePtr)
		}
		return
	}

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Println(err)
		return
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		if !strings.HasSuffix(fileName, ".go") {
			continue
		}

		if strings.HasSuffix(fileName, "_test.go") {
			continue
		}

		hashMapSourceFilePtr, err := tryParseHashMapSourceFile(dirPath, fileName)
		if err != nil {
			log.Println(err)
		}
		if hashMapSourceFilePtr != nil {
			result = append(result, *hashMapSourceFilePtr)
		}
	}

	return
}
