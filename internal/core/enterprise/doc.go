// Package enterprise provides shared helpers (errors, metrics) for enterprise-gated
// features that expose stubs in community builds. It offers a minimal, stable
// surface so individual stub implementations do not duplicate error strings or
// metric wiring.
package enterprise
