package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var ErrFileFound = errors.New("file found")

func main() {
	csvPath := `C:\Users\mamores\Desktop\PowerShell_POC\mapping.csv`
	rootDirectory := `C:\Users\mamores\Desktop\PowerShell_POC\MainFolder`

	csvFile, err := os.Open(csvPath)
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(bufio.NewReader(csvFile))
	reader.FieldsPerRecord = -1

	// Read the header row to get column indexes
	headers, err := reader.Read()
	if err != nil {
		log.Fatal(err)
	}

	var fileNameIndex, directoryIndex int
	for i, h := range headers {
		switch strings.ToLower(h) {
		case "filename":
			fileNameIndex = i
		case "directory":
			directoryIndex = i
		}
	}

	// Process each row in the CSV file
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fileName := record[fileNameIndex]
		destinationDirectory := record[directoryIndex]

		// Search for the file in the root directory and its subdirectories
		var filePath string
		err = filepath.Walk(rootDirectory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.EqualFold(info.Name(), fileName) {
				filePath = path
				return ErrFileFound // Signal to stop walking
			}
			return nil
		})
		if err != nil && err != ErrFileFound {
			log.Fatal(err)
		}

		// If the file is found, move it to the destination directory
		if filePath != "" {
			destinationPath := filepath.Join(destinationDirectory, fileName)
			err = moveFile(filePath, destinationPath)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Moved '%s' to '%s'\n", fileName, destinationDirectory)
		} else {
			fmt.Printf("File '%s' not found in '%s'\n", fileName, rootDirectory)
		}
	}
}
