/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
package smooth

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/davimdo/gott"
	"github.com/davimdo/gott/pkg/ism"
)

// Player implements gottPlayer interface for Microsoft Smooth Streaming
// protocol.
type Player struct {
	manifestURL *url.URL
	ctx         context.Context
	http        *http.Client
	streams     []Stream
	state       gott.State
}

// NewPlayer creates a new PLayer instance.
func NewPlayer(manifestURL string) (*Player, error) {
	u, err := url.Parse(manifestURL)
	if err != nil {
		return nil, err
	}
	return &Player{
		ctx:         context.Background(),
		manifestURL: u,
		state:       gott.StateNotLoaded,
		http:        http.DefaultClient,
	}, nil
}

// NewPlayerWithContext creates a new instance of Player with the given context.
func NewPlayerWithContext(ctx context.Context, manifestURL string) (*Player, error) {
	p, err := NewPlayer(manifestURL)
	if err != nil {
		return nil, err
	}
	p.ctx = ctx
	return p, nil
}

// State gets Player state
func (p *Player) State() gott.State {
	return p.state
}

// Load obtains all the streams available in the manifest.
func (p *Player) Load() error {
	if p.state != gott.StateNotLoaded {
		return fmt.Errorf("load can not be call when player state is diferent than StateNotLoaded")
	}
	manifest, err := p.getManifest()
	if err != nil {
		return err
	}
	err = p.loadStreams(manifest)
	if err != nil {
		return err
	}
	p.state = gott.StateLoaded
	return nil
}

// Play call PlayStream concurrently for each one of the stream passed by
// argument. Play stop the "playout" upon an error on any of the Streams
// If an error is raised, all concurrent playout are canceled.
//
// Play can only be called if PlayerState is PlayerStateLoaded. And when all
// streams are done, or an error is raised PlayerState is set to PlayerLoaded.
//
// streams must not be null.
func (p *Player) Play(streams []gott.Stream, position int) error {
	if p.state != gott.StateLoaded {
		return fmt.Errorf("play can not be call when player state is diferent than StateLoaded")
	}
	p.state = gott.StatePlaying
	err := gott.Play(streams, position)
	if err != nil {
		p.state = gott.StateNotLoaded
		return nil
	}
	p.state = gott.StateNotLoaded
	return nil
}

// Stop cancel all stream playouts by canceling the player context.
func (p *Player) Stop() error {
	return nil
}

// Streams returns all available streams loaded by the player
func (p *Player) Streams() []gott.Stream {
	streams := make([]gott.Stream, len(p.streams))
	for i := range p.streams {
		streams[i] = gott.Stream(&p.streams[i])
	}
	return streams
}

// DefaultStreams returns from loaded streams one default stream for each StreamType.
// For video streams returns the ones with the maximum bitrate.
// For audio and text streams returns the first one.
func (p *Player) DefaultStreams() []gott.Stream {
	var streams []gott.Stream
	video := p.defaultStreamsVideo()
	if video != nil {
		streams = append(streams, gott.Stream(video))
	}
	audio := p.defaultStreamsAudio()
	if video != nil {
		streams = append(streams, gott.Stream(audio))
	}
	text := p.defaultStreamsText()
	if video != nil {
		streams = append(streams, gott.Stream(text))
	}
	return streams
}

// Context returns player context. Player have always a context defined, so
// Context ensure none nil return.
func (p *Player) Context() context.Context {
	if p.ctx == nil {
		return context.Background()
	}
	return p.ctx
}

// WithContext sets player context.
func (p *Player) WithContext(ctx context.Context) {
	p.ctx = ctx
}

func (p *Player) getManifest() (*ism.SmoothStreamingMedia, error) {
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

func (p *Player) loadStreams(m *ism.SmoothStreamingMedia) error {
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

type Stream struct {
	streamType gott.StreamType
	bitrate    uint64
	chunks     []gott.Chunk
	http       *http.Client
}

func (s *Stream) StreamType() gott.StreamType {
	return s.streamType
}

func (s *Stream) Bitrate() uint64 {
	return s.bitrate
}

func (s *Stream) Chunks() []gott.Chunk {
	return s.chunks
}

func (s *Stream) Play(position int) <-chan *http.Response {
	c := make(chan *http.Response, 1)
	go func(c chan *http.Response) {
		firstChunkTime := time.Now()
		acumulateDuration := time.Duration(0)
		for _, chunk := range s.chunks {
			if chunk.Index < position && position < len(s.chunks) {
				continue
			}
			resp, err := s.http.Get(chunk.URL.String())
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

func chunkURL(chunkURL string, bitrate uint64, t uint64) (*url.URL, error) {
	fragBitrate := strconv.Itoa(int(bitrate))
	fragTime := strconv.Itoa(int(t))
	replacedURL := strings.ReplaceAll(chunkURL, "{bitrate}", fragBitrate)
	replacedURL = strings.ReplaceAll(replacedURL, "{Bitrate}", fragBitrate)
	replacedURL = strings.ReplaceAll(replacedURL, "{start time}", fragTime)
	replacedURL = strings.ReplaceAll(replacedURL, "{start_time}", fragTime)
	return url.Parse(replacedURL)
}

func (p *Player) defaultStreamsVideo() *Stream {
	index := 0
	for i, stream := range p.streams {
		if stream.streamType == gott.StreamTypeVideo && stream.bitrate > p.streams[index].bitrate {
			index = i
		}
	}
	return &p.streams[index]
}

func (p *Player) defaultStreamsAudio() *Stream {
	for i, stream := range p.streams {
		if stream.streamType == gott.StreamTypeAudio {
			return &p.streams[i]
		}
	}
	return nil
}

func (p *Player) defaultStreamsText() *Stream {
	for i, stream := range p.streams {
		if stream.streamType == gott.StreamTypeText {
			return &p.streams[i]
		}
	}
	return nil
}

func durationFromTimeScale(duration, timeScale int64) time.Duration {
	d := float64(duration)
	ts := float64(timeScale)
	s := float64(time.Second)
	return time.Duration((d / (ts / s)))
}
