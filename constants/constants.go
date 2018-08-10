package constants

// CustomDateFormat is a date format used in the application to parse date string/object into a usable format
const CustomDateFormat = "2006-01-02"

// ActionItemDataType is datatype for ActionItems's
type ActionItemDataType string

// constants defined to use in ActionItemTypeMap
const (
	Retrospective         ActionItemDataType = "Retrospective"
	RetrospectiveFeedback ActionItemDataType = "RetrospectiveFeedback"
	SprintMember          ActionItemDataType = "SprintMember"
	SprintMemberTask      ActionItemDataType = "SprintMemberTask"
	Sprint                ActionItemDataType = "Sprint"
	SprintTask            ActionItemDataType = "SprintTask"
)

// ActionItemTypeMap is types of ActionItem of Trail model used in adding trails.
var ActionItemTypeMap = map[ActionItemDataType]string{
	Retrospective:         "Retrospective",
	RetrospectiveFeedback: "Retrospective Feedback",
	SprintMember:          "Sprint Member",
	SprintMemberTask:      "Sprint Member Task",
	Sprint:                "Sprint",
	SprintTask:            "Sprint Task",
}

// ActionDataType special data type for action's
type ActionDataType string

// constants defined to use in ActionTypeMap
const (
	CreatedRetrospective    ActionDataType = "CreatedRetrospective"
	AddedGoal               ActionDataType = "AddedGoal"
	UpdatedGoal             ActionDataType = "UpdatedGoal"
	ResolvedGoal            ActionDataType = "ResolvedGoal"
	UnresolvedGoal          ActionDataType = "UnresolvedGoal"
	AddedHighlight          ActionDataType = "AddedHighlight"
	UpdatedHighlight        ActionDataType = "UpdatedHighlight"
	AddedSprintMember       ActionDataType = "AddedSprintMember"
	UpdatedSprintMember     ActionDataType = "UpdatedSprintMember"
	RemovedSprintMember     ActionDataType = "RemovedSprintMember"
	AddedNote               ActionDataType = "AddedNote"
	UpdatedNote             ActionDataType = "UpdatedNote"
	AddedSprintMemberTask   ActionDataType = "AddedSprintMemberTask"
	UpdatedSprintMemberTask ActionDataType = "UpdatedSprintMemberTask"
	CreatedSprint           ActionDataType = "CreatedSprint"
	DeletedSprint           ActionDataType = "DeletedSprint"
	UpdatedSprint           ActionDataType = "UpdatedSprint"
	ActivatedSprint         ActionDataType = "ActivatedSprint"
	FreezeSprint            ActionDataType = "FreezeSprint"
	TriggeredSprintRefresh  ActionDataType = "TriggeredSprintRefresh"
	UpdatedSprintTask       ActionDataType = "UpdatedSprintTask"
	MarkDoneSprintTask      ActionDataType = "MarkDoneSprintTask"
	MarkUndoneSprintTask    ActionDataType = "MarkUndoneSprintTask"
)

// ActionTypeMap is types of Action of Trail model used in adding trails.
var ActionTypeMap = map[ActionDataType]string{
	CreatedRetrospective:    "Created Retrospective",
	AddedGoal:               "Added a Goal",
	UpdatedGoal:             "Updated a Goal",
	ResolvedGoal:            "Mark a goal resolved",
	UnresolvedGoal:          "Mark a goal unresolved",
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
	MarkDoneSprintTask:      "Mark done a task in sprint",
	MarkUndoneSprintTask:    "Mark undone a task in sprint",
}
