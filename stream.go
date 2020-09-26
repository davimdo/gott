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
	Play(position int) <-chan *http.Response
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

// Play call PlayStream concurrently for each one of the stream passed by
// argument. Play stop the "playout" upon an error on any of the Streams
// If an error is raised, all concurrent playout are canceled.
//
// Play can only be called if PlayerState is PlayerStateLoaded. And when all
// streams are done, or an error is raised PlayerState is set to PlayerLoaded.
//
// streams must not be null.
func Play(streams []Stream, position int) error {
	var wg sync.WaitGroup
	wg.Add(len(streams))
	for i, stream := range streams {
		go func(stream Stream, i int) {
			j := 0
			for chunkResp := range stream.Play(position) {
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
