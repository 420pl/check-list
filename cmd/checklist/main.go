package main

import (
	"fmt"
	"os"

	"github.com/420pl/check-list"
)

func main() {
	var fileName, saveFileName string
	if len(os.Args) < 2 {
		fmt.Println("File path argument required")
		return
	} else {
		fileName = os.Args[1]
		if len(os.Args) < 3 {
			saveFileName = fileName
		} else {
			saveFileName = os.Args[2]
		}
	}

	fixReport, err := checklist.CheckFile(fileName, saveFileName)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Following problems have been found and fixed:")
		for _, item := range fixReport {
			fmt.Println("- " + item)
		}
	}
}
