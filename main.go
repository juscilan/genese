package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	templateFilePath := "_template.txt"
	dataFilePath := "_list.txt"
	distFilePath := "_dist.txt"

	template, err := os.ReadFile(templateFilePath)
	if err != nil {
		fmt.Printf("Error reading template file: %s\n", err)
		return
	}
	templateStr := string(template)

	dataFile, err := os.Open(dataFilePath)
	if err != nil {
		fmt.Printf("Error opening list file: %s\n", err)
		return
	}
	defer dataFile.Close()

	distFile, err := os.Create(distFilePath)
	if err != nil {
		fmt.Printf("Error creating dist file: %s\n", err)
		return
	}
	defer distFile.Close()

	scanner := bufio.NewScanner(dataFile)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ";")

		processedData := templateStr
		for i, field := range fields {
			placeholder := fmt.Sprintf("%%%d", i+1)
			processedData = strings.ReplaceAll(processedData, placeholder, field)
		}

		regex := regexp.MustCompile("%[0-9]+")
		processedData = regex.ReplaceAllString(processedData, "")

		_, err := distFile.WriteString(processedData + "\n")
		if err != nil {
			fmt.Printf("Error writing to destination file: %s\n", err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading lines: %s\n", err)
	}

	fmt.Println("File has been written successfully !!! Grab a coffee ☕")
}
