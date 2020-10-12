package smooth

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/davimdo/gott"
	"github.com/davimdo/gott/smooth/ism"
)

// Engine implements gottPlayer interface for Microsoft Smooth Streaming
// protocol.
type Engine struct {
	manifestURL *url.URL
	ctx         context.Context
	http        *http.Client
	streams     []Stream
}

// NewEngine creates a new PLayer instance.
func NewEngine(manifestURL string) (*Engine, error) {
	u, err := url.Parse(manifestURL)
	if err != nil {
		return nil, err
	}
	return &Engine{
		ctx:         context.Background(),
		manifestURL: u,
		http:        http.DefaultClient,
	}, nil
}

// NewEngineWithContext creates a new instance of Player with the given context.
func NewEngineWithContext(ctx context.Context, manifestURL string) (*Engine, error) {
	p, err := NewEngine(manifestURL)
	if err != nil {
		return nil, err
	}
	p.ctx = ctx
	return p, nil
}

// Load obtains all the streams available in the manifest.
func (p *Engine) Load() error {
	manifest, err := p.getManifest()
	if err != nil {
		return err
	}
	err = p.loadStreams(manifest)
	if err != nil {
		return err
	}
	return nil
}

// Streams returns all available streams loaded by the player
func (p *Engine) Streams() []gott.Stream {
	streams := make([]gott.Stream, len(p.streams))
	for i := range p.streams {
		streams[i] = gott.Stream(&p.streams[i])
	}
	return streams
}

// Context returns player context. Player have always a context defined, so
// Context ensure none nil return.
func (p *Engine) Context() context.Context {
	if p.ctx == nil {
		return context.Background()
	}
	return p.ctx
}

func (p *Engine) getManifest() (*ism.SmoothStreamingMedia, error) {
	resp, err := p.http.Get(p.manifestURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	manifest, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return ism.Unmarshal(manifest)
}

func (p *Engine) loadStreams(m *ism.SmoothStreamingMedia) error {
	for _, si := range m.StreamIndexes {
		var streamType gott.StreamType
		switch strings.ToLower(si.Type) {
		case "video":
			streamType = gott.StreamTypeVideo
		case "audio":
			streamType = gott.StreamTypeAudio
		case "text":
			streamType = gott.StreamTypeText
		}
		for _, track := range si.Tracks {
			var stream Stream
			stream.http = p.http
			stream.streamType = streamType
			stream.bitrate = track.Bitrate
			dts := uint64(0)
			for i, c := range si.Chunks {
				if c.T != 0 {
					dts = c.T
				}
				relativeChunkURL, err := chunkURL(si.URL, track.Bitrate, dts)
				absoluteChunkURL := p.manifestURL.ResolveReference(relativeChunkURL)
				if err != nil {
					return err
				}
				chunk := gott.Chunk{
					Index:    i,
					URL:      absoluteChunkURL,
					DTS:      dts,
					Duration: durationFromTimeScale(int64(c.D), int64(m.TimeScale)),
				}
				stream.chunks = append(stream.chunks, chunk)
				dts = dts + c.D
			}
			p.streams = append(p.streams, stream)
		}
	}
	return nil
}

func chunkURL(chunkURL string, bitrate uint64, t uint64) (*url.URL, error) {
	fragBitrate := strconv.Itoa(int(bitrate))
	fragTime := strconv.Itoa(int(t))
	replacedURL := strings.ReplaceAll(chunkURL, "{bitrate}", fragBitrate)
	replacedURL = strings.ReplaceAll(replacedURL, "{Bitrate}", fragBitrate)
	replacedURL = strings.ReplaceAll(replacedURL, "{start time}", fragTime)
	replacedURL = strings.ReplaceAll(replacedURL, "{start_time}", fragTime)
	return url.Parse(replacedURL)
}

func durationFromTimeScale(duration, timeScale int64) time.Duration {
	d := float64(duration)
	ts := float64(timeScale)
	s := float64(time.Second)
	return time.Duration((d / (ts / s)))
}
