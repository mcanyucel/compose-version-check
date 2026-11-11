// Package imageref parses container image references such as
// "ghcr.io/immich-app/posgres:14@sha256:abc" into their parts.
package imageref

import (
	"errors"
	"fmt"
	"strings"
)

const defaultRegistry = "docker.io"

// ErrEmpty is returned when the reference is empty or only whitespace.
var ErrEmpty = errors.New("empty image reference")

// Reference is a parsed image reference
type Reference struct {
	Registry   string // "docker.io", "ghcr.io", "registry:5000"
	Repository string // "library/postgres", "immich-app/immich-server"
	Tag        string // "16", "latest"; empty when only a digest is given
	Digest     string // "sha256:..."; empty when none given
}

// Parse splits an image reference into registry, reposioty, tag, and digest,
// filling in Docker's defaults (docker.io, library/, latest).
func Parse(s string) (Reference, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Reference{}, ErrEmpty
	}
	original := s
	var ref Reference

	// 1. Digest: everything after "@"
	if name, digest, found := strings.Cut(s, "@"); found {
		if !strings.Contains(digest, ":") {
			return Reference{}, fmt.Errorf("invalid digest %q in %q", digest, original)
		}
		s, ref.Digest = name, digest
	}

	// 2. Tag: after the last ":", but only if no "/" comes after it.
	//    In "registry:5000/app" that colon belongs to the registry's port.
	if i := strings.LastIndex(s, ":"); i > strings.LastIndex(s, "/") {
		s, ref.Tag = s[:i], s[i+1:]
		if ref.Tag == "" {
			return Reference{}, fmt.Errorf("empty tag in %q", original)
		}
	}

	// 3. Registry: the first path part, if it looks like a host name.
	ref.Registry = defaultRegistry
	if first, rest, found := strings.Cut(s, "/"); found && isRegistry(first) {
		ref.Registry, s = first, rest
	}
	if ref.Registry == "index.docker.io" {
		ref.Registry = defaultRegistry
	}
	if s == "" {
		return Reference{}, fmt.Errorf("missing repository in %q", original)
	}

	// 4. Official Docker Hub images live under "library/".
	if ref.Registry == defaultRegistry && !strings.Contains(s, "/") {
		s = "library/" + s
	}
	ref.Repository = s

	if ref.Tag == "" && ref.Digest == "" {
		ref.Tag = "latest"
	}
	return ref, nil
}

// isRegistry reports whether the first part of a reference is a registry host rather
// than a Docker Hub user or organization
func isRegistry(s string) bool {
	return strings.ContainsAny(s, ".:") || s == "localhost"
}

// Name returns the image without t ag or digest, e.g. "docker.io/library/postgres".
func (r Reference) Name() string {
	return r.Registry + "/" + r.Repository
}

// String returns the full normalized reference
func (r Reference) String() string {
	s := r.Name()
	if r.Tag != "" {
		s += ":" + r.Tag
	}
	if r.Digest != "" {
		s += "@" + r.Digest
	}
	return s
}
