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

// SetExpectedVelocity ...
func (member *SprintMemberSummary) SetExpectedVelocity(sprint models.Sprint, retrospective models.Retrospective) {
	sprintWorkingDays := utils.GetWorkingDaysBetweenTwoDates(*sprint.StartDate,
		*sprint.EndDate, true)
	memberWorkingDays := float64(sprintWorkingDays) - member.Vacations
	expectationCoefficient := member.ExpectationPercent / 100.00
	allocationCoefficient := member.AllocationPercent / 100.00
	storyPointPerDay := retrospective.StoryPointPerWeek / 5
	member.ExpectedStoryPoint = memberWorkingDays * storyPointPerDay *
		expectationCoefficient * allocationCoefficient
}

// SprintMemberSummaryListSerializer ...
type SprintMemberSummaryListSerializer struct {
	Members []*SprintMemberSummary
}
