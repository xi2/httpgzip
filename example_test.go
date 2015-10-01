// Copyright (c) 2015 The Httpgzip Authors.
// Use of this source code is governed by a Expat-style
// MIT license that can be found in the LICENSE file.

package httpgzip_test

import (
	"log"
	"net/http"

	"xi2.org/x/httpgzip"
)

// This example is the http.FileServer example from the standard
// library but with gzip compression added.
func ExampleNewHandler() {
	// Simple static webserver:
	log.Fatal(http.ListenAndServe(":8080",
		httpgzip.NewHandler(http.FileServer(http.Dir("/usr/share/doc")), nil)))
}
