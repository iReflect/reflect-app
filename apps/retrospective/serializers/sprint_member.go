package serializers

import (
	"github.com/iReflect/reflect-app/apps/retrospective/models"
	"github.com/iReflect/reflect-app/libs/utils"
)

// SprintMemberSummary ...
type SprintMemberSummary struct {
	ID                  uint
	FirstName           string
	LastName            string
	AllocationPercent   float64
	ExpectationPercent  float64
	Vacations           float64
	Rating              uint
	Comment             string
	ActualStoryPoint    float64
	TotalTimeSpentInMin float64
	ExpectedStoryPoint  float64
}

// SetExpectedStoryPoint ...
func (member *SprintMemberSummary) SetExpectedStoryPoint(sprint models.Sprint, retro models.Retrospective) {
	member.ExpectedStoryPoint = utils.CalculateExpectedSP(*sprint.StartDate, *sprint.EndDate,
		member.Vacations, member.ExpectationPercent, member.AllocationPercent, retro.StoryPointPerWeek)
}

// SprintMemberSummaryListSerializer ...
type SprintMemberSummaryListSerializer struct {
	Members []*SprintMemberSummary
}
