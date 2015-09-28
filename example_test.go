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
