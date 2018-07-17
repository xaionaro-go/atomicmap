package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func checkErr(err error) {
	if err == nil {
		return
	}
	panic(err)
}

func averageF32(values ...float32) float32 {
	sum := float32(0)
	for _, value := range values {
		sum += value
	}
	return sum / float32(len(values))
}

func main() {
	if len(os.Args) < 2 {
		panic(`It's required to pass the file path as an argument`)
	}
	if len(os.Args) < 3 {
		panic(`It's required to pass keyType name (for example: "int")`)
	}
	if len(os.Args) < 4 {
		panic(`It's required to pass action name (for example: "Set")`)
	}

	filePath := os.Args[1]
	requiredKeyTypeName := os.Args[2]
	requiredActionName := os.Args[3]

	file, err := os.Open(filePath)
	checkErr(err)
	defer file.Close()

	r := csv.NewReader(file)

	rows, err := r.ReadAll()
	checkErr(err)

	seriesNamesMap := map[string]bool{}
	results := map[int]map[string][]float32{}

	for _, row := range rows {
		mapTypeName := row[1]
		actionName := row[2]
		keyTypeName := row[3]
		blockSizeStr := row[4]
		keysAmountStr := row[5]
		threadSafetyStr := row[6]
		opExecTimeStr := row[9]

		if keyTypeName != requiredKeyTypeName {
			continue
		}

		if actionName != requiredActionName {
			continue
		}

		seriesName := mapTypeName + "_" + actionName + "_" + keyTypeName + "_bs" + blockSizeStr + "_threadSafe" + threadSafetyStr
		seriesNamesMap[seriesName] = true

		keysAmount, err := strconv.Atoi(keysAmountStr)
		checkErr(err)

		if results[keysAmount] == nil {
			results[keysAmount] = map[string][]float32{}
		}

		opExecTime, err := strconv.ParseFloat(opExecTimeStr, 32)
		checkErr(err)

		results[keysAmount][seriesName] = append(results[keysAmount][seriesName], float32(opExecTime))
	}

	seriesNames := []string{}
	for seriesName := range seriesNamesMap {
		seriesNames = append(seriesNames, seriesName)
	}
	sort.Strings(seriesNames)

	keyAmounts := []int{}
	for keyAmount, _ := range results {
		keyAmounts = append(keyAmounts, keyAmount)
	}
	sort.Ints(keyAmounts)

	fmt.Println("," + strings.Join(seriesNames, ","))

	for _, keyAmount := range keyAmounts {
		fmt.Printf("%v", keyAmount)

		serieses := results[keyAmount]
		for _, seriesName := range seriesNames {
			values := serieses[seriesName]
			if len(values) == 0 {
				fmt.Printf(",")
				continue
			}

			fmt.Printf(",%.1f", averageF32(values...))
		}
		fmt.Printf("\n")
	}
}
