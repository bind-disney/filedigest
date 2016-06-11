package main

import (
	"bufio"
	"crypto/md5"
	"io"
	"sort"
)

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
