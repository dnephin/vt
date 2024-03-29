package fs

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func assertNotNil(t *testing.T, err error) {
	if err == nil {
		t.Helper()
		t.Fatalf("should have errored, it returned nil")
	}
}

func TestEqualMissingRoot(t *testing.T) {
	err := PathMatchesManifest("/bogus/path/does/not/exist", NewManifest(t))
	assertNotNil(t, err)
	expected := "stat /bogus/path/does/not/exist: no such file or directory"
	if runtime.GOOS == "windows" {
		expected = "CreateFile /bogus/path/does/not/exist"
	}
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("got=%v, want=%v", err, expected)
	}
}

func TestEqualModeMismatch(t *testing.T) {
	dir := NewDir(t, t.Name(), WithMode(0500))
	defer dir.Remove()

	result := PathMatchesManifest(dir.Path(), NewManifest(t))
	assertNotNil(t, result)
	expected := fmtExpected(`directory %s does not match the manifest:
/
  mode: expected drwx------ got dr-x------
`, dir.Path())
	if runtime.GOOS == "windows" {
		expected = fmtExpected(`directory %s does not match the manifest:
\
  mode: expected drwxrwxrwx got dr-xr-xr-x
`, dir.Path())
	}
	assert.Equal(t, result.Error(), expected)
}

func TestEqualRootIsAFile(t *testing.T) {
	file := NewFile(t, t.Name())
	defer file.Remove()

	result := PathMatchesManifest(file.Path(), NewManifest(t))
	assertNotNil(t, result)
	expected := fmt.Sprintf("path %s must be a directory", file.Path())
	assert.Equal(t, result.Error(), expected)
}

func TestEqualSuccess(t *testing.T) {
	dir := NewDir(t, t.Name(), WithMode(0700))
	defer dir.Remove()

	assert.Assert(t, PathMatchesManifest(dir.Path(), NewManifest(t)))
}

func TestEqualDirectoryHasWithExtraFiles(t *testing.T) {
	dir := NewDir(t, t.Name(),
		WithFile("extra1", "content"))
	defer dir.Remove()

	manifest := NewManifest(t, WithFile("file1", "content"))
	result := PathMatchesManifest(dir.Path(), manifest)
	assertNotNil(t, result)
	expected := fmtExpected(`directory %s does not match the manifest:
/
  file1: expected file to exist
  extra1: unexpected file
`, dir.Path())
	assert.Equal(t, result.Error(), expected)
}

func fmtExpected(format string, args ...interface{}) string {
	return filepath.FromSlash(fmt.Sprintf(format, args...))
}

func TestEqualWithMatchAnyFileContent(t *testing.T) {
	dir := NewDir(t, t.Name(),
		WithFile("data", "this is some data"))
	defer dir.Remove()

	expected := NewManifest(t,
		WithFile("data", "different content", MatchAnyFileContent()))
	assert.Assert(t, PathMatchesManifest(dir.Path(), expected))
}

func TestEqualWithFileContent(t *testing.T) {
	dir := NewDir(t, "assert-test-root",
		WithFile("file1", "line1\nline2\nline3"))
	defer dir.Remove()

	manifest := NewManifest(t,
		WithFile("file1", "line2\nline3"))

	result := PathMatchesManifest(dir.Path(), manifest)
	expected := fmtExpected(`directory %s does not match the manifest:
/file1
  content:
    --- expected
    +++ actual
    @@ -1,2 +1,3 @@
    +line1
     line2
     line3
`, dir.Path())
	assert.Equal(t, result.Error(), expected)
}

func TestEqualWithMatchContentIgnoreCarriageReturn(t *testing.T) {
	dir := NewDir(t, t.Name(),
		WithFile("file1", "line1\r\nline2"))
	defer dir.Remove()

	manifest := NewManifest(t,
		WithFile("file1", "line1\nline2", MatchContentIgnoreCarriageReturn()))

	result := PathMatchesManifest(dir.Path(), manifest)
	assert.Assert(t, result == nil)
}

func TestEqualDirectoryWithMatchExtraFiles(t *testing.T) {
	file1 := WithFile("file1", "same in both")
	dir := NewDir(t, t.Name(),
		file1,
		WithFile("extra", "some content"))
	defer dir.Remove()

	expected := NewManifest(t, file1, MatchExtraFiles())
	assert.Assert(t, PathMatchesManifest(dir.Path(), expected))
}

