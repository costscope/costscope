internal/testutil
==================

This package provides a small, centralized helper for resolving the repository root (the
directory that contains `go.mod`). The helper exists in two forms to make it convenient for
both tests and non-test code (command-line tools, scripts).

APIs
----

- `RepoRoot() (string, error)`
  - Non-test API. Returns the repository root path (directory containing `go.mod`) or an error.
  - Use this from main packages, scripts under `cmd/` or `scripts/` where importing `testing` is
    not allowed.

- `FindRepoRoot(t *testing.T) string`
  - Test-friendly wrapper. Calls `RepoRoot()` and fails the test (`t.Fatalf`) on error.
  - Use this from `_test.go` files to keep tests concise.

Rationale
---------

Project code historically contained small ad-hoc helpers that walked up from the current
working directory to locate the repo root. That duplication caused symbol collisions and
made it error-prone to reuse the behavior from non-test code. Centralizing the logic here
reduces duplication and makes the behaviour consistent across tests and tools.

Examples
--------

In a test:

```go
func TestSomething(t *testing.T) {
    root := testutil.FindRepoRoot(t)
    // use `root` to locate fixtures under `demo/` or `tests/`.
}
```

In a command/tool (non-test code):

```go
root, err := testutil.RepoRoot()
if err != nil {
    fmt.Fprintf(os.Stderr, "cannot locate repo root: %v\n", err)
    os.Exit(2)
}
// use `root` to find templates or demo assets during generation
```

Notes
-----
- The function looks for `go.mod` by walking up parent directories starting from the
  current working directory. If you run tools from a different CWD, `RepoRoot()` will try to
  find the repository root relative to that CWD.
- Prefer `RepoRoot()` in non-test code to avoid importing `testing`.
