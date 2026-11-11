package compose

import (
	"maps"
	"testing"
)

func TestParseEnv(t *testing.T) {
	data := []byte(`
# a comment
export TAG=16
NAME = immich
QUOTED="hello world"
SINGLE='it is # not a comment'
INLINE=value # a comment
EMPTY=
`)
	got, err := ParseEnv(data)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"TAG":    "16",
		"NAME":   "immich",
		"QUOTED": "hello world",
		"SINGLE": "it is # not a comment",
		"INLINE": "value",
		"EMPTY":  "",
	}
	if !maps.Equal(got, want) {
		t.Errorf("ParseEnv:\n got:  %v\n want: %v", got, want)
	}
}

func TestParseEnvRejectsErrorPage(t *testing.T) {
	if _, err := ParseEnv([]byte("404: Not Found")); err == nil {
		t.Error("ParseEnv accepted an error page, want an error")
	}
}

func TestLoadEnvMissingFile(t *testing.T) {
	env, err := LoadEnv("testdata/does-not-exist.env")
	if err != nil {
		t.Fatalf("got error %v, want none", err)
	}
	if len(env) != 0 {
		t.Errorf("got %v, want an empty map", env)
	}
}
