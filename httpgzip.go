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

// Package httpgzip implements an http.Handler which wraps an existing
// http.Handler adding Gzip compression for appropriate requests.
//
// It attempts to properly parse the request's Accept-Encoding header
// according to RFC 2616 and does not just do a
// strings.Contains(header,"gzip"). It will serve either gzip or
// identity content codings (or return 406 Not Acceptable status if it
// can do neither).
//
// It works correctly with handlers such as http.FileServer which
// honour Range request headers by removing the Range header when
// requests prefer gzip encoding. This is necessary since Range
// applies to the Gzipped content and the wrapped handler is not aware
// of the compression when it writes byte ranges.
package httpgzip // import "xi2.org/x/httpgzip"

import (
	"bytes"
	"compress/gzip"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// DefaultContentTypes is the default set of content types with which
// a Handler applies Gzip compression. This set originates from the
// file compression.conf within the Apache configuration found at
// https://html5boilerplate.com/.
var DefaultContentTypes = map[string]struct{}{
	"application/atom+xml":                struct{}{},
	"application/javascript":              struct{}{},
	"application/json":                    struct{}{},
	"application/ld+json":                 struct{}{},
	"application/manifest+json":           struct{}{},
	"application/rdf+xml":                 struct{}{},
	"application/rss+xml":                 struct{}{},
	"application/schema+json":             struct{}{},
	"application/vnd.geo+json":            struct{}{},
	"application/vnd.ms-fontobject":       struct{}{},
	"application/x-font-ttf":              struct{}{},
	"application/x-javascript":            struct{}{},
	"application/x-web-app-manifest+json": struct{}{},
	"application/xhtml+xml":               struct{}{},
	"application/xml":                     struct{}{},
	"font/eot":                            struct{}{},
	"font/opentype":                       struct{}{},
	"image/bmp":                           struct{}{},
	"image/svg+xml":                       struct{}{},
	"image/vnd.microsoft.icon":            struct{}{},
	"image/x-icon":                        struct{}{},
	"text/cache-manifest":                 struct{}{},
	"text/css":                            struct{}{},
	"text/html":                           struct{}{},
	"text/javascript":                     struct{}{},
	"text/plain":                          struct{}{},
	"text/vcard":                          struct{}{},
	"text/vnd.rim.location.xloc":          struct{}{},
	"text/vtt":                            struct{}{},
	"text/x-component":                    struct{}{},
	"text/x-cross-domain-policy":          struct{}{},
	"text/xml":                            struct{}{},
}

var gzipWriterPool = sync.Pool{
	New: func() interface{} { return gzip.NewWriter(nil) },
}

var gzipBufPool = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

// A gzipResponseWriter is a modified http.ResponseWriter. If the
// request only accepts Gzip encoding or the content to be written is
// of a type contained in contentTypes and the request prefers Gzip
// encoding then the response is compressed and the Content-Encoding
// header is set. Otherwise a gzipResponseWriter behaves mostly like a
// normal http.ResponseWriter. It is important to call the Close
// method when writing is finished in order to flush and close the
// Writer. The encoding slice encs must contain at least one encoding.
type gzipResponseWriter struct {
	http.ResponseWriter
	httpStatus   int
	contentTypes map[string]struct{}
	encs         []encoding
	gw           *gzip.Writer
	buf          *bytes.Buffer
}

func newGzipResponseWriter(w http.ResponseWriter, contentTypes map[string]struct{}, encs []encoding) *gzipResponseWriter {
	buf := gzipBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	return &gzipResponseWriter{
		ResponseWriter: w,
		httpStatus:     http.StatusOK,
		contentTypes:   contentTypes,
		encs:           encs,
		buf:            buf}
}

// init gets called by Write once at least 512 bytes have been written
// to the temporary buffer buf, or by Close if it has not yet been
// called. Firstly it determines the content type, either from the
// Content-Type header, or by calling http.DetectContentType on
// buf. Then, if needed, a gzip.Writer is initialized. Lastly,
// appropriate headers are set and the ResponseWriter's WriteHeader
// method is called.
func (w *gzipResponseWriter) init() {
	cth := w.Header().Get("Content-Type")
	var ct string
	if cth != "" {
		ct = cth
	} else {
		ct = http.DetectContentType(w.buf.Bytes())
	}
	var gzipContentType bool
	if mt, _, err := mime.ParseMediaType(ct); err == nil {
		if _, ok := w.contentTypes[mt]; ok {
			gzipContentType = true
		}
	}
	var useGzip bool
	if w.Header().Get("Content-Encoding") == "" {
		switch {
		case w.encs[0] == encGzip && gzipContentType,
			w.encs[0] == encGzip && len(w.encs) == 1:
			useGzip = true
		}
	}
	if useGzip {
		w.gw = gzipWriterPool.Get().(*gzip.Writer)
		w.gw.Reset(w.ResponseWriter)
		w.Header().Del("Accept-Ranges")
		w.Header().Del("Content-Length")
		w.Header().Del("Content-Range")
		w.Header().Set("Content-Encoding", "gzip")
	}
	if cth == "" {
		w.Header().Set("Content-Type", ct)
	}
	w.ResponseWriter.WriteHeader(w.httpStatus)
}

func (w *gzipResponseWriter) Write(p []byte) (int, error) {
	var n, written int
	var err error
	if w.buf != nil {
		written = w.buf.Len()
		_, _ = w.buf.Write(p)
		if w.buf.Len() < 512 {
			return len(p), nil
		}
		w.init()
		p = w.buf.Bytes()
		defer func() {
			gzipBufPool.Put(w.buf)
			w.buf = nil
		}()
	}
	switch {
	case w.gw != nil:
		n, err = w.gw.Write(p)
	default:
		n, err = w.ResponseWriter.Write(p)
	}
	n -= written
	if n < 0 {
		n = 0
	}
	return n, err
}

func (w *gzipResponseWriter) WriteHeader(httpStatus int) {
	// postpone WriteHeader call until end of init method
	w.httpStatus = httpStatus
}

func (w *gzipResponseWriter) Close() (err error) {
	if w.buf != nil {
		w.init()
		p := w.buf.Bytes()
		defer func() {
			gzipBufPool.Put(w.buf)
			w.buf = nil
		}()
		switch {
		case w.gw != nil:
			_, err = w.gw.Write(p)
		default:
			_, err = w.ResponseWriter.Write(p)
		}
	}
	if w.gw != nil {
		e := w.gw.Close()
		if e != nil {
			err = e
		}
		gzipWriterPool.Put(w.gw)
		w.gw = nil
	}
	return err
}

// An encoding is a supported content coding.
type encoding int

const (
	encIdentity encoding = iota
	encGzip
)

// acceptedEncodings returns the supported content codings that are
// accepted by the request r. It returns a slice of encodings in
// client preference order.
//
// If the Sec-WebSocket-Key header is present then compressed content
// encodings are not considered.
//
// ref: http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html
func acceptedEncodings(r *http.Request) []encoding {
	h := r.Header.Get("Accept-Encoding")
	swk := r.Header.Get("Sec-WebSocket-Key")
	if h == "" {
		return []encoding{encIdentity}
	}
	gzip := float64(-1)    // -1 means not accepted, 0 -> 1 means value of q
	identity := float64(0) // -1 means not accepted, 0 -> 1 means value of q
	for _, s := range strings.Split(h, ",") {
		f := strings.Split(s, ";")
		f0 := strings.ToLower(strings.Trim(f[0], " "))
		q := float64(1.0)
		if len(f) > 1 {
			f1 := strings.ToLower(strings.Trim(f[1], " "))
			if strings.HasPrefix(f1, "q=") {
				if flt, err := strconv.ParseFloat(f1[2:], 32); err == nil {
					if flt >= 0 && flt <= 1 {
						q = flt
					}
				}
			}
		}
		if (f0 == "gzip" || f0 == "*") && q > gzip && swk == "" {
			gzip = q
		}
		if (f0 == "gzip" || f0 == "*") && q == 0 {
			gzip = -1
		}
		if (f0 == "identity" || f0 == "*") && q > identity {
			identity = q
		}
		if (f0 == "identity" || f0 == "*") && q == 0 {
			identity = -1
		}
	}
	switch {
	case gzip == -1 && identity == -1:
		return []encoding{}
	case gzip == -1:
		return []encoding{encIdentity}
	case identity == -1:
		return []encoding{encGzip}
	case identity > gzip:
		return []encoding{encIdentity, encGzip}
	default:
		return []encoding{encGzip, encIdentity}
	}
}

// NewHandler returns a new http.Handler which wraps a handler h
// adding Gzip compression to responses whose content types are in
// contentTypes (unless the corresponding request does not allow or
// prefer Gzip compression). If contentTypes is nil then it is set to
// DefaultContentTypes.
//
// The new http.Handler sets the Content-Encoding, Vary and
// Content-Type headers in its responses as appropriate. If the
// request expresses a preference for gzip encoding then any Range
// headers are removed from the request before forwarding it to
// h. This happens regardless of whether gzip encoding is eventually
// used in the response or not.
func NewHandler(h http.Handler, contentTypes map[string]struct{}) http.Handler {
	if contentTypes == nil {
		contentTypes = DefaultContentTypes
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// add Vary header
		w.Header().Add("Vary", "Accept-Encoding")
		// check client's accepted encodings
		encs := acceptedEncodings(r)
		// return if no acceptable encodings
		if len(encs) == 0 {
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
		if encs[0] == encGzip {
			// cannot accept Range requests for possibly gzipped
			// responses
			r.Header.Del("Range")
		}
		w = newGzipResponseWriter(w, contentTypes, encs)
		defer w.(*gzipResponseWriter).Close()
		// call original handler's ServeHTTP
		h.ServeHTTP(w, r)
	})
}