func TestEqualManyFailures(t *testing.T) {
	dir := NewDir(t, t.Name(),
		WithFile("file1", "same in both"),
		WithFile("extra", "some content"),
		WithSymlink("sym1", "extra"))
	defer dir.Remove()

	manifest := NewManifest(t,
		WithDir("subdir",
			WithFile("somefile", "")),
		WithFile("file1", "not the\nsame in both"))

	result := PathMatchesManifest(dir.Path(), manifest)
	assertNotNil(t, result)

	expected := fmtExpected(`directory %s does not match the manifest:
/
  subdir: expected directory to exist
  extra: unexpected file
  sym1: unexpected symlink
/file1
  content:
    --- expected
    +++ actual
    @@ -1,2 +1 @@
    -not the
     same in both
`, dir.Path())
	assert.Equal(t, result.Error(), expected)
}

type cmpFailure interface {
	FailureMessage() string
}

func TestMatchAnyFileMode(t *testing.T) {
	dir := NewDir(t, t.Name(),
		WithFile("data", "content",
			WithMode(0777)))
	defer dir.Remove()

	expected := NewManifest(t,
		WithFile("data", "content", MatchAnyFileMode()))
	assert.Assert(t, PathMatchesManifest(dir.Path(), expected))
}

func TestMatchFileContent(t *testing.T) {
	dir := NewDir(t, t.Name(),
		WithFile("data", "content"))
	defer dir.Remove()

	t.Run("content matches", func(t *testing.T) {
		matcher := func(b []byte) error {
			return nil
		}
		manifest := NewManifest(t,
			WithFile("data", "different", MatchFileContent(matcher)))
		assert.Assert(t, PathMatchesManifest(dir.Path(), manifest))
	})

	t.Run("content does not match", func(t *testing.T) {
		matcher := func(b []byte) error {
			return fmt.Errorf("data content differs from expected")
		}
		manifest := NewManifest(t,
			WithFile("data", "content", MatchFileContent(matcher)))
		result := PathMatchesManifest(dir.Path(), manifest)
		assertNotNil(t, result)

		expected := fmtExpected(`directory %s does not match the manifest:
/data
  content: data content differs from expected
`, dir.Path())
		assert.Equal(t, result.Error(), expected)
	})
}

func TestMatchExtraFilesGlob(t *testing.T) {
	dir := NewDir(t, t.Name(),
		WithFile("t.go", "data"),
		WithFile("a.go", "data"),
		WithFile("conf.yml", "content", WithMode(0600)))
	defer dir.Remove()

	t.Run("matching globs", func(t *testing.T) {
		manifest := NewManifest(t,
			MatchFilesWithGlob("*.go", MatchAnyFileMode(), MatchAnyFileContent()),
			MatchFilesWithGlob("*.yml", MatchAnyFileMode(), MatchAnyFileContent()))
		assert.Assert(t, PathMatchesManifest(dir.Path(), manifest))
	})

	t.Run("matching globs with wrong mode", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("expect mode does not match on windows")
		}
		manifest := NewManifest(t,
			MatchFilesWithGlob("*.go", MatchAnyFileMode(), MatchAnyFileContent()),
			MatchFilesWithGlob("*.yml", MatchAnyFileContent(), WithMode(0700)))

		result := PathMatchesManifest(dir.Path(), manifest)

		assertNotNil(t, result)
		expected := fmtExpected(`directory %s does not match the manifest:
conf.yml
  mode: expected -rwx------ got -rw-------
`, dir.Path())
		assert.Equal(t, result.Error(), expected)
	})

	t.Run("matching partial glob", func(t *testing.T) {
		manifest := NewManifest(t, MatchFilesWithGlob("*.go", MatchAnyFileMode(), MatchAnyFileContent()))
		result := PathMatchesManifest(dir.Path(), manifest)
		assertNotNil(t, result)

		expected := fmtExpected(`directory %s does not match the manifest:
/
  conf.yml: unexpected file
`, dir.Path())
		assert.Equal(t, result.Error(), expected)
	})

	t.Run("invalid glob", func(t *testing.T) {
		manifest := NewManifest(t, MatchFilesWithGlob("[-x]"))
		result := PathMatchesManifest(dir.Path(), manifest)
		assertNotNil(t, result)

		expected := fmtExpected(`directory %s does not match the manifest:
/
  a.go: unexpected file
  conf.yml: unexpected file
  t.go: unexpected file
a.go
  failed to match glob pattern: syntax error in pattern
conf.yml
  failed to match glob pattern: syntax error in pattern
t.go
  failed to match glob pattern: syntax error in pattern
`, dir.Path())
		assert.Equal(t, result.Error(), expected)
	})
}
