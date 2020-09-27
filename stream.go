package gott

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Stream interface {
	StreamType() StreamType
	Bitrate() uint64
	Chunks() []Chunk
	HttpClient() *http.Client
}

type StreamType int

const (
	StreamTypeVideo StreamType = iota
	StreamTypeTrickmode
	StreamTypeAudio
	StreamTypeText
)

type Chunk struct {
	Index    int
	URL      *url.URL
	DTS      uint64
	Duration time.Duration
}

func PlayStream(stream Stream) <-chan *http.Response {
	httpClient := stream.HttpClient()
	chuncks := stream.Chunks()

	c := make(chan *http.Response, 1)
	go func(c chan *http.Response) {
		firstChunkTime := time.Now()
		acumulateDuration := time.Duration(0)
		for _, chunk := range chuncks {
			resp, err := httpClient.Get(chunk.URL.String())
			if err != nil {
				close(c)
				return
			}
			c <- resp
			acumulateDuration += chunk.Duration
			waitFor := acumulateDuration - time.Since(firstChunkTime)
			time.Sleep(waitFor)
		}
		close(c)
	}(c)
	return c
}

// Play call PlayStream concurrently for each one of the stream passed by
// argument. Play stop the "playout" upon an error on any of the Streams
// If an error is raised, all concurrent playout are canceled.
//
// Play can only be called if PlayerState is PlayerStateLoaded. And when all
// streams are done, or an error is raised PlayerState is set to PlayerLoaded.
//
// streams must not be null.
func Play(streams []Stream) error {
	var wg sync.WaitGroup
	wg.Add(len(streams))
	for i, stream := range streams {
		go func(stream Stream, i int) {
			j := 0
			for chunkResp := range PlayStream(stream) {
				fmt.Printf("%d - %d - %s\n", j, i, chunkResp.Request.URL)
				j++
				io.Copy(ioutil.Discard, chunkResp.Body)
				chunkResp.Body.Close()
			}
			wg.Done()
		}(stream, i)
	}
	wg.Wait()
	return nil
}
