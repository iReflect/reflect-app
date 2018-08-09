package constants

// CustomDateFormat is a date format used in the application to parse date string/object into a usable format
const CustomDateFormat = "2006-01-02"

// constants defined to use in ActionItemTypeMap
const (
	Retrospective         = "Retrospective"
	RetrospectiveFeedback = "RetrospectiveFeedback"
	SprintMember          = "SprintMember"
	SprintMemberTask      = "SprintMemberTask"
	Sprint                = "Sprint"
	SprintTask            = "SprintTask"
)

// ActionItemTypeMap is types of ActionItem of Trail model used in adding trails.
var ActionItemTypeMap = map[string]string{
	Retrospective:         "Retrospective",
	RetrospectiveFeedback: "Retrospective Feedback",
	SprintMember:          "Sprint Member",
	SprintMemberTask:      "Sprint Member Task",
	Sprint:                "Sprint",
	SprintTask:            "Sprint Task",
}

// constants defined to use in ActionTypeMap
const (
	CreatedRetrospective    = "CreatedRetrospective"
	AddedGoal               = "AddedGoal"
	UpdatedGoal             = "UpdatedGoal"
	ResolvedGoal            = "ResolvedGoal"
	UnresolvedGoal          = "UnresolvedGoal"
	AddedHighlight          = "AddedHighlight"
	UpdatedHighlight        = "UpdatedHighlight"
	AddedSprintMember       = "AddedSprintMember"
	UpdatedSprintMember     = "UpdatedSprintMember"
	RemovedSprintMember     = "RemovedSprintMember"
	AddedNote               = "AddedNote"
	UpdatedNote             = "UpdatedNote"
	AddedSprintMemberTask   = "AddedSprintMemberTask"
	UpdatedSprintMemberTask = "UpdatedSprintMemberTask"
	CreatedSprint           = "CreatedSprint"
	DeletedSprint           = "DeletedSprint"
	UpdatedSprint           = "UpdatedSprint"
	ActivatedSprint         = "ActivatedSprint"
	FreezeSprint            = "FreezeSprint"
	TriggeredSprintRefresh  = "TriggeredSprintRefresh"
	UpdatedSprintTask       = "UpdatedSprintTask"
	MarkDoneSprintTask      = "MarkDoneSprintTask"
	MarkUndoneSprintTask    = "MarkUndoneSprintTask"
)

// ActionTypeMap is types of Action of Trail model used in adding trails.
var ActionTypeMap = map[string]string{
	CreatedRetrospective:    "Created Retrospective",
	AddedGoal:               "Added Goal",
	UpdatedGoal:             "Updated Goal",
	ResolvedGoal:            "Resolved Goal",
	UnresolvedGoal:          "Unresolved Goal",
	AddedHighlight:          "Added Highlight",
	UpdatedHighlight:        "Updated Highlight",
	AddedSprintMember:       "Added Sprint Member",
	UpdatedSprintMember:     "Updated Sprint Member",
	RemovedSprintMember:     "Removed Sprint Member",
	AddedNote:               "Added Note",
	UpdatedNote:             "Updated Note",
	AddedSprintMemberTask:   "Added Member To The Sprint on Task",
	UpdatedSprintMemberTask: "Updated Member To The Sprint on Task",
	CreatedSprint:           "Created Sprint",
	DeletedSprint:           "Deleted Sprint",
	UpdatedSprint:           "Updated Sprint",
	ActivatedSprint:         "Activated Sprint",
	FreezeSprint:            "Freeze the Sprint",
	TriggeredSprintRefresh:  "Triggered Sprint Refresh",
	UpdatedSprintTask:       "Updated The Task Of This Sprint",
	MarkDoneSprintTask:      "Mark Done A Task In This Sprint",
	MarkUndoneSprintTask:    "Mark Undone A Task In This Sprint",
}
