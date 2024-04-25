package utils

import (
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

func GetRandomTxtFilePath(dir string) (string, error) {
	// Get list of files in the directory
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	// Filter .txt files
	var txtFiles []string
	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".txt" {
			txtFiles = append(txtFiles, entry.Name())
		}
	}

	// Check if there are any .txt files
	if len(txtFiles) == 0 {
		return "", errors.New("no .txt files found in the directory")
	}

	// Create a new source for random number generation
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	// Generate a random index
	randomIndex := r.Intn(len(txtFiles))

	// Get the randomly selected file name
	randomFileName := txtFiles[randomIndex]

	// Construct the file path
	filePath := filepath.Join(dir, randomFileName)

	return filePath, nil
}
