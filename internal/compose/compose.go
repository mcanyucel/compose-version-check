// Package compose reads Compose files: their services, images and the
// x-upstream block, plus .env files.
package compose

import (
	"errors"
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"

	"github.com/mcanyucel/compose-version-check/internal/imageref"
	"github.com/mcanyucel/compose-version-check/internal/interpolate"
)

// Upstream is the x-upstream block: where the maintainer's compose file lives.
type Upstream struct {
	URL    string   `yaml:"url"`
	Env    string   `yaml:"env"`
	Ignore []string `yaml:"ignore"`
}

// Service is the part of a Compose service this tool cares about.
type Service struct {
	Image string `yaml:"image"`
}

// File is the part of a Compose file this tool cares about.
// Everything else in the file (ports, volumes, ...) is ignored.
type File struct {
	Name     string             `yaml:"name"`
	Services map[string]Service `yaml:"services"`
	Upstream *Upstream          `yaml:"x-upstream"` // nil when the block is missing
}

// Parse decodes Compose YAML. A file without services is an error, which also
// catches error pages such as GitHub's "404: Not Found".
func Parse(data []byte) (*File, error) {
	var f File
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse compose file: %w", err)
	}
	if len(f.Services) == 0 {
		return nil, errors.New("compose file has no services")
	}
	return &f, nil
}

// Load reads and parses a Compose file from disk.
func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err // already says which file and why
	}
	f, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return f, nil
}

// Image is a service's image after variables are resolved.
type Image struct {
	Raw string             // as written in the compose file
	Ref imageref.Reference // the parsed image; valid only when Err is nil
	Err error              // unset variable or invalid reference
}

// Images resolves every service's image, taking variable values from env.
// Services without an image (built from source) are left out.
func (f *File) Images(env map[string]string) map[string]Image {
	lookup := func(name string) (string, bool) {
		value, ok := env[name]
		return value, ok
	}

	images := make(map[string]Image, len(f.Services))
	for name, svc := range f.Services {
		if svc.Image == "" {
			continue
		}
		img := Image{Raw: svc.Image}
		expanded, err := interpolate.Expand(svc.Image, lookup)
		if err == nil {
			img.Ref, err = imageref.Parse(expanded)
		}
		img.Err = err
		images[name] = img
	}
	return images
}
