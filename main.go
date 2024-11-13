package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var ErrFileFound = errors.New("file found")

func main() {
	start := time.Now()

	csvPath := flag.String("csvPath", "", "Path to the CSV file")
	rootDirectory := flag.String("rootDirectory", "", "Root directory of folders to search through")
	maxGoroutines := flag.Int("maxGoroutines", 32, "Maximum number of concurrent goroutines")
	flag.Parse()

	csvPathValue := *csvPath
	rootDirectoryValue := *rootDirectory
	maxGoroutinesValue := *maxGoroutines

	csvFile, err := os.Open(csvPathValue)
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

	// Map the file structure
	fileMap := make(map[string]string)
	err = filepath.WalkDir(rootDirectoryValue, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			fileMap[strings.ToLower(d.Name())] = path
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Use a WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Create a buffered channel to limit the number of concurrent goroutines
	semaphore := make(chan struct{}, maxGoroutinesValue)

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

		semaphore <- struct{}{}

		// Increment the WaitGroup counter
		wg.Add(1)

		// Launch a goroutine for each record
		go func(fileName, destinationDirectory string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			// Look up the file in the map
			filePath, found := fileMap[strings.ToLower(fileName)]
			if found {

				destinationPath := filepath.Join(destinationDirectory, fileName)
				err = moveFile(filePath, destinationPath)
				if err != nil {
					log.Printf("Failed to move '%s': %v\n", fileName, err)
					return
				}
				fmt.Printf("Moved '%s' to '%s'\n", fileName, destinationDirectory)
			} else {
				fmt.Printf("File '%s' not found in '%s'\n", fileName, rootDirectoryValue)
			}
		}(fileName, destinationDirectory)
	}

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("Process completed in %s\n", elapsed)
}
