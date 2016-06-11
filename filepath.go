package main

import (
	"errors"
	"fmt"
	"path/filepath"
)

func expandInputFilePath() (string, error) {
	inputFilePath, err := expandFilePath("input", inputFileName)
	return inputFilePath, err
}

func expandOutputFilePath() (string, error) {
	outputFilePath, err := expandFilePath("output", outputFileName)
	return outputFilePath, err
}

func expandFilePath(prefix, fileName string) (string, error) {
	if fileName == "" {
		return "", errors.New(prefix + " file is not specified")
	}

	filePath, err := filepath.Abs(fileName)
	if err != nil {
		return "", fmt.Errorf("Could not get %s file full path: %v", prefix, err)
	}

	return filePath, nil
}
