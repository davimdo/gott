package ism

import (
	"io/ioutil"
	"net/http"
	"testing"
)

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

	for n := 0; n < b.N; n++ {
		_, err := Unmarshal(manifest)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkMarshal(b *testing.B) {
	resp, err := http.Get("http://amssamples.streaming.mediaservices.windows.net/683f7e47-bd83-4427-b0a3-26a6c4547782/BigBuckBunny.ism/manifest")
	if err != nil {
		b.Error(err)
	}
	defer resp.Body.Close()
	manifest, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		b.Error(err)
	}
	ism, err := Unmarshal(manifest)
	if err != nil {
		b.Error(err)
	}

	for n := 0; n < b.N; n++ {
		_, err := ism.Marshal()
		if err != nil {
			b.Error(err)
		}
	}
}
