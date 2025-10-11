package azure

import "strings"

// HeaderIndex captures column positions for Azure export CSV/JSON headers.
type HeaderIndex struct {
	BillingAccountId   int
	EnrollmentNumber   int
	BillingProfileId   int
	BillingAccountName int
	BillingProfileName int
	BillingCurrency    int
	Currency           int

	SubscriptionId   int
	SubscriptionName int
	ResourceId       int
	ResourceName     int
	InstanceName     int
	ResourceType     int
	ServiceTier      int

	ServiceFamily  int
	MeterCategory  int
	ServiceName    int
	Product        int
	MeterName      int
	MeterSubCat    int
	MeterId        int
	SkuId          int
	PartNumber     int
	ProductOrderNo int

	Quantity       int
	UnitOfMeasure  int
	MeterUnit      int
	RetailPrice    int
	UnitPrice      int
	PricingModel   int
	PricingModelNm int

	AmortizedCost         int
	CostInBillingCurrency int
	Cost                  int
	CostInUSD             int
	CostLikeIdx           []int

	UsageStart int
	UsageEnd   int
	Date       int
	UsageDate  int

	ChargeType  int
	BillingType int

	BenefitType     int
	BenefitId       int
	BenefitName     int
	ReservationId   int
	ReservationName int
	SavingsPlanId   int
	SavingsPlanName int

	Tags             int
	ResourceLocation int
	Location         int
}

// NewHeaderIndex builds a HeaderIndex by scanning header names case-insensitively.
func NewHeaderIndex(headers []string) *HeaderIndex {
	idx := func(n string) int { return colIndex(headers, n) }
	ai := &HeaderIndex{
		BillingAccountId:      idx("BillingAccountId"),
		EnrollmentNumber:      idx("EnrollmentNumber"),
		BillingProfileId:      idx("BillingProfileId"),
		BillingAccountName:    idx("BillingAccountName"),
		BillingProfileName:    idx("BillingProfileName"),
		BillingCurrency:       idx("BillingCurrency"),
		Currency:              idx("Currency"),
		SubscriptionId:        idx("SubscriptionId"),
		SubscriptionName:      idx("SubscriptionName"),
		ResourceId:            idx("ResourceId"),
		ResourceName:          idx("ResourceName"),
		InstanceName:          idx("InstanceName"),
		ResourceType:          idx("ResourceType"),
		ServiceTier:           idx("ServiceTier"),
		ServiceFamily:         idx("ServiceFamily"),
		MeterCategory:         idx("MeterCategory"),
		ServiceName:           idx("ServiceName"),
		Product:               idx("Product"),
		MeterName:             idx("MeterName"),
		MeterSubCat:           idx("MeterSubCategory"),
		MeterId:               idx("MeterId"),
		SkuId:                 idx("SkuId"),
		PartNumber:            idx("PartNumber"),
		ProductOrderNo:        idx("ProductOrderNumber"),
		Quantity:              idx("Quantity"),
		UnitOfMeasure:         idx("UnitOfMeasure"),
		MeterUnit:             idx("MeterUnit"),
		RetailPrice:           idx("RetailPrice"),
		UnitPrice:             idx("UnitPrice"),
		PricingModel:          idx("PricingModel"),
		PricingModelNm:        idx("PricingModelName"),
		AmortizedCost:         idx("AmortizedCost"),
		CostInBillingCurrency: idx("CostInBillingCurrency"),
		Cost:                  idx("Cost"),
		CostInUSD:             idx("CostInUSD"),
		UsageStart:            idx("UsageStart"),
		UsageEnd:              idx("UsageEnd"),
		Date:                  idx("Date"),
		UsageDate:             idx("UsageDate"),
		ChargeType:            idx("ChargeType"),
		BillingType:           idx("BillingType"),
		BenefitType:           idx("BenefitType"),
		BenefitId:             idx("BenefitId"),
		BenefitName:           idx("BenefitName"),
		ReservationId:         idx("ReservationId"),
		ReservationName:       idx("ReservationName"),
		SavingsPlanId:         idx("SavingsPlanId"),
		SavingsPlanName:       idx("SavingsPlanName"),
		Tags:                  idx("Tags"),
		ResourceLocation:      idx("ResourceLocation"),
		Location:              idx("Location"),
	}
	for i, h := range headers { // collect *cost* columns
		if strings.Contains(strings.ToLower(strings.TrimSpace(h)), "cost") {
			ai.CostLikeIdx = append(ai.CostLikeIdx, i)
		}
	}
	return ai
}

func colIndex(headers []string, name string) int {
	for i, h := range headers {
		if strings.EqualFold(strings.TrimSpace(h), name) {
			return i
		}
	}
	return -1
}

// Get returns r[i] or empty string if out of bounds.
func Get(r []string, i int) string {
	if i < 0 || i >= len(r) {
		return ""
	}
	return r[i]
}

// FirstNonEmpty returns the first non-empty trimmed string at the provided indexes.
func FirstNonEmpty(r []string, idxs ...int) string {
	for _, i := range idxs {
		if s := strings.TrimSpace(Get(r, i)); s != "" {
			return s
		}
	}
	return ""
}
