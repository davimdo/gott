package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/davimdo/gott"
	"github.com/davimdo/gott/dash"
	"github.com/davimdo/gott/smooth"
)

var (
	playoutFile string
	verbose     bool
	interval    int
)

func init() {
	flag.StringVar(&playoutFile, "input", "", "playout url list file")
	flag.BoolVar(&verbose, "verbose", false, "verbose log")
	flag.IntVar(&interval, "interval", 1000, "Interval in ms to read a new playout from input file")

	client := http.DefaultClient
	client.Timeout = 10 * time.Second
	client.Transport = &http.Transport{
		Proxy: nil,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          4096,
		MaxIdleConnsPerHost:   1024,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func main() {
	flag.Parse()
	count := 1
	var wg sync.WaitGroup
	for p := range getPlayoutFromFile(playoutFile) {
		wg.Add(1)
		go func(p playout, count int) {
			for {
				fmt.Printf("%d - Stating playout: %#v\n", count, p)
				engine, err := loadEngine(p.URL)
				if err != nil {
					fmt.Println(err)
					break
				}

				defaultStreams := defaultStreams(engine.Streams())
				err = play(defaultStreams, p.Position)
				if err != nil {
					fmt.Println(err)
					break
				}
			}
			wg.Done()
		}(p, count)
		count++
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
	wg.Wait()
}

type playout struct {
	URL      string
	Position int
}

func getPlayoutFromFile(path string) <-chan playout {
	c := make(chan playout, 1)
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	go func(file *os.File) {
		defer file.Close()
		r := csv.NewReader(file)
		r.Comma = ';'
		r.Comment = '#'
		for {
			record, err := r.Read()
			if err == io.EOF {
				close(c)
				break
			}
			if err != nil {
				close(c)
				log.Fatal(err)
				break
			}
			position, err := strconv.Atoi(record[1])
			if err != nil {
				close(c)
				log.Fatal(err)
				break
			}
			p := playout{
				URL:      record[0],
				Position: position,
			}
			c <- p
		}
	}(file)
	return c
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
	if manifestIsDash(manifest) {
		return loadDashEngine(u, manifest)
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

func loadDashEngine(u *url.URL, manifest []byte) (gott.Engine, error) {
	engine, err := dash.NewEngine()
	if err != nil {
		return nil, err
	}
	err = engine.LoadManifest(u, manifest)
	if err != nil {
		return nil, err
	}
	return gott.Engine(engine), err
}

func manifestIsDash(manifest []byte) bool {
	if manifest == nil {
		return false
	}
	return bytes.Contains(manifest, []byte("<MPD"))
}

func play(streams []gott.Stream, position int) error {
	var wg sync.WaitGroup
	wg.Add(len(streams))
	for _, stream := range streams {
		go func(stream gott.Stream) {
			j := 0
			for resp := range gott.PlayStream(stream, position, true) {
				if resp.StatusCode != 200 || verbose {
					fmt.Printf("%d - %5s - [GET %d] %s %d\n", j+position, stream.StreamType(), resp.StatusCode, resp.Request.URL, resp.ContentLength)
				}
				j++
				io.Copy(ioutil.Discard, resp.Body)
				resp.Body.Close()
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
