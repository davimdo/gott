package dash

import (
	"net/http"

	"github.com/davimdo/gott"
)

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

func (s *Stream) HttpClient() *http.Client {
	return s.http
}
