package aws

// Duplicated constants to avoid importing the root conversion package while we
// finish decoupling. Values must match legacy implementation for parity.
const (
	ctypeSavingsPlan      = "SavingsPlan"
	ctypeReservedInstance = "ReservedInstance"
	curOpCredit           = "Credit"
)
