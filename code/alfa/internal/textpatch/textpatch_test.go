package textpatch_test

import (
	"reflect"
	"testing"

	"github.com/tenzoki/agen/alfa/internal/textpatch"
)

func TestPatchLines_basic(t *testing.T) {
	lines := []string{"one", "two", "three", "four"}
	patch := `[
        {"line": 1, "type": "replace", "content": ["new two", "new three"]},
        {"line": 3, "type": "delete"},
        {"line": 3, "type": "insert", "content": ["inserted line"]}
    ]`
	want := []string{"one", "new two", "new three", "inserted line", "four"}
	got, err := textpatch.PatchLines(lines, patch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("PatchLines got %q, want %q", got, want)
	}
}

func TestPatchLines_insert_delete(t *testing.T) {
	orig := []string{"a", "b", "c"}
	patch := `[{"line":1, "type":"delete"}, {"line":2, "type":"insert", "content":["X"]}]`
	want := []string{"a", "c", "X"}
	got, err := textpatch.PatchLines(orig, patch)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestPatchText(t *testing.T) {
	in := "a\nb\nc"
	patch := `[{"line":1, "type":"replace", "content":["hello", "world"]}]`
	want := "a\nhello\nworld\nc"
	got, err := textpatch.PatchText(in, patch)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
