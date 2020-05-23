package zipfs_test

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"gerace.dev/zipfs"
	"github.com/stretchr/testify/assert"
)

// Let's set up some data to test with
const (
	rootIndexHTML     = "<html><head><title>Root index</title></head><body><h1>Root index.html file</h1></body></html>"
	subdirIndexHTML   = "<html><head><title>Subdirectory index</title></head><body><h1>Subdirectory index.html file</h1></body></html>"
	exampleTxtContent = "Example text file"
)

type testDataFile struct {
	header  *zip.FileHeader
	content []byte
}

var testDataFiles = []testDataFile{
	{
		header: &zip.FileHeader{
			Name:   "index.html",
			Method: zip.Deflate,
		},
		content: []byte(rootIndexHTML),
	},
	{
		header: &zip.FileHeader{
			Name:   "emptydirectory/",
			Method: zip.Deflate,
		},
	},
	{
		header: &zip.FileHeader{
			Name:   "nonemptydirectory/",
			Method: zip.Deflate,
		},
	},
	{
		header: &zip.FileHeader{
			Name:   "nonemptydirectory/file.txt",
			Method: zip.Deflate,
		},
		content: []byte(exampleTxtContent),
	},
	{
		header: &zip.FileHeader{
			Name:   "directorywithindex/",
			Method: zip.Deflate,
		},
	},
	{
		header: &zip.FileHeader{
			Name:   "directorywithindex/index.html",
			Method: zip.Deflate,
		},
		content: []byte(subdirIndexHTML),
	},
}

var zipReader *zip.Reader

func init() {
	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)

	for _, tdf := range testDataFiles {
		w, _ := zw.CreateHeader(tdf.header)
		w.Write(tdf.content)
	}

	zw.Close()

	zr, _ := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	zipReader = zr
}

func checkGetMatches(t *testing.T, url string, expectedContent []byte) {
	resp, err := http.Get(url)
	assert.NoErrorf(t, err, "GET %s should not result in error", url)
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	assert.NoErrorf(t, err, "reading response body of %s should not result in error", url)

	assert.Equalf(t, expectedContent, respBody, "content of %s should match expected exactly", url)
}

func checkNotFound(t *testing.T, url string) {
	resp, err := http.Get(url)
	assert.NoErrorf(t, err, "GET %s should not result in error", url)
	defer resp.Body.Close()

	assert.Equalf(t, http.StatusNotFound, resp.StatusCode, "GET %s should result in a 404 not found response", url)
}

// Test the zip filesystem in its default mode
func TestZipFileSystem(t *testing.T) {
	fs, err := zipfs.NewZipFileSystem(zipReader)
	assert.NoError(t, err, "NewZipFileSystem should not result in error")

	// Set up test HTTP server
	fileServer := http.FileServer(fs)
	srv := httptest.NewServer(fileServer)
	defer srv.Close()

	// Get both `/` and `/index.html` which should serve the index content.
	checkGetMatches(t, srv.URL, []byte(rootIndexHTML))
	checkGetMatches(t, srv.URL+"/index.html", []byte(rootIndexHTML))

	// Check a subdirectory with an index.html file, both with and without a trailing slash
	checkGetMatches(t, srv.URL+"/directorywithindex", []byte(subdirIndexHTML))
	checkGetMatches(t, srv.URL+"/directorywithindex/", []byte(subdirIndexHTML))
	checkGetMatches(t, srv.URL+"/directorywithindex/index.html", []byte(subdirIndexHTML))

	// Test get of a standard file
	checkGetMatches(t, srv.URL+"/nonemptydirectory/file.txt", []byte(exampleTxtContent))

	// Check for nonexistent files
	checkNotFound(t, srv.URL+"/nonexistentfile.txt")
	checkNotFound(t, srv.URL+"/emptydirectory/anothernonexistentfile.png")
}

// Test that the ServeIndexForMissing option works
func TestServeIndexForMissing(t *testing.T) {
	fs, err := zipfs.NewZipFileSystem(zipReader, zipfs.ServeIndexForMissing())
	assert.NoError(t, err, "NewZipFileSystem should not result in error")

	// Set up test HTTP server
	fileServer := http.FileServer(fs)
	srv := httptest.NewServer(fileServer)
	defer srv.Close()

	// The root and index.html should still work as expected
	checkGetMatches(t, srv.URL, []byte(rootIndexHTML))
	checkGetMatches(t, srv.URL+"/index.html", []byte(rootIndexHTML))

	// Files that exist should still work
	checkGetMatches(t, srv.URL+"/nonemptydirectory/file.txt", []byte(exampleTxtContent))

	// A file that doesn't exist should give us the content of index.html
	checkGetMatches(t, srv.URL+"/nonexistentfile.txt", []byte(rootIndexHTML))
}
