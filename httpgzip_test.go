// Copyright (c) 2015 The Httpgzip Authors.
// Use of this source code is governed by an Expat-style
// MIT license that can be found in the LICENSE file.

package httpgzip_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/klauspost/compress/gzip"

	"github.com/xi2/httpgzip"
)

const defComp = httpgzip.DefaultCompression

type fsRequestResponse struct {
	reqFile    string
	reqHeaders []string
	resGzip    bool
	resCode    int
	resHeaders []string
}

var fsTests = []fsRequestResponse{
	// test downloading of all test files in testdata with/without
	// requesting gzip encoding
	{
		reqFile:    "0bytes.txt",
		reqHeaders: nil,
		resGzip:    false,
		resCode:    http.StatusOK,
		resHeaders: []string{
			"Content-Type: text/plain; charset=utf-8",
			"Content-Encoding: ",
			"Content-Length: 0",
			"Accept-Ranges: bytes",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "0bytes.bin",
		reqHeaders: nil,
		resGzip:    false,
		resCode:    http.StatusOK,
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
		resGzip:    false,
		resCode:    http.StatusOK,
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
		resGzip:    false,
		resCode:    http.StatusOK,
		resHeaders: []string{
			"Content-Type: application/octet-stream",
			"Content-Encoding: ",
			"Content-Length: 0",
			"Accept-Ranges: ",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "511bytes.txt",
		reqHeaders: nil,
		resGzip:    false,
		resCode:    http.StatusOK,
		resHeaders: []string{
			"Content-Type: text/plain; charset=utf-8",
			"Content-Encoding: ",
			"Content-Length: 511",
			"Accept-Ranges: bytes",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "511bytes.bin",
		reqHeaders: nil,
		resGzip:    false,
		resCode:    http.StatusOK,
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
		resGzip:    false,
		resCode:    http.StatusOK,
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
		resGzip:    false,
		resCode:    http.StatusOK,
		resHeaders: []string{
			"Content-Type: application/octet-stream",
			"Content-Encoding: ",
			"Content-Length: 511",
			"Accept-Ranges: ",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "512bytes.txt",
		reqHeaders: nil,
		resGzip:    false,
		resCode:    http.StatusOK,
		resHeaders: []string{
			"Content-Type: text/plain; charset=utf-8",
			"Content-Encoding: ",
			"Content-Length: 512",
			"Accept-Ranges: bytes",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "512bytes.bin",
		reqHeaders: nil,
		resGzip:    false,
		resCode:    http.StatusOK,
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
		resGzip:    true,
		resCode:    http.StatusOK,
		resHeaders: []string{
			"Content-Type: text/plain; charset=utf-8",
			"Content-Encoding: gzip",
			"Content-Length: 512|NOMATCH", // look for value != 512
			"Accept-Ranges: ",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "512bytes.bin",
		reqHeaders: []string{"Accept-Encoding: gzip"},
		resGzip:    false,
		resCode:    http.StatusOK,
		resHeaders: []string{
			"Content-Type: application/octet-stream",
			"Content-Encoding: ",
			"Content-Length: 512",
			"Accept-Ranges: ",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: nil,
		resGzip:    false,
		resCode:    http.StatusOK,
		resHeaders: []string{
			"Content-Type: text/plain; charset=utf-8",
			"Content-Encoding: ",
			"Content-Length: 4096",
			"Accept-Ranges: bytes",
			"Vary: Accept-Encoding"},
	},
	{
		reqFile:    "4096bytes.bin",
		reqHeaders: nil,
		resGzip:    false,
		resCode:    http.StatusOK,
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
		resGzip:    true,
		resCode:    http.StatusOK,
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
		resGzip:    false,
		resCode:    http.StatusOK,
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
		resGzip:    false,
		resCode:    http.StatusOK,
	},
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: gzip;q=0"},
		resGzip:    false,
		resCode:    http.StatusOK,
	},
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: identity;q=0"},
		resGzip:    false,
		resCode:    http.StatusNotAcceptable,
	},
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: identity;q=0.5, gzip;q=0.4"},
		resGzip:    false,
		resCode:    http.StatusOK,
	},
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: *"},
		resGzip:    true,
		resCode:    http.StatusOK,
	},
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: *;q=0"},
		resGzip:    false,
		resCode:    http.StatusNotAcceptable,
	},
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: *,gzip;q=0"},
		resGzip:    false,
		resCode:    http.StatusOK,
	},
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: deflate"},
		resGzip:    false,
		resCode:    http.StatusOK,
	},
	// test gzip encoding of non compressible files when forced to by
	// Accept-Encoding header
	{
		reqFile:    "4096bytes.bin",
		reqHeaders: []string{"Accept-Encoding: identity;q=0,gzip"},
		resGzip:    true,
		resCode:    http.StatusOK,
	},
	// test websocket requests are not gzipped
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: gzip", "Sec-WebSocket-Key: XX"},
		resGzip:    false,
		resCode:    http.StatusOK,
	},
	// test Range requests are ignored when requesting gzip encoding
	// and actioned otherwise
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Accept-Encoding: gzip", "Range: bytes=500-"},
		resGzip:    true,
		resCode:    http.StatusOK,
		resHeaders: []string{
			"Accept-Ranges: ",
			"Content-Length: ",
			"Content-Range: ",
		},
	},
	{
		reqFile:    "4096bytes.txt",
		reqHeaders: []string{"Range: bytes=500-"},
		resGzip:    false,
		resCode:    http.StatusPartialContent,
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

// isGzip returns true if the slice b is gzipped data
func isGzip(b []byte) bool {
	if len(b) < 2 {
		return false
	}
	return b[0] == 0x1f && b[1] == 0x8b
}

// getPath starts a temporary test server using handler h (wrapped
// with httpgzip with the given compression level) and issues a
// request for path. The request has the specified headers
// added. getPath returns the http.Response (with Body closed) and the
// result of reading the response Body.
func getPath(t *testing.T, h http.Handler, level int, path string, headers []string) (*http.Response, []byte) {
	gzh, _ := httpgzip.NewHandlerLevel(h, nil, level)
	ts := httptest.NewServer(gzh)
	defer ts.Close()
	req, err := http.NewRequest("GET", ts.URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range headers {
		req.Header.Add(parseHeader(h))
	}
	// explicitly disable automatic sending of "Accept-Encoding"
	transport := &http.Transport{DisableCompression: true}
	client := http.Client{Transport: transport}
	res, err := client.Do(req)
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
	for i, fst := range fsTests {
		res, body := getPath(t, h, defComp, "/"+fst.reqFile, fst.reqHeaders)
		if res.StatusCode != fst.resCode {
			t.Fatalf(
				"\nfile %s, request headers %v\n"+
					"expected status code %d, got %d\n",
				fst.reqFile, fst.reqHeaders, fst.resCode, res.StatusCode)
		}
		if isGzip(body) != fst.resGzip {
			t.Fatalf(
				"\n#%d# file %s, request headers %v\n"+
					"expected gzip status %v, got %v\n",
				i, fst.reqFile, fst.reqHeaders, fst.resGzip, isGzip(body))
		}
		for _, h := range fst.resHeaders {
			k, v := parseHeader(h)
			if strings.HasSuffix(v, "|NOMATCH") {
				v = strings.TrimSuffix(v, "|NOMATCH")
				// fail on match or empty value instead of a non-match
				if res.Header.Get(k) == v || res.Header.Get(k) == "" {
					t.Fatalf(
						"\nfile %s, request headers %v\n"+
							"unexpected response header %s: %s\n",
						fst.reqFile, fst.reqHeaders, k, res.Header.Get(k))
				}
			} else {
				if res.Header.Get(k) != v {
					t.Fatalf(
						"\nfile %s, request headers %v\n"+
							"expected response header %s: %s, got %s: %s\n",
						fst.reqFile, fst.reqHeaders,
						k, v, k, res.Header.Get(k))
				}
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
	res, _ := getPath(t, h, defComp, "/", []string{"Accept-Encoding: gzip"})
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
	res, body := getPath(t, h, defComp, "/", []string{"Accept-Encoding: gzip"})
	expectedEnc := "text/foobar"
	if res.Header.Get("Content-Encoding") != expectedEnc {
		t.Fatalf(
			"\nexpected Content-Encoding %s, got %s\n",
			expectedEnc, res.Header.Get("Content-Encoding"))
	}
	if isGzip(body) {
		t.Fatalf(
			"\nexpected non-gzipped body, got gzipped\n")
	}
}

// TestCompressionLevels creates a handler serving a text file and
// requests that file with Accept-Encoding: gzip with different
// compression levels set. It checks that the sizes of the responses
// vary.
func TestCompressionLevels(t *testing.T) {
	data, err := ioutil.ReadFile(
		filepath.Join("testdata", "4096bytes.txt"))
	if err != nil {
		t.Fatal(err)
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(w, bytes.NewBuffer(data))
	})
	sizes := map[int]struct{}{}
	for _, level := range []int{
		httpgzip.BestSpeed,
		httpgzip.BestCompression,
	} {
		_, body :=
			getPath(t, h, level, "/", []string{"Accept-Encoding: gzip"})
		if _, ok := sizes[len(body)]; ok {
			t.Fatalf(
				"\nlevel %d, body of length %d already received\n",
				level, len(body))
		}
		sizes[len(body)] = struct{}{}
	}
}

// TestGzipped creates a handler serving a gziped content
func TestGzipped(t *testing.T) {
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
	res, body := getPath(t, handler, defComp, "/", []string{"Accept-Encoding: *"})
	expectedEnc := "gzip"
	if res.Header.Get("Content-Encoding") != expectedEnc {
		t.Fatalf(
			"\nexpected Content-Encoding %s, got %s\n",
			expectedEnc, res.Header.Get("Content-Encoding"))
	}
	expectedType := "text/html; charset=utf-8"
	if res.Header.Get("Content-Type") != expectedType {
		t.Fatalf(
			"\nexpected Content-Type %s, got %s\n",
			expectedType, res.Header.Get("Content-Type"))
	}
	if !isGzip(body) {
		t.Fatalf(
			"\nexpected gzipped body, got non-gzipped\n")
	}
	if bytes.Compare(body, contents) != 0 {
		t.Fatalf(
			"\nbad response body\n")
	}
}

func TestGzippedReader(t *testing.T) {
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
		// write compressed message to response
		w.(io.ReaderFrom).ReadFrom(httpgzip.NewGzReader(&buf, func() (io.Reader, error) {
			return bytes.NewBuffer(contents), nil
		}))
	})
	res, body := getPath(t, handler, defComp, "/", []string{"Accept-Encoding: *"})
	expectedEnc := "gzip"
	if res.Header.Get("Content-Encoding") != expectedEnc {
		t.Fatalf(
			"\nexpected Content-Encoding %s, got %s\n",
			expectedEnc, res.Header.Get("Content-Encoding"))
	}
	expectedType := "text/html; charset=utf-8"
	if res.Header.Get("Content-Type") != expectedType {
		t.Fatalf(
			"\nexpected Content-Type %s, got %s\n",
			expectedType, res.Header.Get("Content-Type"))
	}
	if !isGzip(body) {
		t.Fatalf(
			"\nexpected gzipped body, got non-gzipped\n")
	}
	if bytes.Compare(body, contents) != 0 {
		t.Fatalf(
			"\nbad response body\n")
	}
}