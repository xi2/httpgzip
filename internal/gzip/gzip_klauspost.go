// +build klauspost

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

// Package gzip is a partial implementation of the gzip API using the
// github.com/klauspost/compress/gzip package. It contains the part of
// the API used by httpgzip.
package gzip

import (
	"io"

	"github.com/klauspost/compress/gzip"
)

const (
	NoCompression      = gzip.NoCompression
	BestSpeed          = gzip.BestSpeed
	BestCompression    = gzip.BestCompression
	DefaultCompression = gzip.DefaultCompression
)

type Writer gzip.Writer

func NewWriterLevel(w io.Writer, level int) (*Writer, error) {
	z, err := gzip.NewWriterLevel(w, level)
	return (*Writer)(z), err
}

func (z *Writer) Reset(w io.Writer) {
	(*gzip.Writer)(z).Reset(w)
}

func (z *Writer) Write(p []byte) (int, error) {
	return (*gzip.Writer)(z).Write(p)
}

func (z *Writer) Close() error {
	return (*gzip.Writer)(z).Close()
}
