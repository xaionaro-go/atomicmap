// Generate "_test.go" files for HashMaper implementations
//
// The implementations should have comment "//go:generate genHashMaperBenchmarks"

package main

import (
	"flag"
	"log"
)

func main() {
	var dirPath string

	flag.Parse()

	if flag.NArg() == 0 {
		dirPath = "."
	} else {
		dirPath = flag.Arg(0)
	}

	files := parseHashMapSourceFiles(dirPath) // get informations about files contains "//go:generate benchmarkCodeGen"

	if len(files) != 1 {
		log.Panicf(`There found %v files with "//go:generate benchmarkCodeGen" comments. This case is not implemented (there should be only one file).`, len(files))
	}

	err := files[0].GenerateTestFile()
	checkErr(err)
}
