package main

import (
	"bufio"
	"crypto/md5"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

type (
	digestRequest struct {
		id   uint64
		data []byte
	}

	digestResponse struct {
		id  uint64
		sum [md5.Size]byte
	}

	byId []*digestResponse
)

const defaultBlockSize int = 1 << 20

var (
	inputFileName  string
	outputFileName string
	blockSize      int
	workersCount   int
	outputWriter   *bufio.Writer
)

func init() {
	initFlags()
	workersCount = runtime.NumCPU()
}

func main() {
	if flag.NFlag() < 2 {
		flag.Usage()
		os.Exit(1)
	}

	inputFilePath, err := expandInputFilePath()
	checkError(err)

	inputFile, err := os.Open(inputFilePath)
	checkError(err)
	defer inputFile.Close()

	outputFile, err := createOutputFile()
	checkError(err)
	outputWriter = bufio.NewWriter(outputFile)
	defer outputFile.Close()
	defer outputWriter.Flush()

	reader := bufio.NewReaderSize(inputFile, blockSize)
	requestChannel := make(chan *digestRequest, workersCount)
	doneChannel := make(chan struct{})
	defer close(doneChannel)
	requestId := uint64(0)

	go processRequests(requestChannel, doneChannel)

	for {
		data := make([]byte, blockSize)
		bytesCount, err := reader.Read(data)

		if err != nil && err != io.EOF {
			log.Fatal(err)
		}

		// End of file reached
		if bytesCount == 0 {
			break
		}

		request := digestRequest{id: requestId, data: data}
		requestChannel <- &request
		requestId++
	}
}

func processRequests(requestChannel <-chan *digestRequest, doneChannel <-chan struct{}) {
	responseChannel := make(chan *digestResponse)
	go processResponses(responseChannel, doneChannel)
	defer close(responseChannel)

	for {
		select {
		case request := <-requestChannel:
			go calculateDigest(request, responseChannel, doneChannel)
		case <-doneChannel:
			return
		}
	}
}

func processResponses(responseChannel <-chan *digestResponse, doneChannel <-chan struct{}) {
	responses := make([]*digestResponse, 0, workersCount)
	digestChannel := make(chan *digestResponse, workersCount)
	defer close(digestChannel)

	go writeDigests(outputWriter, digestChannel)

	for {
		select {
		case response := <-responseChannel:
			responses = append(responses, response)

			if len(responses) < workersCount {
				continue
			}

			sort.Sort(byId(responses))

			for _, r := range responses {
				digestChannel <- r
			}

			responses = nil
		case <-doneChannel:
			return
		}
	}
}

func calculateDigest(request *digestRequest, responseChannel chan<- *digestResponse, doneChannel <-chan struct{}) {
	select {
	case responseChannel <- &digestResponse{id: request.id, sum: md5.Sum(request.data)}:
	case <-doneChannel:
	}
}

func writeDigests(writer io.Writer, responseChannel <-chan *digestResponse) {
	for response := range responseChannel {
		_, err := writer.Write(response.sum[:])
		checkError(err)
	}
}

func createOutputFile() (*os.File, error) {
	outputFilePath, err := expandOutputFilePath()
	if err != nil {
		return nil, fmt.Errorf("Could not determine output file full path: %v", err)
	}

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return nil, fmt.Errorf("Could not create output file: %v", err)
	}

	return outputFile, nil
}

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

func showUsage() {
	fmt.Fprintln(os.Stderr, "Usage:\n")
	flag.PrintDefaults()
}

func initFlags() {
	flag.StringVar(&inputFileName, "input-file", "", "Path to the input file")
	flag.StringVar(&outputFileName, "output-file", "", "Path to the output file")
	flag.IntVar(&blockSize, "block-size", defaultBlockSize, "Block size in bytes")
	flag.Usage = showUsage

	flag.Parse()
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func (collection byId) Len() int {
	return len(collection)
}

func (collection byId) Swap(i, j int) {
	collection[i], collection[j] = collection[j], collection[i]
}

func (collection byId) Less(i, j int) bool {
	return collection[i].id < collection[j].id
}
