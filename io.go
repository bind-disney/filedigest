package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"os"
)

func createInputReader() (*os.File, *bufio.Reader, error) {
	inputFilePath, err := expandInputFilePath()
	if err != nil {
		return nil, nil, err
	}

	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not read input file: %v", err)
	}

	inputFileReader := bufio.NewReaderSize(inputFile, blockSize)

	return inputFile, inputFileReader, nil
}

func createOutputWriter() (*os.File, *bufio.Writer, error) {
	outputFilePath, err := expandOutputFilePath()
	if err != nil {
		return nil, nil, err
	}

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not create output file: %v", err)
	}

	outputWriter := bufio.NewWriterSize(outputFile, md5.Size)

	return outputFile, outputWriter, nil
}
