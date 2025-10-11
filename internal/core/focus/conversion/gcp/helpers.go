package gcp

import (
	c "local/costscope/internal/core/focus/conversion/common"
)

func unitPrice(cost float64, qty float64) float64 {
	if qty == 0 {
		return 0
	}
	return cost / qty
}

// firstNonEmpty is deprecated; use common.FirstNonEmpty
func firstNonEmpty(ss ...string) string { return c.FirstNonEmpty(ss...) }

func firstFloatNonZero(vals ...float64) float64 {
	for _, v := range vals {
		if v != 0 {
			return v
		}
	}
	return 0
}
