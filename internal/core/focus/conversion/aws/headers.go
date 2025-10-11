package aws

import "fmt"

// HeaderIndex stores column indices for fast extraction.
// Mirrors the legacy awsHeaderIndex layout.
type HeaderIndex struct {
	// Required
	IBillingAccountId   int
	IBillingAccountName int
	IBillingCurrency    int
	IUnblendedCost      int
	IUsageAmount        int
	IUsageStartDate     int
	IUsageEndDate       int
	ILineItemDesc       int
	IOperation          int
	IUsageType          int
	IProductName        int
	IProductFamily      int
	ILineItemType       int
	IPriceId            int
	IUsageAccountId     int
	IResourceId         int

	// Commitment-related (optional)
	ISPArn            int
	ISPId             int
	IRIArn            int
	IRISubscriptionId int

	// Optional
	IAvailabilityZone int
	IRegion           int

	// Total columns to validate row length
	Cols int
}

// NewHeaderIndex builds the index map for required/optional columns.
func NewHeaderIndex(headers []string) (*HeaderIndex, error) { // nolint:funlen
	pos := make(map[string]int, len(headers))
	for i, h := range headers {
		pos[h] = i
	}
	must := func(name string) (int, error) {
		if v, ok := pos[name]; ok {
			return v, nil
		}
		return -1, fmt.Errorf("missing required AWS CUR column: %s", name)
	}
	// Required minimal
	iUnblendedCost, err := must("lineItem/UnblendedCost")
	if err != nil {
		return nil, err
	}
	iUsageStartDate, err := must("lineItem/UsageStartDate")
	if err != nil {
		return nil, err
	}
	iUsageEndDate, err := must("lineItem/UsageEndDate")
	if err != nil {
		return nil, err
	}
	iProductName, err := must("product/ProductName")
	if err != nil {
		return nil, err
	}
	iUsageAccountId, err := must("lineItem/UsageAccountId")
	if err != nil {
		return nil, err
	}

	opt := func(k string) int {
		if v, ok := pos[k]; ok {
			return v
		}
		return -1
	}

	idx := &HeaderIndex{
		IBillingAccountId:   opt("bill/BillingAccountId"),
		IBillingAccountName: opt("bill/BillingAccountName"),
		IBillingCurrency:    opt("bill/BillingCurrency"),
		IUnblendedCost:      iUnblendedCost,
		IUsageAmount:        opt("lineItem/UsageAmount"),
		IUsageStartDate:     iUsageStartDate,
		IUsageEndDate:       iUsageEndDate,
		ILineItemDesc:       opt("lineItem/LineItemDescription"),
		IOperation:          opt("lineItem/Operation"),
		IUsageType:          opt("lineItem/UsageType"),
		IProductName:        iProductName,
		IProductFamily:      opt("product/ProductFamily"),
		ILineItemType:       opt("lineItem/LineItemType"),
		IPriceId:            opt("pricing/PriceId"),
		IUsageAccountId:     iUsageAccountId,
		IResourceId:         opt("lineItem/ResourceId"),
		ISPArn:              opt("savingsPlan/SavingsPlanArn"),
		ISPId:               opt("savingsPlan/SavingsPlanId"),
		IRIArn:              opt("reservation/ReservationARN"),
		IRISubscriptionId:   opt("reservation/SubscriptionId"),
		IAvailabilityZone:   opt("lineItem/AvailabilityZone"),
		IRegion:             opt("product/Region"),
		Cols:                len(headers),
	}
	return idx, nil
}
