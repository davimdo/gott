package main

import (
	"fmt"
	"log"
	"os"

	"github.com/davimdo/gott"
	"github.com/davimdo/gott/smooth"
)

// USAGE is the CMD help
const USAGE string = `usage: %s MANIFEST

	MANIFEST   OTT Manifest URL (for now Smooth)
`

func main() {
	if len(os.Args) != 2 {
		fmt.Printf(USAGE, os.Args[0])
		os.Exit(-1)
	}
	smoothEngine, err := smooth.NewEngine(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	err = smoothEngine.Load()
	if err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			log.Fatalln(err)
		}
		os.Exit(-1)
	}
	defaultStreams := defaultStreams(smoothEngine.Streams())
	err = gott.Play(defaultStreams)
	if err != nil {
		log.Fatalln(err)
	}
}

func defaultStreams(streams []gott.Stream) []gott.Stream {
	var defaultStreams []gott.Stream
	video := defaultStreamsVideo(streams)
	if video != nil {
		defaultStreams = append(defaultStreams, video)
	}
	audio := defaultStreamsAudio(streams)
	if video != nil {
		defaultStreams = append(defaultStreams, audio)
	}
	text := defaultStreamsText(streams)
	if video != nil {
		defaultStreams = append(defaultStreams, text)
	}
	return defaultStreams
}

func defaultStreamsVideo(streams []gott.Stream) gott.Stream {
	index := 0
	for i, stream := range streams {
		if stream.StreamType() == gott.StreamTypeVideo && stream.Bitrate() > streams[index].Bitrate() {
			index = i
		}
	}
	return streams[index]
}

func defaultStreamsAudio(streams []gott.Stream) gott.Stream {
	for i, stream := range streams {
		if stream.StreamType() == gott.StreamTypeAudio {
			return streams[i]
		}
	}
	return nil
}

func defaultStreamsText(streams []gott.Stream) gott.Stream {
	for i, stream := range streams {
		if stream.StreamType() == gott.StreamTypeText {
			return streams[i]
		}
	}
	return nil
}
