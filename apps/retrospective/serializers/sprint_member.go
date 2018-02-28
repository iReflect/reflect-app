package serializers

// SprintMemberSummary ...
type SprintMemberSummary struct {
	ID                 uint
	FirstName          string
	LastName           string
	AllocationPercent  float64
	ExpectationPercent float64
	Vacations          float64
	Rating             uint
	Comment            string
	ActualVelocity     float64
	ExpectedVelocity   float64
}

// SprintMemberSummaryListSerializer ...
type SprintMemberSummaryListSerializer struct {
	Members []*SprintMemberSummary
}
