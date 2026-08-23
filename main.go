package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// processLine replaces the placeholders (%1, %2, ...) in the template
// with the fields of a ";"-separated CSV line and removes any unused
// placeholders.
func processLine(template, line string) string {
	fields := strings.Split(line, ";")

	result := template
	for i, field := range fields {
		placeholder := fmt.Sprintf("%%%d", i+1)
		result = strings.ReplaceAll(result, placeholder, field)
	}

	regex := regexp.MustCompile(`%[0-9]+`)
	result = regex.ReplaceAllString(result, "")

	return result
}

// generate reads lines from reader, processes each one with the template,
// and writes the result to writer.
func generate(templateStr string, reader io.Reader, writer io.Writer) error {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		processed := processLine(templateStr, line)

		_, err := fmt.Fprintln(writer, processed)
		if err != nil {
			return fmt.Errorf("error writing to destination file: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading lines: %w", err)
	}

	return nil
}

// run orchestrates reading the template, opening the input and output files,
// and calling generate.
func run(templatePath, dataPath, distPath string) error {
	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("error reading template file: %w", err)
	}

	dataFile, err := os.Open(dataPath)
	if err != nil {
		return fmt.Errorf("error opening list file: %w", err)
	}
	defer dataFile.Close()

	distFile, err := os.Create(distPath)
	if err != nil {
		return fmt.Errorf("error creating dist file: %w", err)
	}
	defer distFile.Close()

	return generate(string(templateBytes), dataFile, distFile)
}

func main() {
	if err := run("_template.txt", "_list.txt", "_dist.txt"); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("File has been written successfully !!! Grab a coffee ☕")
}
