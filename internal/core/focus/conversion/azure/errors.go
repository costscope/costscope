package azure

import "errors"

// ErrNilHeaderIndex is returned when a mapper is called without a header index.
var ErrNilHeaderIndex = errors.New("azure: nil header index")
