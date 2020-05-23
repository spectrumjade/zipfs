# zipfs
Package `zipfs` is an implementation of a Go `http.FileSystem`. Instead of serving files from a local filesystem directory, it serves the
contents of a provided zip file. This might be useful to you if you have a Golang HTTP server (perhaps an API server) and would like to
serve static content (perhaps a single-page application).

Hosting static content from a zip archive has several nice properties:
- Data is stored in compressed form (of course), which makes it relatively efficient for large static asset bundles.
- File metadata (particularly modification time) is preserved, which allows for caching.
- Because zip  archive data can be prefixed with any amount of arbitrary data, you can actually embed it into your server's binary
  executable file by concatenating it. This is extremely convenient and is perhaps the most compelling use case. An example of this is
  below.

## Usage
### Embedded zip data
As described above, it is possible to embed zip data directly into your server's binary executable file. Here's some example code for this:
```go
package main

import (
	"log"
	"net/http"

	"gerace.dev/zipfs"
)

func main() {
	fs, err := zipfs.NewEmbeddedZipFileSystem()
	if err != nil {
		log.Fatalf("Error setting up zipfs: %v", err)
	}

	log.Fatal(http.ListenAndServe("127.0.0.1:8080", http.FileServer(fs)))
}

```
After you build your executable, append your zip file to it like so:
```
cat webassets.zip >>myserverbinary
zip -A myserverbinary
```
That's it!

### Loading a normal zip file
For the more general case of supplying a zip file normally, you can pass any `*zip.Reader` and create a `ZipFileSystem` instance from that.
```go
package main

import (
	"archive/zip"
	"log"
	"net/http"
	"os"

	"gerace.dev/zipfs"
)

func main() {
	f, err := os.Open("webassets.zip")
	if err != nil {
		log.Fatalf("Error opening webassets.zip: %v", err)
	}

	fi, err := f.Stat()
	if err != nil {
		log.Fatalf("Error getting file information about webassets.zip: %v", err)
	}

	zr, err := zip.NewReader(f, fi.Size())
	if err != nil {
		log.Fatalf("Error reading zip data from webassets.zip: %v", err)
	}

	fs, err := zipfs.NewZipFileSystem(zr)
	if err != nil {
		log.Fatalf("Error creating zip filesystem: %v", err)
	}

	log.Fatal(http.ListenAndServe("127.0.0.1:8080", http.FileServer(fs)))
}
```

### Specifying Options
Options can be specified during instantiation, like so:
```go
package main

import (
	"log"
	"net/http"

	"gerace.dev/zipfs"
)

func main() {
	fs, err := zipfs.NewEmbeddedZipFileSystem(zipfs.ServeIndexForMissing())
	if err != nil {
		log.Fatalf("Error setting up zipfs: %v", err)
	}

	log.Fatal(http.ListenAndServe("127.0.0.1:8080", http.FileServer(fs)))
}
```

## Options
### `ServeIndexForMissing`
This option is useful for hosting single-page applications that mutate the browser history state locally. When specified, requests for
files that aren't in the zip archive will receive the content of `index.html` at the root of the zip archive, if it exists.
