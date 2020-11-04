package gott

import (
	"context"
	"net/http"
	"net/url"
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

func (st StreamType) String() string {
	switch st {
	case StreamTypeVideo:
		return "video"
	case StreamTypeTrickmode:
		return "trickmode"
	case StreamTypeAudio:
		return "audio"
	case StreamTypeText:
		return "text"
	default:
		return "unknown"
	}
}

type Chunk struct {
	Index    int
	URL      *url.URL
	DTS      uint64
	Duration time.Duration
}

func PlayStream(stream Stream) <-chan *http.Response {
	return PlayStreamWithContext(context.Background(), stream)
}

func PlayStreamWithContext(ctx context.Context, stream Stream) <-chan *http.Response {
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
