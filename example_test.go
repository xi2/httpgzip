// Copyright (c) 2015 The Httpgzip Authors.
// Use of this source code is governed by a Expat-style
// MIT license that can be found in the LICENSE file.

package httpgzip_test

import (
	"bytes"
	"compress/gzip"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/xi2/httpgzip"
)

// This example is the http.FileServer example from the standard
// library but with gzip compression added.
func ExampleNewHandler() {
	// Simple static webserver:
	log.Fatal(http.ListenAndServe(":8080",
		httpgzip.NewHandler(http.FileServer(http.Dir("/usr/share/doc")), nil)))
}

// This example is the http.FileServer example from the standard
// library but with gzip compression added and writes raw gzipped contents.
func ExampleGzipped() {
	// create buffer to store gzip data
	var buf bytes.Buffer
	// create the gzipped content
	gz := gzip.NewWriter(&buf)
	// ungzipped message
	contents := []byte(strings.Repeat("Hello to raw gzipped content body!!", 200))
	// write it into buffer
	gz.Write(contents)
	// closes gzip writer
	gz.Close()
	// bytes slice of message
	contents = buf.Bytes()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// set default page headers
		w.Header().Set("Content-Length", strconv.Itoa(len(contents))) // the compressed messagem length
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// disable response compression because the contents was compressed
		httpgzip.Gzipped(r)
		// write compressed message to response
		w.Write(contents)
	})
	// Simple static webserver:
	log.Fatal(http.ListenAndServe(":8080", httpgzip.NewHandler(handler, nil)))
}
