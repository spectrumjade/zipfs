package zipfs

import (
	"bytes"
	"os"
)

type httpResource struct {
	*bytes.Reader
	fileInfo os.FileInfo
}

// bytes.Reader doesn't need to be closed, this is just here to implement the io.Reader interface
func (hr *httpResource) Close() error {
	return nil
}

func (hr *httpResource) Stat() (os.FileInfo, error) {
	return hr.fileInfo, nil
}

// Since we don't want to serve a directory listing, serve an empty slice
func (hr *httpResource) Readdir(count int) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}
