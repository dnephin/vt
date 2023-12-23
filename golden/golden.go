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
	"fmt"
	"os"
	"path/filepath"

	"gotest.tools/v3/internal/format"
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
		return fmt.Errorf("read wantFilename: %w", err)
	}
	if bytes.Equal([]byte(got), want) {
		return nil
	}

	gotHash := hash(got)
	if update.Requested(gotHash) {
		if dir := filepath.Dir(wantFilename); dir != "." {
			_ = os.MkdirAll(dir, 0755)
		}
		if err := os.WriteFile(wantFilename, []byte(got), 0644); err != nil {
			return fmt.Errorf("write wantfilename: %v", err)
		}
		return nil
	}

	diff := format.UnifiedDiff(format.DiffConfig{
		A:    got,
		B:    string(want),
		From: "got",
		To:   "want",
	})
	msg := "%v\nRun 'go test . -update=%v' to update %s to the new value."
	return fmt.Errorf(msg, diff, gotHash, wantFilename)
}

func hash(got string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(got)))[:10]
}
