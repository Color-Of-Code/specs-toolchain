// Package framework provides the embedded framework template that ships
// inside the specs binary. It is used by `specs framework seed` to
// pre-populate a new framework directory without any network access.
package framework

import "embed"

// Template is the embedded filesystem rooted at the template/ directory.
// It contains the minimal framework skeleton (templates/, lint/, process/,
// skills/, agents/, tools-manifest.yaml).
//
//go:embed all:template
var Template embed.FS
