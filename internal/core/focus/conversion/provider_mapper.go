package conversion

// FocusRecordLite is a projection of key parity fields for incremental refactor
// safety checks. Used in parity tests across providers.
type FocusRecordLite struct {
	EffectiveCost  float64
	UsageQuantity  float64
	ProviderName   string
	ServiceName    string
	ChargeCategory string
}
