/*
Package fs provides tools for creating temporary files, and testing the
contents and structure of a directory.
*/
package fs

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dnephin/vt/internal/cleanup"
)

// Path objects return their filesystem path. Path may be implemented by a
// real filesystem object (such as File and Dir) or by a type which updates
// entries in a Manifest.
type Path interface {
	Path() string
	Remove()
}

var (
	_ Path = &Dir{}
	_ Path = &File{}
)

// File is a temporary file on the filesystem
type File struct {
	path string
}

type helperT interface {
	Helper()
}

type TestingT interface {
	Fatalf(string, ...any)
	Log(...interface{})
}

// NewFile creates a new file in a temporary directory using prefix as part of
// the filename. The PathOps are applied to the before returning the File.
//
// When used with Go 1.14+ the file will be automatically removed when the test
// ends, unless the TEST_NOCLEANUP env var is set to true.
func NewFile(t TestingT, prefix string, ops ...PathOp) *File {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	tempfile, err := os.CreateTemp("", cleanPrefix(prefix)+"-")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	fh := &File{path: tempfile.Name()}
	cleanup.Cleanup(t, fh.Remove)

	if err := tempfile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}
	if err := applyPathOps(fh, ops); err != nil {
		t.Fatalf("failed to apply operations to file: %v", err)
	}
	return fh
}

func cleanPrefix(prefix string) string {
	// windows requires both / and \ are replaced
	if runtime.GOOS == "windows" {
		prefix = strings.Replace(prefix, string(os.PathSeparator), "-", -1)
	}
	return strings.Replace(prefix, "/", "-", -1)
}

// Path returns the full path to the file
func (f *File) Path() string {
	return f.path
}

// Remove the file
func (f *File) Remove() {
	_ = os.Remove(f.path)
}

// Dir is a temporary directory
type Dir struct {
	path string
}

// NewDir returns a new temporary directory using prefix as part of the directory
// name. The PathOps are applied before returning the Dir.
//
// When used with Go 1.14+ the directory will be automatically removed when the test
// ends, unless the TEST_NOCLEANUP env var is set to true.
func NewDir(t TestingT, prefix string, ops ...PathOp) *Dir {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	path, err := os.MkdirTemp("", cleanPrefix(prefix)+"-")
	if err != nil {
		t.Fatalf("failed to make temp dir: %v", err)
	}
	dir := &Dir{path: path}
	cleanup.Cleanup(t, dir.Remove)

	if err := applyPathOps(dir, ops); err != nil {
		t.Fatalf("failed to apply operations: %v", err)
	}
	return dir
}

// Path returns the full path to the directory
func (d *Dir) Path() string {
	return d.path
}

// Remove the directory
func (d *Dir) Remove() {
	_ = os.RemoveAll(d.path)
}

// Join returns a new path with this directory as the base of the path
func (d *Dir) Join(parts ...string) string {
	return filepath.Join(append([]string{d.Path()}, parts...)...)
}

// DirFromPath returns a Dir for a path that already exists. No directory is created.
// Unlike NewDir the directory will not be removed automatically when the test exits,
// it is the callers responsibly to remove the directory.
// DirFromPath can be used with Apply to modify an existing directory.
//
// If the path does not already exist, use NewDir instead.
func DirFromPath(t TestingT, path string, ops ...PathOp) *Dir {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}

	dir := &Dir{path: path}
	if err := applyPathOps(dir, ops); err != nil {
		t.Fatalf("failed to apply operations: %v", err)
	}
	return dir
}
