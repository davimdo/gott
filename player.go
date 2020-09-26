package gott

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Player is an interface to define an OTT player, to simulate real player
// request for manifest and fragments.
type Player interface {
	// Load requests the manifest URL to get all available streams and prepares
	// the player to start requesting fragments from selected streams. Load can
	// only be called if player state is PlayerStateNotLoaded, after executing
	// it If no error is returned, this method will change player state to
	// PlayerStateLoaded.
	Load() error

	// Play, for each one of the stream passed by argument will call PlayStream
	// concurrently. Play stop the "playout" upon an error on any of the Streams
	// and returns it. If an error is raised, all concurrent playout are canceled.
	//
	// Play can only be called if PlayerState is PlayerStateLoaded. And when all
	// streams are done, or an error is raised PlayerState is set to PlayerLoaded.
	//
	// streams must not be null.
	Play(streams []Stream, position int) error

	// Stop cancel all current Stream actually playing by the player by. This
	// method can only be executed if player state is StatePlaying.
	Stop() error

	// State returns the player state.
	State() State

	// Streams returns all available streams.
	// Streams can not be loaded if player state is StateNotLoaded.
	Streams() []Stream

	// Streams returns default streams. What means "default" depends on
	// player implementation.
	// Streams can not be loaded if player state is StateNotLoaded.
	DefaultStreams() []Stream

	// Context returns player context.
	Context() context.Context

	// WithContext sets the context to the player
	WithContext(ctx context.Context)
}

// State represents the state of a player
type State int

const (
	// StateNotLoaded state for the player when it has not parse the
	// manifest to get the available streams.
	StateNotLoaded State = iota
	// StateLoaded state for the player when it has parse the manifest to
	// get the available streams.
	StateLoaded
	// StatePlaying when the player is actually requesting Stream
	// fragments.
	StatePlaying
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
