package enterprise

import (
	"errors"
	"fmt"
)

// ErrEnterpriseFeatureDisabled is the sentinel error returned when a caller
// invokes an enterprise-only feature in a non-enterprise build.
var ErrEnterpriseFeatureDisabled = errors.New("enterprise feature disabled: rebuild with -tags enterprise")

// DisabledError returns a wrapped sentinel error including the feature name.
func DisabledError(feature string) error {
	return fmt.Errorf("%w (%s)", ErrEnterpriseFeatureDisabled, feature)
}

// IsDisabled returns true if the error (possibly wrapped) signals a disabled enterprise feature.
func IsDisabled(err error) bool { return errors.Is(err, ErrEnterpriseFeatureDisabled) }
