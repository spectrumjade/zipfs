package zipfs

import (
	"os"
	"time"
)

// dummyDirectoryFileInfo is a pseudo-type used to indicate that a file is just a directory. It's useful to define a root directory.
type dummyDirectoryFileInfo struct{}

var rootDirectoryFileInfo dummyDirectoryFileInfo

func (d dummyDirectoryFileInfo) Name() string       { return "" }
func (d dummyDirectoryFileInfo) Size() int64        { return 0 }
func (d dummyDirectoryFileInfo) IsDir() bool        { return true }
func (d dummyDirectoryFileInfo) ModTime() time.Time { return time.Now() }
func (d dummyDirectoryFileInfo) Mode() os.FileMode  { return 0755 }
func (d dummyDirectoryFileInfo) Sys() interface{}   { return nil }
