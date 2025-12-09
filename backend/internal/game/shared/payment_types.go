package shared

// PaymentSubstitute represents a resource type substitution for payments
type PaymentSubstitute struct {
	ResourceType   ResourceType
	ConversionRate int
}

// RequirementModifier represents a modification to requirements
type RequirementModifier struct {
	Amount                int
	AffectedResources     []ResourceType
	CardTarget            *string
	StandardProjectTarget *StandardProject
}
