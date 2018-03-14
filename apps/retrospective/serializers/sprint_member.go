package serializers

import (
	"github.com/iReflect/reflect-app/apps/retrospective/models"
	"github.com/iReflect/reflect-app/libs/utils"
)

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
	TotalTimeSpentMinutes     float64
	ExpectedVelocity   float64
}

func (member *SprintMemberSummary) SetExpectedVelocity(sprint models.Sprint, retrospective models.Retrospective) {
	sprintWorkingDays := utils.GetWorkingDaysBetweenTwoDates(*sprint.StartDate,
		*sprint.EndDate, true)
	memberWorkingDays := float64(sprintWorkingDays) - member.Vacations
	expectationCoefficient := member.ExpectationPercent / 100.00
	allocationCoefficient := member.AllocationPercent / 100.00
	storyPointPerDay := retrospective.StoryPointPerWeek / 5
	member.ExpectedVelocity = memberWorkingDays * storyPointPerDay *
		expectationCoefficient * allocationCoefficient
}

// SprintMemberSummaryListSerializer ...
type SprintMemberSummaryListSerializer struct {
	Members []*SprintMemberSummary
}
