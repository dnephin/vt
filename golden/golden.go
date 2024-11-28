/*
Package golden provides tools for comparing large mutli-line strings.

Golden files are files in the ./testdata/ subdirectory of the package under test.
Golden files can be automatically updated to match new values by running
`go test pkgname -update`. To ensure the update is correct
compare the diff of the old expected value to the new expected value.
*/
package golden

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dnephin/vt/internal/format"
)

// MatchStringToFile compares got to the contents of wantFile and returns nil if the
// strings are equal, otherwise returns a unified diff of the values. Whitespace
// only changes will be highlighted using visible characters.
//
// Running `go test pkgname -update` will write the value of actual
// to the golden file.
func MatchStringToFile(got string, wantFilename string) error {
	want, err := os.ReadFile(wantFilename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) && update.Requested(hash(got)) {
			return updateFile(got, wantFilename)
		}
		return fmt.Errorf("read wantFilename: %w", err)
	}
	if bytes.Equal([]byte(got), want) {
		return nil
	}

	gotHash := hash(got)
	if update.Requested(gotHash) {
		return updateFile(got, wantFilename)
	}

	diff := format.UnifiedDiff(format.DiffConfig{
		A:    got,
		B:    string(want),
		From: "got",
		To:   "want",
	})
	pkg, _ := currentTestName()
	msg := "(-got +want):\n%v\nRun 'go test %v -update=%v' to update %s to the new value."
	return fmt.Errorf(msg, diff, pkg, gotHash, wantFilename)
}

func hash(got string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(got)))[:10]
}

func updateFile(got string, wantFilename string) error {
	if dir := filepath.Dir(wantFilename); dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}
	if err := os.WriteFile(wantFilename, []byte(got), 0644); err != nil {
		return fmt.Errorf("write wantfilename: %v", err)
	}
	return nil
}

func currentTestName() (pkg string, test string) {
	pc, _, _, _ := runtime.Caller(2) // currentTestName + MatchStringToFile
	name := runtime.FuncForPC(pc).Name()
	if i := strings.LastIndex(name, "."); i >= 0 {
		return name[:i], name[i+1:]
	}
	return "", ""
}
