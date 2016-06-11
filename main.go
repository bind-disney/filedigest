package main

import (
	"bufio"
	"crypto/md5"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
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
	inputFileName    string
	outputFileName   string
	blockSize        int
	concurrencyLevel int
)

func init() {
	initFlags()

	concurrencyLevel = int(math.Min(float64(runtime.NumCPU()), float64(runtime.GOMAXPROCS(0))))
}

func main() {
	checkFlags()

	inputFile, inputReader, err := createInputReader()
	checkError(err)
	defer inputFile.Close()

	// Will be flushed and closed in writeDigestResults
	outputFile, outputWriter, err := createOutputWriter()
	checkError(err)

	// Will be closed manually after reading of the input file
	doneChannel := make(chan struct{})

	quitChannel := make(chan struct{})
	defer close(quitChannel)

	// Will be closed automatically in processDigestResponses
	resultChannel := make(chan *digestResponse, concurrencyLevel)

	responseChannel := make(chan *digestResponse)
	defer close(responseChannel)

	requestChannel := make(chan *digestRequest)
	defer close(requestChannel)

	requestId := uint64(0)

	for i := 0; i < concurrencyLevel; i++ {
		go handleDigestRequest(requestChannel, responseChannel, doneChannel, quitChannel)
	}

	go processDigestResponses(responseChannel, resultChannel)
	go writeDigestResults(resultChannel, outputFile, outputWriter)

	for {
		data := make([]byte, blockSize)
		bytesCount, err := inputReader.Read(data)

		if err != nil && err != io.EOF {
			log.Println(err)
			break
		}

		// End of file reached, exit loop and cleanup all resources
		if bytesCount == 0 {
			break
		}

		request := digestRequest{id: requestId, data: data}
		requestChannel <- &request
		requestId++
	}

	close(doneChannel)

	// Wait for all request handlers to be finished
	for i := 0; i < concurrencyLevel; i++ {
		<-quitChannel
	}
}

func handleDigestRequest(requestChannel <-chan *digestRequest, responseChannel chan<- *digestResponse, doneChannel <-chan struct{}, quitChannel chan<- struct{}) {
	for {
		select {
		case request := <-requestChannel:
			responseChannel <- &digestResponse{
				id:  request.id,
				sum: md5.Sum(request.data),
			}
		case <-doneChannel:
			quitChannel <- struct{}{}
			return
		}
	}
}

func processDigestResponses(responseChannel <-chan *digestResponse, resultChannel chan<- *digestResponse) {
	responses := make([]*digestResponse, 0, concurrencyLevel)
	defer close(resultChannel)

	flushResults := func() {
		if len(responses) == 0 {
			return
		}

		sort.Sort(byId(responses))

		for _, result := range responses {
			resultChannel <- result
		}

		responses = nil
	}

	for response := range responseChannel {
		responses = append(responses, response)

		if len(responses) == concurrencyLevel {
			flushResults()
		}
	}

	flushResults()
}

func writeDigestResults(resultChannel <-chan *digestResponse, file io.Closer, writer *bufio.Writer) {
	defer file.Close()
	defer writer.Flush()

	for result := range resultChannel {
		_, err := writer.Write(result.sum[:])
		checkError(err)
	}
}

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
	fmt.Fprintln(os.Stderr, "Usage:")
	flag.PrintDefaults()
}

func initFlags() {
	flag.StringVar(&inputFileName, "input-file", "", "Path to the input file")
	flag.StringVar(&outputFileName, "output-file", "", "Path to the output file")
	flag.IntVar(&blockSize, "block-size", defaultBlockSize, "Block size in bytes")
	flag.Usage = showUsage

	flag.Parse()
}

func checkFlags() {
	if flag.NFlag() < 2 {
		flag.Usage()
		os.Exit(1)
	}
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
	return collection[i] != nil && collection[j] != nil && collection[i].id < collection[j].id
}
