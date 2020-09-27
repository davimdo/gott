package gott

import (
	"context"
)

// Engine is an interface to define an OTT player, to simulate real player
// request for manifest and fragments.
type Engine interface {
	// Load requests the manifest URL to get all available streams and prepares
	// the player to start requesting fragments from selected streams. Load can
	// only be called if player state is PlayerStateNotLoaded, after executing
	// it If no error is returned, this method will change player state to
	// PlayerStateLoaded.
	Load() error

	// IsLive check if loaded OTT manifest is Live.
	IsLive() bool

	// Streams returns all available streams.
	// Streams can not be loaded if player state is StateNotLoaded.
	Streams() []Stream

	// Context returns player context.
	Context() context.Context
}
