package models

import "github.com/jinzhu/gorm"

type Retrospective struct {
	//Fields:
	//		TaskProvider, TaskProviderConfig, TimeProvider, TimeProviderConfig
}

type Sprint struct {
	gorm.Model
	//Fields:
	//		SprintID, Retrospective(FK), Title, Status, StartDate, EndDate
	//		Status - Active, Completed, Draft, Deleted.
	// Only 1 Sprint can be active at one time and "StartDate of new Active Sprint" = "EndDate of most recently closed sprint" + 1.
	// i.e. all Closed+Active Sprints should represent a continuous time range.
}

type Task struct {
	gorm.Model
	//Fields:
	//		TaskID, Retrospective(FK), Summary, Type, Status, Estimate, Fields (JSON)
}

type SprintMember struct {
	gorm.Model
	//Fields:
	//		Sprint(FK),Member(User FK), AllocationPercent(int), EfficiencyPercent(int)
}

type SprintMemberTask struct {
	gorm.Model
	//Fields:
	//		Sprint(FK), Member(User FK), Task(FK),TimeSpentMinutes(int),PointsEarned(int),PointAssigned(int),
}
