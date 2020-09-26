package ism_test

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/davimdo/gott/pkg/ism"
)

// from fib_test.go
func BenchmarkUnmarshal(b *testing.B) {
	resp, err := http.Get("http://amssamples.streaming.mediaservices.windows.net/683f7e47-bd83-4427-b0a3-26a6c4547782/BigBuckBunny.ism/manifest")
	if err != nil {
		b.Error(err)
	}
	defer resp.Body.Close()
	manifest, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		b.Error(err)
	}
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		_, err := ism.Unmarshal(manifest)
		if err != nil {
			b.Error(err)
		}
	}
}
