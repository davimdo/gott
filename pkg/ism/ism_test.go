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
