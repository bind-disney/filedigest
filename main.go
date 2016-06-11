package main

import (
	"crypto/md5"
	"io"
	"log"
	"math"
	"runtime"
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

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
