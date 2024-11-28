package fs_test

import (
	"os"
	"testing"

	"github.com/dnephin/vt/fs"
	"github.com/dnephin/vt/tma"
)

var t = &testing.T{}

// Create a temporary directory that contains a single file
func ExampleNewDir() {
	dir := fs.NewDir(t, "test-name",
		fs.WithFile("file1", "content\n"))

	files, err := os.ReadDir(dir.Path())
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(files), 1; got != want {
		t.Fatal(tma.GotWant(got, want))
	}
}

// Create a new file with some content
func ExampleNewFile() {
	file := fs.NewFile(t, "test-name",
		fs.WithContent("content\n"), fs.AsUser(0, 0))

	content, err := os.ReadFile(file.Path())
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(content), "content\n"; got != want {
		t.Fatal(tma.GotWant(got, want))
	}
}

// Create a directory and subdirectory with files
func ExampleWithDir() {
	dir := fs.NewDir(t, "test-name",
		fs.WithDir("subdir",
			fs.WithMode(os.FileMode(0700)),
			fs.WithFile("file1", "content\n")),
	)
	_ = dir
}

// Test that a directory contains the expected files, and all the files have the
// expected properties.
func ExampleEqual() {
	path := operationWhichCreatesFiles()
	expected := fs.NewManifest(t,
		fs.WithFile("one", "",
			fs.WithBytes([]byte("content")),
			fs.WithMode(0600)),
		fs.WithDir("data",
			fs.WithFile("config", "", fs.MatchAnyFileContent())))

	if err := fs.PathMatchesManifest(path, expected); err != nil {
		t.Fatal(err)
	}
}

func operationWhichCreatesFiles() string {
	return "example-path"
}

// Add a file to an existing directory
func ExampleApply() {
	dir := fs.NewDir(t, "test-name")
	fs.Apply(t, dir, fs.WithFile("file1", "content\n"))
}
