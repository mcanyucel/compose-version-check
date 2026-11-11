package imageref

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		in   string
		want Reference
	}{
		{"postgres", Reference{"docker.io", "library/postgres", "latest", ""}},
		{"postgres:16", Reference{"docker.io", "library/postgres", "16", ""}},
		{"redis/redis-stack:7.2.0-v10", Reference{"docker.io", "redis/redis-stack", "7.2.0-v10", ""}},
		{"ghcr.io/immich-app/immich-server:v2.1.0", Reference{"ghcr.io", "immich-app/immich-server", "v2.1.0", ""}},
		{"registry:5000/app", Reference{"registry:5000", "app", "latest", ""}},
		{"localhost/app:dev", Reference{"localhost", "app", "dev", ""}},
		{"index.docker.io/library/nginx:1.27", Reference{"docker.io", "library/nginx", "1.27", ""}},
		{"postgres@sha256:abc123", Reference{"docker.io", "library/postgres", "", "sha256:abc123"}},
		{"ghcr.io/immich-app/postgres:14-vectorchord0.4.3@sha256:abc123", Reference{"ghcr.io", "immich-app/postgres", "14-vectorchord0.4.3", "sha256:abc123"}},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := Parse(tt.in)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("Parse(%q)\n got:  %+v\n want: %+v", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	for _, in := range []string{"", "   ", "postgres:", "postgres@nodigest", ":16"} {
		t.Run(in, func(t *testing.T) {
			if _, err := Parse(in); err == nil {
				t.Errorf("Parse(%q) returned no error, want one", in)
			}
		})
	}
}

func TestString(t *testing.T) {
	ref, err := Parse("ghcr.io/immich-app/postgres:14@sha256:abc")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := ref.String(), "ghcr.io/immich-app/postgres:14@sha256:abc"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	ref, _ = Parse("postgres")
	if got, want := ref.String(), "docker.io/library/postgres:latest"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
