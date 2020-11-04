package gott

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Engine is an interface to define an OTT player, to simulate real player
// request for manifest and fragments.
type Engine interface {
	LoadURL(manifest *url.URL) error

	LoadManifest(url *url.URL, manifest []byte) error

	// Streams returns all available streams.
	// Streams can not be loaded if player state is StateNotLoaded.
	Streams() []Stream

	// Context returns player context.
	Context() context.Context
}

// Fetch request the url using the http.Client provided, returning a byte
// slice if an 200 OK obtained, in other case an error is returned.
func Fetch(client *http.Client, manifest *url.URL) ([]byte, error) {
	resp, err := client.Get(manifest.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %s response for Manifest url: %s", resp.Status, resp.Request.URL)
	}
	return ioutil.ReadAll(resp.Body)
}
