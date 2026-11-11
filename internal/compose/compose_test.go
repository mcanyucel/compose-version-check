package compose

import (
	"errors"
	"testing"

	"github.com/mcanyucel/compose-version-check/internal/interpolate"
)

func TestParse(t *testing.T) {
	data := []byte(`
name: blog
x-upstream:
  url: https://example.com/compose.yaml
  ignore: [cache]
services:
  web:
    image: ghost:5
  db:
    image: mysql:8.0
  worker:
    build: ./worker
`)
	f, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if f.Name != "blog" {
		t.Errorf("Name = %q, want %q", f.Name, "blog")
	}
	if f.Upstream == nil || f.Upstream.URL != "https://example.com/compose.yaml" {
		t.Errorf("Upstream = %+v, want URL https://example.com/compose.yaml", f.Upstream)
	}
	if len(f.Services) != 3 {
		t.Errorf("got %d services, want 3", len(f.Services))
	}
	if got := f.Services["db"].Image; got != "mysql:8.0" {
		t.Errorf("db image = %q, want %q", got, "mysql:8.0")
	}
}

func TestParseErrors(t *testing.T) {
	tests := map[string]string{
		"github 404 page": "404: Not Found",
		"no services":     "name: empty\n",
		"empty services":  "services: {}\n",
		"broken yaml":     "services: [\n",
	}
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(data)); err == nil {
				t.Errorf("Parse(%q) returned no error, want one", data)
			}
		})
	}
}

func TestImmich(t *testing.T) {
	f, err := Load("testdata/immich/docker-compose.yml")
	if err != nil {
		t.Fatal(err)
	}
	env, err := LoadEnv("testdata/immich/immich.env")
	if err != nil {
		t.Fatal(err)
	}
	images := f.Images(env)

	tests := map[string]string{
		"immich-server":           "ghcr.io/immich-app/immich-server:v2.1.0",
		"immich-machine-learning": "ghcr.io/immich-app/immich-machine-learning:v2.1.0",
		"redis":                   "docker.io/valkey/valkey:8-bookworm@sha256:fec42f39",
		"database":                "ghcr.io/immich-app/postgres:14-vectorchord0.4.3-pgvectors0.2.0@sha256:41eacbe8",
	}
	for service, want := range tests {
		img, ok := images[service]
		if !ok {
			t.Errorf("no image for service %q", service)
			continue
		}
		if img.Err != nil {
			t.Errorf("%s: %v", service, img.Err)
			continue
		}
		if got := img.Ref.String(); got != want {
			t.Errorf("%s = %q, want %q", service, got, want)
		}
	}
}

func TestImagesUnsetVariable(t *testing.T) {
	f, err := Parse([]byte("services:\n  app:\n    image: app:${APP_TAG}\n"))
	if err != nil {
		t.Fatal(err)
	}
	img := f.Images(nil)["app"]

	var unset *interpolate.UnsetError
	if !errors.As(img.Err, &unset) || unset.Name != "APP_TAG" {
		t.Errorf("Err = %v, want an UnsetError for APP_TAG", img.Err)
	}
	if img.Raw != "app:${APP_TAG}" {
		t.Errorf("Raw = %q, want the original text", img.Raw)
	}
}
