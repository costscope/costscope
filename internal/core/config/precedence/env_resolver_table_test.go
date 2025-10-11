package precedence

import "testing"

// Historical placeholder: EnvResolver has been removed. This file remains only to
// guard against accidental test placement in the precedence package. Real field
// precedence matrix tests live in internal/core/config/ (resolve_field_*_test.go).
func Test_NoOp_Precedence_PackageGuard(t *testing.T) {
	t.Skip("noop in precedence package; real tests in config package")
}
