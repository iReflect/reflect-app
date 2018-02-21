package serializers

import userSerializers "github.com/iReflect/reflect-app/apps/user/serializers"

// SprintMemberSummary ...
type SprintMemberSummary struct {
	userSerializers.User
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
