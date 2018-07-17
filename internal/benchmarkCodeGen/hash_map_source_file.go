package main

import (
	"bufio"
	"fmt"
	"os"
	"text/template"
)

var (
	benchmarkActionNames = []string{"Set" /*"ReSet", */, "Get" /*"GetMiss", */, "Unset" /*"UnsetMiss"*/}
	blockSizes           = []int{16, 64, 128, 1024, 65536, 4 * 1024 * 1024, 16 * 1024 * 1024}
	keyAmounts           = []int{16, 512, 65536, 1024 * 1024}
	keyTypes             = []string{"int", "string", /*"slice", "map", "struct"*/}
	threadSafeties       = []bool{false}
)

type hashMapSourceFile struct {
	Name        string
	PackageName string
}

type hashMapSourceFiles []hashMapSourceFile

func (files hashMapSourceFiles) GenerateTestFiles() error {
	for _, file := range files {
		err := file.GenerateTestFile()
		if err != nil {
			return err
		}
	}

	return nil
}

func (file hashMapSourceFile) GenerateTestFile() error {
	if len(file.Name) <= len(".go") { // The file should have something before ".go" in it's name
		return fmt.Errorf(`len(file.Name) < 4: file.Name == "%v"`, file.Name)
	}

	// Open the file

	data := make(map[string]interface{})
	data["PackageName"] = file.PackageName

	outFileName := file.Name[:len(file.Name)-3] + "_test.go" // Replacing ".go" by "_test.go" on the end (in the file name): myMap.go -> myMap_test.go

	outFile, err := os.Create(outFileName)
	if err != nil {
		return err
	}
	defer outFile.Close()

	outFileWriter := bufio.NewWriter(outFile)
	defer outFileWriter.Flush()

	// Parse the template to be used

	tpl, err := template.New("benchmarksFileTemplate").Parse(benchmarksFileTemplate)
	if err != nil {
		return err
	}

	// Write the file header

	err = tpl.ExecuteTemplate(outFileWriter, "header", data)
	if err != nil {
		return err
	}

	// Write the test function

	if file.PackageName != "stupidMap" {
		err = tpl.ExecuteTemplate(outFileWriter, "testFunction", data)
		if err != nil {
			return err
		}
	}
	if file.PackageName == "openAddressGrowingMap" {
		err = tpl.ExecuteTemplate(outFileWriter, "testCollisionsFunction", data)
		if err != nil {
			return err
		}
		err = tpl.ExecuteTemplate(outFileWriter, "testConcurrencyFunction", data)
		if err != nil {
			return err
		}
	}

	// Write the benchmark functions

	keyTypesFixed := keyTypes
	blockSizesFixed := blockSizes
	switch file.PackageName {
	case "stupidMap":
		blockSizesFixed = []int{0}
		keyTypesFixed = []string{"int", "string"}
	case "cgoTsearch":
		blockSizesFixed = []int{1024 * 1024}
		keyTypesFixed = []string{"int"}
	case "openAddressGrowingMap":
		threadSafeties = []bool{false, true}
	}

	for _, actionName := range benchmarkActionNames {
		data["Action"] = actionName
		for _, blockSize := range blockSizesFixed {
			data["BlockSize"] = blockSize
			for _, keyAmount := range keyAmounts {
				if keyAmount*1024 < blockSize {
					continue
				}
				data["KeyAmount"] = keyAmount
				for _, keyType := range keyTypesFixed {
					data["KeyType"] = keyType
					for _, threadSafety := range threadSafeties {
						data["ThreadSafety"] = threadSafety
						err = tpl.ExecuteTemplate(outFileWriter, "benchmarkFunction", data)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}
