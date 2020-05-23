// Package zipfs provides an implementation of an http.FileSystem backed by the contents of a zip archive.
package zipfs

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type zipFileSystem struct {
	files                map[string]*file
	serveIndexForMissing bool
}

type file struct {
	content  []byte
	fileInfo os.FileInfo
}

// NewEmbeddedZipFileSystem creates an http.FileSystem backed by the contents of a zip archive. It reads the zip archive content from the
// currently-running binary file itself. It also accepts an optional list of options.
func NewEmbeddedZipFileSystem(setters ...Option) (http.FileSystem, error) {
	// Open ourselves and try to read
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("could not determine our executable path: %w", err)
	}

	f, err := os.Open(execPath)
	if err != nil {
		return nil, fmt.Errorf("could not open executable for reading: %w", err)
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not stat executable file: %w", err)
	}

	zr, err := zip.NewReader(f, fi.Size())
	if err != nil {
		return nil, fmt.Errorf("error reading embedded zip data: %w", err)
	}

	return NewZipFileSystem(zr, setters...)
}

// NewZipFileSystem creates an http.FileSystem backed by the contents of a zip archive (using from the provided *zip.Reader). It also
// accepts an optional list of options.
func NewZipFileSystem(zr *zip.Reader, setters ...Option) (http.FileSystem, error) {
	zipFS := &zipFileSystem{
		files: make(map[string]*file),
	}

	// Iterate through each file in the zip reader and add them to our map
	for _, zf := range zr.File {
		f := &file{
			fileInfo: zf.FileInfo(),
		}

		reader, err := zf.Open()
		if err != nil {
			return nil, fmt.Errorf("error opening file %s: %w", zf.Name, err)
		}

		content, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %w", zf.Name, err)
		}

		f.content = content

		// Normalize the name by removing trailing slashes. This is needed because the HTTP server will "Open" directories without
		// a trailing slash, while directories in the zip archive always have a trailing slash.
		normalizedName := strings.Trim(zf.Name, "/")

		// We're not protecting against poorly-structured zip archives. If for some reason there are two files with the same name,
		// the last one wins.
		zipFS.files[normalizedName] = f
	}

	// Set any options that were specified
	for _, optionSetter := range setters {
		optionSetter(zipFS)
	}

	return zipFS, nil
}

// Open finds and opens a given file by its name. It normalizes the file path by stripping the leading slash.
func (z *zipFileSystem) Open(name string) (http.File, error) {
	// If we're opening the root (/), return a phony directory. This is to allow an index.html file in the root to work properly.
	if name == "/" {
		return &httpResource{
			Reader:   nil,
			fileInfo: rootDirectoryFileInfo,
		}, nil
	}

	// We need to strip the leading slash off of the name, since the files in the zip archive do not include this.
	normalizedName := strings.Trim(name, "/")

	if f, ok := z.files[normalizedName]; ok {
		return &httpResource{
			Reader:   bytes.NewReader(f.content),
			fileInfo: f.fileInfo,
		}, nil
	}

	// If we get here, no file was found. If the serveIndexForMissing option is set, we should serve the content of index.html.
	// Special case: if the file requested was index.html and it wasn't found, don't try to do the rewrite since it would result in an
	// infinite loop.
	if normalizedName != "index.html" && z.serveIndexForMissing {
		return z.Open("index.html")
	}

	return nil, os.ErrNotExist
}
