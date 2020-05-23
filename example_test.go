package zipfs_test

import (
	"archive/zip"
	"log"
	"net/http"
	"os"

	"gerace.dev/zipfs"
)

func ExampleNewEmbeddedZipFileSystem() {
	fs, err := zipfs.NewEmbeddedZipFileSystem(zipfs.ServeIndexForMissing())
	if err != nil {
		log.Fatalf("Error setting up zipfs: %v", err)
	}

	log.Fatal(http.ListenAndServe("127.0.0.1:8080", http.FileServer(fs)))
}

func ExampleNewZipFileSystem() {
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

func ExampleServeIndexForMissing() {
	fs, err := zipfs.NewEmbeddedZipFileSystem(zipfs.ServeIndexForMissing())
	if err != nil {
		log.Fatalf("Error setting up zipfs: %v", err)
	}

	log.Fatal(http.ListenAndServe("127.0.0.1:8080", http.FileServer(fs)))
}
