package dash

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/davimdo/gott"
	"github.com/davimdo/gott/dash/mpd"
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
func NewEngine() (*Engine, error) {
	return &Engine{
		ctx:  context.Background(),
		http: http.DefaultClient,
	}, nil
}

// NewEngineWithContext creates a new instance of Player with the given context.
func NewEngineWithContext(ctx context.Context, manifestURL string) (*Engine, error) {
	p, err := NewEngine()
	if err != nil {
		return nil, err
	}
	p.ctx = ctx
	return p, nil
}

func (p *Engine) LoadURL(manifest *url.URL) error {
	m, err := gott.Fetch(p.http, manifest)
	if err != nil {
		return err
	}
	return p.LoadManifest(manifest, m)
}

func (p *Engine) LoadManifest(url *url.URL, manifest []byte) error {
	p.manifestURL = url
	m, err := mpd.Unmarshal(manifest)
	if err != nil {
		return err
	}
	err = p.loadStreams(m)
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

func (p *Engine) loadStreams(m *mpd.MPD) error {
	period := m.Periods[0]
	for _, as := range period.AdaptationSets {
		streamType, err := getStreamType(&as)
		if err != nil {
			continue
		}

		chunks := as.SegmentTemplate.SegmentTimeline.S
		timescale := as.SegmentTemplate.Timescale
		url := as.SegmentTemplate.Media
		for _, representation := range as.Representations {
			var stream Stream
			stream.http = p.http
			stream.streamType = streamType
			stream.bitrate = representation.Bandwidth

			dts := uint64(0)
			chunkCount := 0
			for _, c := range chunks {
				if c.T != 0 {
					dts = c.T
				}
				for i := 0; i < int(c.R+1); i++ {
					relativeChunkURL, err := chunkURL(url, stream.bitrate, dts)
					absoluteChunkURL := p.manifestURL.ResolveReference(relativeChunkURL)
					if err != nil {
						return err
					}
					chunk := gott.Chunk{
						Index:    chunkCount,
						URL:      absoluteChunkURL,
						DTS:      dts,
						Duration: durationFromTimeScale(int64(c.D), int64(timescale)),
					}
					stream.chunks = append(stream.chunks, chunk)
					dts = dts + c.D
					chunkCount++
				}
			}
			p.streams = append(p.streams, stream)
		}
	}
	return nil
}

var errUnknownStreamType = errors.New("Imposible to extrapolate StreamType from given adaptation set")

// getStreamType return the stream type for the adaptation set provided. If an error
// is returned, no stream type could be obtain and that Adaptation set should be
// discarded.
func getStreamType(as *mpd.AdaptationSet) (gott.StreamType, error) {
	switch as.MimeType {
	case "video/mp4":
		if ep := as.EssentialProperty; ep != nil {
			if ep.SchemeIDURI == "http://dashif.org/guide-lines/trickmode" {
				return gott.StreamTypeTrickmode, nil
			}
		} else {
			return gott.StreamTypeVideo, nil
		}
	case "audio/mp4":
		return gott.StreamTypeAudio, nil
	case "application/mp4":
		role := as.Role
		if role == nil {
			return gott.StreamTypeText, errUnknownStreamType
		}
		if role.SchemeIDURI == "urn:mpeg:dash:role:2011" && role.Value == "subtitle" {
			return gott.StreamTypeText, nil
		}
		return gott.StreamTypeText, errUnknownStreamType
	default:
		return gott.StreamTypeText, errUnknownStreamType
	}
	return gott.StreamTypeText, errUnknownStreamType
}

func chunkURL(chunkURL string, bitrate uint64, t uint64) (*url.URL, error) {
	fragBitrate := strconv.Itoa(int(bitrate))
	fragTime := strconv.Itoa(int(t))
	replacedURL := strings.ReplaceAll(chunkURL, "$Bandwidth$", fragBitrate)
	replacedURL = strings.ReplaceAll(replacedURL, "$Time$", fragTime)
	return url.Parse(replacedURL)
}

func durationFromTimeScale(duration, timeScale int64) time.Duration {
	d := float64(duration)
	ts := float64(timeScale)
	s := float64(time.Second)
	return time.Duration((d / (ts / s)))
}
