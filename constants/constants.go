package constants

// CustomDateFormat is a date format used in the application to parse date string/object into a usable format
const CustomDateFormat = "2006-01-02"

// ActionItemType is datatype for ActionItems's
type ActionItemType string

// constants defined to use in ActionItemTypeMap
const (
	Retrospective         ActionItemType = "Retrospective"
	RetrospectiveFeedback ActionItemType = "RetrospectiveFeedback"
	SprintMember          ActionItemType = "SprintMember"
	SprintMemberTask      ActionItemType = "SprintMemberTask"
	Sprint                ActionItemType = "Sprint"
	SprintTask            ActionItemType = "SprintTask"
)

// ActionItemTypeMap is types of ActionItem of Trail model used in adding trails.
var ActionItemTypeMap = map[ActionItemType]string{
	Retrospective:         "Retrospective",
	RetrospectiveFeedback: "Retrospective Feedback",
	SprintMember:          "Sprint Member",
	SprintMemberTask:      "Sprint Member Task",
	Sprint:                "Sprint",
	SprintTask:            "Sprint Task",
}

// ActionType special data type for action
type ActionType string

// constants defined to use in ActionTypeMap
const (
	CreatedRetrospective    ActionType = "CreatedRetrospective"
	AddedGoal               ActionType = "AddedGoal"
	UpdatedGoal             ActionType = "UpdatedGoal"
	ResolvedGoal            ActionType = "ResolvedGoal"
	UnresolvedGoal          ActionType = "UnresolvedGoal"
	AddedHighlight          ActionType = "AddedHighlight"
	UpdatedHighlight        ActionType = "UpdatedHighlight"
	AddedSprintMember       ActionType = "AddedSprintMember"
	UpdatedSprintMember     ActionType = "UpdatedSprintMember"
	RemovedSprintMember     ActionType = "RemovedSprintMember"
	AddedNote               ActionType = "AddedNote"
	UpdatedNote             ActionType = "UpdatedNote"
	AddedSprintMemberTask   ActionType = "AddedSprintMemberTask"
	UpdatedSprintMemberTask ActionType = "UpdatedSprintMemberTask"
	CreatedSprint           ActionType = "CreatedSprint"
	DeletedSprint           ActionType = "DeletedSprint"
	UpdatedSprint           ActionType = "UpdatedSprint"
	ActivatedSprint         ActionType = "ActivatedSprint"
	FreezeSprint            ActionType = "FreezeSprint"
	TriggeredSprintRefresh  ActionType = "TriggeredSprintRefresh"
	UpdatedSprintTask       ActionType = "UpdatedSprintTask"
	MarkDoneSprintTask      ActionType = "MarkDoneSprintTask"
	MarkUndoneSprintTask    ActionType = "MarkUndoneSprintTask"
)

// ActionTypeMap is types of Action of Trail model used in adding trails.
var ActionTypeMap = map[ActionType]string{
	CreatedRetrospective:    "Created Retrospective",
	AddedGoal:               "Added a Goal",
	UpdatedGoal:             "Updated a Goal",
	ResolvedGoal:            "Marked a goal resolved",
	UnresolvedGoal:          "Marked a goal unresolved",
	AddedHighlight:          "Added a highlight",
	UpdatedHighlight:        "Updated a highlight",
	AddedSprintMember:       "Added member in sprint",
	UpdatedSprintMember:     "Updated members in sprint",
	RemovedSprintMember:     "Removed member from sprint",
	AddedNote:               "Added a note",
	UpdatedNote:             "Updated a note",
	AddedSprintMemberTask:   "Added a member on task in sprint",
	UpdatedSprintMemberTask: "Updated member on task in sprint",
	CreatedSprint:           "Created sprint",
	DeletedSprint:           "Deleted sprint",
	UpdatedSprint:           "Updated sprint",
	ActivatedSprint:         "Activated sprint",
	FreezeSprint:            "Freeze the sprint",
	TriggeredSprintRefresh:  "Triggered sprint refresh",
	UpdatedSprintTask:       "Updated the task in sprint",
	MarkDoneSprintTask:      "Marked done a task in sprint",
	MarkUndoneSprintTask:    "Marked undone a task in sprint",
}
