/*
   Copyright 2015 The Httpgzip Authors. See the AUTHORS file at the
   top-level directory of this distribution and at
   <https://xi2.org/x/httpgzip/m/AUTHORS>.

   This file is part of Httpgzip.

   Httpgzip is free software: you can redistribute it and/or modify it
   under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   Httpgzip is distributed in the hope that it will be useful, but
   WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
   General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with Httpgzip.  If not, see <https://www.gnu.org/licenses/>.
*/

package httpgzip_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"xi2.org/x/httpgzip"
)

type fsRequestResponse struct {
	reqFile    string
	reqHeaders []string
	resCode    int
	resLength  int
	resHeaders []string
}

var fsTests = []fsRequestResponse{
	// test all test files with/without requesting gzip encoding
	{
		reqFile:    "0bytes.txt",
		reqHeaders: []string{"Accept-Encoding: "},
		resCode:    http.StatusOK,
		resLength:  0,
		resHeaders: []string{
			"Content-Type: text/plain; charset=utf-8",
			"Content-Encoding: ",
			"Content-Length: 0",
			"Accept-Ranges: bytes",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "0bytes.bin",
		reqHeaders: []string{"Accept-Encoding: "},
		resCode:    http.StatusOK,
		resLength:  0,
		resHeaders: []string{
			"Content-Type: application/octet-stream",
			"Content-Encoding: ",
			"Content-Length: 0",
			"Accept-Ranges: bytes",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "0bytes.txt",
		reqHeaders: []string{"Accept-Encoding: gzip"},
		resCode:    http.StatusOK,
		resLength:  0,
		resHeaders: []string{
			"Content-Type: text/plain; charset=utf-8",
			"Content-Encoding: ",
			"Content-Length: 0",
			"Accept-Ranges: ",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "0bytes.bin",
		reqHeaders: []string{"Accept-Encoding: gzip"},
		resCode:    http.StatusOK,
		resLength:  0,
		resHeaders: []string{
			"Content-Type: application/octet-stream",
			"Content-Encoding: ",
			"Content-Length: 0",
			"Accept-Ranges: ",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "511bytes.txt",
		reqHeaders: []string{"Accept-Encoding: "},
		resCode:    http.StatusOK,
		resLength:  511,
		resHeaders: []string{
			"Content-Type: text/plain; charset=utf-8",
			"Content-Encoding: ",
			"Content-Length: 511",
			"Accept-Ranges: bytes",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "511bytes.bin",
		reqHeaders: []string{"Accept-Encoding: "},
		resCode:    http.StatusOK,
		resLength:  511,
		resHeaders: []string{
			"Content-Type: application/octet-stream",
			"Content-Encoding: ",
			"Content-Length: 511",
			"Accept-Ranges: bytes",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "511bytes.txt",
		reqHeaders: []string{"Accept-Encoding: gzip"},
		resCode:    http.StatusOK,
		resLength:  511,
		resHeaders: []string{
			"Content-Type: text/plain; charset=utf-8",
			"Content-Encoding: ",
			"Content-Length: 511",
			"Accept-Ranges: ",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "511bytes.bin",
		reqHeaders: []string{"Accept-Encoding: gzip"},
		resCode:    http.StatusOK,
		resLength:  511,
		resHeaders: []string{
			"Content-Type: application/octet-stream",
			"Content-Encoding: ",
			"Content-Length: 511",
			"Accept-Ranges: ",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "512bytes.txt",
		reqHeaders: []string{"Accept-Encoding: "},
		resCode:    http.StatusOK,
		resLength:  512,
		resHeaders: []string{
			"Content-Type: text/plain; charset=utf-8",
			"Content-Encoding: ",
			"Content-Length: 512",
			"Accept-Ranges: bytes",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "512bytes.bin",
		reqHeaders: []string{"Accept-Encoding: "},
		resCode:    http.StatusOK,
		resLength:  512,
		resHeaders: []string{
			"Content-Type: application/octet-stream",
			"Content-Encoding: ",
			"Content-Length: 512",
			"Accept-Ranges: bytes",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "512bytes.txt",
		reqHeaders: []string{"Accept-Encoding: gzip"},
		resCode:    http.StatusOK,
		resLength:  339,
		resHeaders: []string{
			"Content-Type: text/plain; charset=utf-8",
			"Content-Encoding: gzip",
			"Content-Length: 339",
			"Accept-Ranges: ",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "512bytes.bin",
		reqHeaders: []string{"Accept-Encoding: gzip"},
		resCode:    http.StatusOK,
		resLength:  512,
		resHeaders: []string{
			"Content-Type: application/octet-stream",
			"Content-Encoding: ",
			"Content-Length: 512",
			"Accept-Ranges: ",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: "},
		resCode:    http.StatusOK,
		resLength:  4096,
		resHeaders: []string{
			"Content-Type: text/plain; charset=utf-8",
			"Content-Encoding: ",
			"Content-Length: 4096",
			"Accept-Ranges: bytes",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "4096bytes.bin",
		reqHeaders: []string{"Accept-Encoding: "},
		resCode:    http.StatusOK,
		resLength:  4096,
		resHeaders: []string{
			"Content-Type: application/octet-stream",
			"Content-Encoding: ",
			"Content-Length: 4096",
			"Accept-Ranges: bytes",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: gzip"},
		resCode:    http.StatusOK,
		resLength:  2327,
		resHeaders: []string{
			"Content-Type: text/plain; charset=utf-8",
			"Content-Encoding: gzip",
			"Content-Length: ",
			"Accept-Ranges: ",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "4096bytes.bin",
		reqHeaders: []string{"Accept-Encoding: gzip"},
		resCode:    http.StatusOK,
		resLength:  4096,
		resHeaders: []string{
			"Content-Type: application/octet-stream",
			"Content-Encoding: ",
			"Content-Length: 4096",
			"Accept-Ranges: ",
			"Vary: Accept-Encoding"},
	},
	// test Accept-Encoding parsing
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: gzip;q=0.5"},
		resCode:    http.StatusOK,
		resLength:  2327, // gzipped
	},
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: gzip;q=0"},
		resCode:    http.StatusOK,
		resLength:  4096,
	},
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: identity;q=0"},
		resCode:    http.StatusNotAcceptable,
		resLength:  0,
	},
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: identity;q=0.5, gzip;q=0.4"},
		resCode:    http.StatusOK,
		resLength:  4096,
	},
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: *"},
		resCode:    http.StatusOK,
		resLength:  2327, // gzipped
	},
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: *;q=0"},
		resCode:    http.StatusNotAcceptable,
		resLength:  0,
	},
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: *,gzip;q=0"},
		resCode:    http.StatusOK,
		resLength:  4096,
	},
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: deflate"},
		resCode:    http.StatusOK,
		resLength:  4096,
	},
	// test gzip encoding of non compressible files when forced to by
	// Accept-Encoding header
	{
		reqFile:    "4096bytes.bin",
		reqHeaders: []string{"Accept-Encoding: identity;q=0,gzip"},
		resCode:    http.StatusOK,
		resLength:  4124, // gzipped
	},
	// test websocket requests are not gzipped
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: gzip", "Sec-WebSocket-Key: XX"},
		resCode:    http.StatusOK,
		resLength:  4096,
	},
	// test Range requests are ignored when requesting gzip encoding
	// and actioned otherwise
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: gzip", "Range: bytes=500-"},
		resCode:    http.StatusOK,
		resLength:  2327,
		resHeaders: []string{
			"Accept-Ranges: ",
			"Content-Length: ",
			"Content-Range: ",
		},
	},
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Range: bytes=500-"},
		resCode:    http.StatusPartialContent,
		resLength:  3596,
		resHeaders: []string{
			"Accept-Ranges: bytes",
			"Content-Length: 3596",
			"Content-Range: bytes 500-4095/4096",
		},
	},
}

// parseHeader returns a header key and value from a "Key: Value" string
func parseHeader(header string) (key, value string) {
	i := strings.IndexByte(header, ':')
	key = header[:i]
	value = strings.TrimLeft(header[i+1:], " ")
	return
}

// getPath starts a temporary test server using handler h (wrapped
// with httpgzip) and requests the given path. The request has the
// given headers added. getPath returns the http.Response (with Body
// closed) and the result of reading the response Body.
func getPath(t *testing.T, h http.Handler, path string, headers []string) (*http.Response, []byte) {
	ts := httptest.NewServer(httpgzip.NewHandler(h, nil))
	defer ts.Close()
	req, err := http.NewRequest("GET", ts.URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range headers {
		req.Header.Add(parseHeader(h))
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return res, body
}

// TestFileServer runs all tests in fsTests against an http.FileServer
// serving the testdata directory.
func TestFileServer(t *testing.T) {
	h := http.FileServer(http.Dir("testdata"))
	for _, fst := range fsTests {
		res, body := getPath(t, h, "/"+fst.reqFile, fst.reqHeaders)
		if res.StatusCode != fst.resCode {
			t.Fatalf(
				"\nfile %s, request headers %v\n"+
					"expected status code %d, got %d\n",
				fst.reqFile, fst.reqHeaders, fst.resCode, res.StatusCode)
		}
		if len(body) != fst.resLength {
			t.Fatalf(
				"\nfile %s, request headers %v\n"+
					"expected body length %d, got %d\n",
				fst.reqFile, fst.reqHeaders, fst.resLength, len(body))
		}
		for _, h := range fst.resHeaders {
			k, v := parseHeader(h)
			if res.Header.Get(k) != v {
				t.Fatalf(
					"\nfile %s, request headers %v\n"+
						"expected response header %s: %s, got %s: %s\n",
					fst.reqFile, fst.reqHeaders, k, v, k, res.Header.Get(k))
			}
		}
	}
}

// TestDetectContentType creates a handler serving a text file which
// does not set Content-Type, wraps it with httpgzip, and requests
// that file with Accept-Encoding: gzip. It checks that httpgzip sets
// Content-Type and it is not left to the standard library (which
// would set it to "application/x-gzip").
func TestDetectContentType(t *testing.T) {
	data, err := ioutil.ReadFile(
		filepath.Join("testdata", "4096bytes.txt"))
	if err != nil {
		t.Fatal(err)
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(w, bytes.NewBuffer(data))
	})
	res, _ := getPath(t, h, "/", []string{"Accept-Encoding: gzip"})
	expected := "text/plain; charset=utf-8"
	if res.Header.Get("Content-Type") != expected {
		t.Fatalf(
			"\nexpected Content-Type %s, got %s\n",
			expected, res.Header.Get("Content-Type"))
	}
}

// TestPresetContentEncoding creates a handler serving a text file
// which sets Content-Encoding, wraps it with httpgzip, and requests
// that file with Accept-Encoding: gzip. It checks that httpgzip does
// not mess with Content-Encoding and serves the file without
// compression as expected.
func TestPresetContentEncoding(t *testing.T) {
	data, err := ioutil.ReadFile(
		filepath.Join("testdata", "4096bytes.txt"))
	if err != nil {
		t.Fatal(err)
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "text/foobar")
		_, _ = io.Copy(w, bytes.NewBuffer(data))
	})
	res, body := getPath(t, h, "/", []string{"Accept-Encoding: gzip"})
	expectedEnc := "text/foobar"
	if res.Header.Get("Content-Encoding") != expectedEnc {
		t.Fatalf(
			"\nexpected Content-Encoding %s, got %s\n",
			expectedEnc, res.Header.Get("Content-Encoding"))
	}
	expectedLen := 4096 // not compressed by httpgzip
	if len(body) != expectedLen {
		t.Fatalf(
			"\nexpected body length %d, got %d\n",
			expectedLen, len(body))
	}
}
