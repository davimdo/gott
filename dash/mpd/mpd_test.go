package mpd

import (
	"io/ioutil"
	"net/http"
	"testing"
)

func BenchmarkUnmarshal(b *testing.B) {
	resp, err := http.Get("https://dash.akamaized.net/dash264/TestCases/2c/qualcomm/1/MultiResMPEG2.mpd")
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
	resp, err := http.Get("https://dash.akamaized.net/dash264/TestCases/2c/qualcomm/1/MultiResMPEG2.mpd")
	if err != nil {
		b.Error(err)
	}
	defer resp.Body.Close()
	manifest, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		b.Error(err)
	}
	mpd, err := Unmarshal(manifest)
	if err != nil {
		b.Error(err)
	}

	for n := 0; n < b.N; n++ {
		_, err := mpd.Marshal()
		if err != nil {
			b.Error(err)
		}
	}
}
