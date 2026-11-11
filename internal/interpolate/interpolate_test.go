package interpolate

import (
	"errors"
	"testing"
)

// mapLookup turns a map into a LookupFunc.
func mapLookup(m map[string]string) LookupFunc {
	return func(name string) (string, bool) {
		value, ok := m[name]
		return value, ok
	}
}

var env = map[string]string{
	"TAG":   "16",
	"EMPTY": "",
	"REG":   "ghcr.io",
}

func TestExpand(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"postgres:16", "postgres:16"},
		{"postgres:${TAG}", "postgres:16"},
		{"postgres:$TAG", "postgres:16"},
		{"postgres:${MISSING:-15}", "postgres:15"},
		{"postgres:${EMPTY:-15}", "postgres:15"},
		{"postgres:${EMPTY-15}", "postgres:"},
		{"postgres:${MISSING-15}", "postgres:15"},
		{"${REG}/app:${TAG:+v$TAG}", "ghcr.io/app:v16"},
		{"${MISSING:-${REG}}/app", "ghcr.io/app"},
		{"${TAG:-${MISSING}}", "16"},
		{"price: $$5", "price: $5"},
		{"ends with $", "ends with $"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := Expand(tt.in, mapLookup(env))
			if err != nil {
				t.Fatalf("Expand(%q) returned error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("Expand(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestExpandErrors(t *testing.T) {
	for _, in := range []string{"${MISSING}", "$MISSING", "${TAG", "${1BAD}", "${TAG:x}"} {
		t.Run(in, func(t *testing.T) {
			if _, err := Expand(in, mapLookup(env)); err == nil {
				t.Errorf("Expand(%q) returned no error, want one", in)
			}
		})
	}
}

func TestUnsetError(t *testing.T) {
	_, err := Expand("${DB_TAG:?set it in .env}", mapLookup(nil))

	var unset *UnsetError
	if !errors.As(err, &unset) {
		t.Fatalf("got error %v, want an *UnsetError", err)
	}
	if unset.Name != "DB_TAG" || unset.Message != "set it in .env" {
		t.Errorf("got %+v, want Name DB_TAG and Message %q", unset, "set it in .env")
	}
}
