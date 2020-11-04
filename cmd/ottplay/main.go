package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/davimdo/gott"
	"github.com/davimdo/gott/smooth"
)

var manifestURL string

func init() {
	flag.StringVar(&manifestURL, "i", "", "Manifest URL")
}

func main() {
	flag.Parse()
	if manifestURL == "" {
		flag.Usage()
		fmt.Printf("Err: manifest URL not specified.\n")
		os.Exit(-1)
	}

	engine, err := loadEngine(manifestURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	defaultStreams := defaultStreams(engine.Streams())
	err = playWithContext(engine.Context(), defaultStreams)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func loadEngine(manifestURL string) (gott.Engine, error) {
	u, err := url.Parse(manifestURL)
	if err != nil {
		return nil, err
	}
	manifest, err := gott.Fetch(http.DefaultClient, u)
	if err != nil {
		return nil, err
	}

	if manifestIsSmooth(manifest) {
		return loadSmoothEngine(u, manifest)
	}
	return nil, fmt.Errorf("Unknown Engine for manifest fetch from: %s", manifestURL)
}

func loadSmoothEngine(u *url.URL, manifest []byte) (gott.Engine, error) {
	engine, err := smooth.NewEngine()
	if err != nil {
		return nil, err
	}
	err = engine.LoadManifest(u, manifest)
	if err != nil {
		return nil, err
	}
	return gott.Engine(engine), err
}

func manifestIsSmooth(manifest []byte) bool {
	if manifest == nil {
		return false
	}
	return bytes.Contains(manifest, []byte("<SmoothStreamingMedia"))
}

func playWithContext(ctx context.Context, streams []gott.Stream) error {
	var wg sync.WaitGroup
	wg.Add(len(streams))
	for _, stream := range streams {
		go func(stream gott.Stream) {
			j := 0
			for chunkResp := range gott.PlayStreamWithContext(ctx, stream) {
				fmt.Printf("%d - %5s - %s\n", j, stream.StreamType(), chunkResp.Request.URL)
				j++
				io.Copy(ioutil.Discard, chunkResp.Body)
				chunkResp.Body.Close()
			}
			wg.Done()
		}(stream)
	}
	wg.Wait()
	return nil
}

func defaultStreams(streams []gott.Stream) []gott.Stream {
	var defaultStreams []gott.Stream
	video := defaultStreamsVideo(streams)
	if video != nil {
		defaultStreams = append(defaultStreams, *video)
	}
	audio := defaultStreamsAudio(streams)
	if audio != nil {
		defaultStreams = append(defaultStreams, *audio)
	}
	text := defaultStreamsText(streams)
	if text != nil {
		defaultStreams = append(defaultStreams, *text)
	}
	return defaultStreams
}

func defaultStreamsVideo(streams []gott.Stream) *gott.Stream {
	index := 0
	for i, stream := range streams {
		if stream.StreamType() == gott.StreamTypeVideo && stream.Bitrate() > streams[index].Bitrate() {
			index = i
		}
	}
	return &streams[index]
}

func defaultStreamsAudio(streams []gott.Stream) *gott.Stream {
	for i, stream := range streams {
		if stream.StreamType() == gott.StreamTypeAudio {
			return &streams[i]
		}
	}
	return nil
}

func defaultStreamsText(streams []gott.Stream) *gott.Stream {
	for i, stream := range streams {
		if stream.StreamType() == gott.StreamTypeText {
			return &streams[i]
		}
	}
	return nil
}
