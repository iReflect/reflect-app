package constants

// CustomDateFormat is a date format used in the application to parse date string/object into a usable format
const CustomDateFormat = "2006-01-02"

// ActionItemType is datatype for ActionItems's
type ActionItemType string

// constants defined to use in ActionItemTypeMap
const (
	Retrospective         ActionItemType = "Retrospective"
	RetrospectiveFeedback                = "RetrospectiveFeedback"
	SprintMember                         = "SprintMember"
	SprintMemberTask                     = "SprintMemberTask"
	Sprint                               = "Sprint"
	SprintTask                           = "SprintTask"
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
	AddedGoal                          = "AddedGoal"
	UpdatedGoal                        = "UpdatedGoal"
	ResolvedGoal                       = "ResolvedGoal"
	UnresolvedGoal                     = "UnresolvedGoal"
	AddedHighlight                     = "AddedHighlight"
	UpdatedHighlight                   = "UpdatedHighlight"
	AddedSprintMember                  = "AddedSprintMember"
	UpdatedSprintMember                = "UpdatedSprintMember"
	RemovedSprintMember                = "RemovedSprintMember"
	AddedNote                          = "AddedNote"
	UpdatedNote                        = "UpdatedNote"
	AddedSprintMemberTask              = "AddedSprintMemberTask"
	UpdatedSprintMemberTask            = "UpdatedSprintMemberTask"
	CreatedSprint                      = "CreatedSprint"
	DeletedSprint                      = "DeletedSprint"
	UpdatedSprint                      = "UpdatedSprint"
	ActivatedSprint                    = "ActivatedSprint"
	FreezeSprint                       = "FreezeSprint"
	TriggeredSprintRefresh             = "TriggeredSprintRefresh"
	UpdatedSprintTask                  = "UpdatedSprintTask"
	MarkDoneSprintTask                 = "MarkDoneSprintTask"
	MarkUndoneSprintTask               = "MarkUndoneSprintTask"
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
	UpdatedSprintMember:     "Updated member in sprint",
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

// constants for error messages
const (
	InvalidEmailOrPassword = "Invalid email or password"
)

// <----------- constants for email --------------->

// OTPEmailSubject ...
const OTPEmailSubject = "One Time Password"

// EmailMIME ...
const EmailMIME = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

// GenricTimeTrackersList ...
var GenricTimeTrackersList = []string{"gsheet"}

// TaskTrackerNameIsMustError ...
const TaskTrackerNameIsMustError = "no task tracker name provided in the request"

// TeamIDIsMustError ...
const TeamIDIsMustError = "no team ID provided in the request"
