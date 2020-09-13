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
const USAGE string = `usage: %s MANIFEST POSITION [INTERVAL]

	MANIFEST   OTT Manifest URL (for now Smooth)
	POSITION   Position in which start \"playing\"
	INTERVAL   Time interval in seconds to repeat the same manifest playout
`

func main() {
	if len(os.Args) != 4 {
		log.Fatalf(USAGE, os.Args[0])
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
		time.Sleep(time.Duration(interval) * time.Second)
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
