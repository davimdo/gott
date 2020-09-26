package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/davimdo/gott/smooth"
)

// USAGE is the CMD help
const USAGE string = `usage: %s MANIFEST POSITION INTERVAL

	MANIFEST   OTT Manifest URL (for now Smooth)
	POSITION   Position in which start \"playing\"
	INTERVAL   Time interval in millisec to repeat the same manifest playout
`

func main() {
	if len(os.Args) != 4 {
		fmt.Printf(USAGE, os.Args[0])
		os.Exit(-1)
	}

	position, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalln(err)
	}
	interval, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatalln(err)
	}

	for {
		go func(url string, position int) {
			launchPlayer(url, position)
		}(os.Args[1], position)
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

func launchPlayer(url string, position int) {
	smoothPlayer, err := smooth.NewPlayer(url)
	if err != nil {
		log.Fatalln(err)
	}
	err = smoothPlayer.Load()
	if err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			log.Fatalln(err)
		}
		os.Exit(-1)
	}

	err = smoothPlayer.Play(smoothPlayer.DefaultStreams(), position)
	if err != nil {
		log.Fatalln(err)
	}
}
