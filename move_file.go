package main

import (
	"io"
	"os"
	"path/filepath"
)

// moveFile moves a file from sourcePath to destinationPath, overwriting if necessary
func moveFile(sourcePath, destinationPath string) error {
	// Remove the destination file if it exists
	if _, err := os.Stat(destinationPath); err == nil {
		err = os.Remove(destinationPath)
		if err != nil {
			return err
		}
	}

	// Create the destination directory if it doesn't exist
	destDir := filepath.Dir(destinationPath)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		err = os.MkdirAll(destDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Open the source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create the destination file
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// Copy the file contents
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	// Remove the source file (optional)
	// err = os.Remove(sourcePath)
	// if err != nil {
	// 	return err
	// }

	return nil
}
