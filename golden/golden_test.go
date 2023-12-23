package golden

import (
	"os"
	"strings"
	"testing"
)

func TestString_WithUpdate(t *testing.T) {
	patch(t, &update, "yes")
	filename := setupGoldenFile(t, "foo")

	err := MatchStringToFile("new value", filename)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(content), "new value"; got != want {
		t.Fatalf("MatchStringToFile(): got=\n%v\n, want=\n%v\n", got, want)
	}

}

func TestString_NotEqual(t *testing.T) {
	patch(t, &update, "no")
	filename := setupGoldenFile(t, "this is\nthe text")

	err := MatchStringToFile("this is\nnot the text", filename)
	if err == nil {
		t.Fatal("MatchStringToFile(): expected an error, got nil")
	}
	want := `
--- got
+++ want
@@ -1,2 +1,2 @@
 this is
-not the text
+the text

Run 'go test
`
	if got := err.Error(); strings.Contains(got, want) {
		t.Fatalf("MatchStringToFile(): got\n%v\n, want\n%v\n", got, want)
	}
}

func setupGoldenFile(t *testing.T, content string) string {
	f, err := os.CreateTemp(t.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = f.Close()
	})

	_, err = f.Write([]byte(content))
	if err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

func patch[T any](t *testing.T, dest *T, stub T) {
	orig := *dest
	*dest = stub
	t.Cleanup(func() {
		*dest = orig
	})
}
