package zipfs

// Option represents a single option for a ZipFileSystem
type Option func(*ZipFileSystem)

// ServeIndexForMissing configures the filesystem to always serve the content of index.html (at the root of the zip archive) when a
// nonexistent path is requested. This is useful for single-page applications.
func ServeIndexForMissing() Option {
	return func(z *ZipFileSystem) {
		z.serveIndexForMissing = true
	}
}
