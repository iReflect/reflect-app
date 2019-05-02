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
	DeletedGoal             ActionType = "DeletedGoal"
	AddedHighlight          ActionType = "AddedHighlight"
	UpdatedHighlight        ActionType = "UpdatedHighlight"
	DeletedHighlight        ActionType = "DeletedHighlight"
	AddedSprintMember       ActionType = "AddedSprintMember"
	UpdatedSprintMember     ActionType = "UpdatedSprintMember"
	RemovedSprintMember     ActionType = "RemovedSprintMember"
	AddedNote               ActionType = "AddedNote"
	UpdatedNote             ActionType = "UpdatedNote"
	DeletedNote             ActionType = "DeletedNote"
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
	DeletedSprintTask       ActionType = "DeletedSprintTask"
)

// ActionTypeMap is types of Action of Trail model used in adding trails.
var ActionTypeMap = map[ActionType]string{
	CreatedRetrospective:    "Created Retrospective",
	AddedGoal:               "Added a Goal",
	UpdatedGoal:             "Updated a Goal",
	ResolvedGoal:            "Marked a goal resolved",
	UnresolvedGoal:          "Marked a goal unresolved",
	DeletedGoal:             "Deleted a goal",
	AddedHighlight:          "Added a highlight",
	UpdatedHighlight:        "Updated a highlight",
	DeletedHighlight:        "Deleted a highlight",
	AddedSprintMember:       "Added member in sprint",
	UpdatedSprintMember:     "Updated member in sprint",
	RemovedSprintMember:     "Removed member from sprint",
	AddedNote:               "Added a note",
	UpdatedNote:             "Updated a note",
	DeletedNote:             "Deleted a note",
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
	DeletedSprintTask:       "Deleted the task in sprint",
}

// ErrorResponse ...
type ErrorResponse struct {
	Message string
	Code    string
}

// APIErrorMessages ...
var APIErrorMessages = map[string]ErrorResponse{
	TaskTrackerNameIsMustError: ErrorResponse{
		Message: "No task tracker name provided in the request",
		Code:    "IR-1001",
	},
	TeamIDIsMustError: ErrorResponse{
		Message: "No team ID provided in the request",
		Code:    "IR-1002",
	},
	UserCanAccessRetroError: ErrorResponse{
		Message: "You do not have access to this retro!",
		Code:    "IR-1003",
	},
	RetrospectiveListError: ErrorResponse{
		Message: "Failed to get retrospective list",
		Code:    "IR-1004",
	},
	RetrospectiveNotFoundError: ErrorResponse{
		Message: "Retrospective not found",
		Code:    "IR-1005",
	},
	RetrospectiveDetailsError: ErrorResponse{
		Message: "Failed to get retrospective details",
		Code:    "IR-1006",
	},
	NotTeamMemberError: ErrorResponse{
		Message: "You are not a part of team!",
		Code:    "IR-1007",
	},
	MemberListNotFoundError: ErrorResponse{
		Message: "No member found for this retrospective",
		Code:    "IR-1008",
	},
	GetSprintIssueMemberSummaryError: ErrorResponse{
		Message: "Failed to get Issue Member Summary!",
		Code:    "IR-1009",
	},
	RetrospectiveNoSprintError: ErrorResponse{
		Message: "Retro has no active or frozen sprints",
		Code:    "IR-1010",
	},
	RetrospectiveLatestSprintError: ErrorResponse{
		Message: "Failed to get retrospective latest sprint",
		Code:    "IR-1011",
	},
	CreateRetrospectivePermissionError: ErrorResponse{
		Message: "You do not have permission to create the retro",
		Code:    "IR-1012",
	},
	CreateRetrospectiveError: ErrorResponse{
		Message: "Failed to create retrospective",
		Code:    "IR-1013",
	},
	InvalidRequestDataError: ErrorResponse{
		Message: "Invalid request data",
		Code:    "IR-1014",
	},
	GetUserTeamListError: ErrorResponse{
		Message: "Failed to get team list",
		Code:    "IR-1015",
	},
	RetrospectiveFeedbackAccessError: ErrorResponse{
		Message: "You do not have permission to access the retrospective feedback",
		Code:    "IR-1016",
	},
	UserCanEditSprintError: ErrorResponse{
		Message: "You do not have permission to edit the sprint",
		Code:    "IR-1017",
	},
	InvalidRetrospectiveIDError: ErrorResponse{
		Message: "Invalid retrospective id",
		Code:    "IR-1018",
	},
	SprintNotFoundError: ErrorResponse{
		Message: "Sprint not found",
		Code:    "IR-1019",
	},
	UnableToGetSprintError: ErrorResponse{
		Message: "Failed to get sprint details",
		Code:    "IR-1020",
	},
	AddRetrospectiveFeedbackHighligtError: ErrorResponse{
		Message: "Failed to add retrospective sprint highlight!",
		Code:    "IR-1021",
	},
	AddRetrospectiveFeedbackNoteError: ErrorResponse{
		Message: "Failed to add retrospective sprint note!",
		Code:    "IR-1022",
	},
	AddRetrospectiveFeedbackGoalError: ErrorResponse{
		Message: "Failed to add retrospective sprint goal!",
		Code:    "IR-1023",
	},
	RetroFeedbackNotFoundError: ErrorResponse{
		Message: "Retrospective feedback not found",
		Code:    "IR-1024",
	},
	GetRetroFeedbackError: ErrorResponse{
		Message: "Failed to get retrospective feedback details",
		Code:    "IR-1025",
	},
	UpdateResolvedGoalError: ErrorResponse{
		Message: "Can not update resolved goal",
		Code:    "IR-1026",
	},
	FeedbackExpectedAtUpdationError: ErrorResponse{
		Message: "ExpectedAt can be updated only for goal type retrospective feedback",
		Code:    "IR-1027",
	},
	RetroFeedbackGoalNotFoundError: ErrorResponse{
		Message: "Feedback goal not found",
		Code:    "IR-1028",
	},
	GetRetroFeedbackGoalError: ErrorResponse{
		Message: "Failed to get the sprint goal!",
		Code:    "IR-1029",
	},
	RetroFeedbackResolvedAtUpdationError: ErrorResponse{
		Message: "Only goal type retrospective feedback could be resolved or unresolved",
		Code:    "IR-1030",
	},
	FailedToResolveFeedbackGoalError: ErrorResponse{
		Message: "Failed to resolve the goal!",
		Code:    "IR-1031",
	},
	InvalidGoalTypeError: ErrorResponse{
		Message: "Invalid Goal type",
		Code:    "IR-1032",
	},
	UpdateRetroFeedbackHighligtError: ErrorResponse{
		Message: "Cannot update sprint highlight!",
		Code:    "IR-1023",
	},
	UpdateRetroFeedbackNoteError: ErrorResponse{
		Message: "Cannot update sprint note!",
		Code:    "IR-1034",
	},
	UpdateRetroFeedbackGoalError: ErrorResponse{
		Message: "Cannot update sprint goal!",
		Code:    "IR-1035",
	},
	DeleteRetroFeedbackGoalError: ErrorResponse{
		Message: "Failed to delete the goal!",
		Code:    "IR-1036",
	},
	DeleteRetroFeedbackHighlightError: ErrorResponse{
		Message: "Failed to delete the highlight!",
		Code:    "IR-1037",
	},
	DeleteRetroFeedbackNoteError: ErrorResponse{
		Message: "Failed to delete the note!",
		Code:    "IR-1038",
	},
	UserCanAccessSprintError: ErrorResponse{
		Message: "User doesn't have permission to access the sprint",
		Code:    "IR-1039",
	},
	GetRetroFeedbackNoteListError: ErrorResponse{
		Message: "Failed to get the sprint notes!",
		Code:    "IR-1040",
	},
	GetRetroFeedbackHighlightListError: ErrorResponse{
		Message: "Failed to get the sprint highlights!",
		Code:    "IR-1041",
	},
	FailedToUnResolveFeedbackGoalError: ErrorResponse{
		Message: "Failed to un-resolve the goal!",
		Code:    "IR-1042",
	},
	GetRetroFeedbackAddedGoalsError: ErrorResponse{
		Message: "Failed to get the sprint added goals!",
		Code:    "IR-1043",
	},
	GetRetroFeedbackCompletedGoalsError: ErrorResponse{
		Message: "Failed to get the sprint completed goals!",
		Code:    "IR-1044",
	},
	GetRetroFeedbackPendingGoalsError: ErrorResponse{
		Message: "Failed to get the sprint pending goals!",
		Code:    "IR-1045",
	},
	MemberAlreadyInSprintError: ErrorResponse{
		Message: "Member already a part of the sprint",
		Code:    "IR-1046",
	},
	NotRetroTeamMemberError: ErrorResponse{
		Message: "Member is not a part of the retrospective team",
		Code:    "IR-1047",
	},
	UnableToAddMemberError: ErrorResponse{
		Message: "Failed to add member in team",
		Code:    "IR-1048",
	},
	MemberNotInSprintError: ErrorResponse{
		Message: "Member not found in sprint",
		Code:    "IR-1049",
	},
	GetMemberSummaryError: ErrorResponse{
		Message: "Failed to get sprint member details",
		Code:    "IR-1050",
	},
	GetSprintMemberListError: ErrorResponse{
		Message: "Failed to get sprint members list",
		Code:    "IR-1051",
	},
	SprintMemberNotFoundError: ErrorResponse{
		Message: "Sprint member not found",
		Code:    "IR-1052",
	},
	RemoveSprintMemberError: ErrorResponse{
		Message: "Failed to remove sprint member",
		Code:    "IR-1053",
	},
	GetSprintMemberError: ErrorResponse{
		Message: "Failed to get sprint member",
		Code:    "IR-1054",
	},
	UpdateSprintMemberError: ErrorResponse{
		Message: "Cannot update sprint member details!",
		Code:    "IR-1055",
	},
	GetSprintListError: ErrorResponse{
		Message: "Failed to get sprints",
		Code:    "IR-1056",
	},
	GetTaskProviderConfigError: ErrorResponse{
		Message: "Failed to get task provider config. please contact admin",
		Code:    "IR-1057",
	},
	InvalidConnectionConfigError: ErrorResponse{
		Message: "Invalid connection configurations",
		Code:    "IR-1058",
	},
	SprintNotFoundInTaskTrackerError: ErrorResponse{
		Message: "Sprint not found in task tracker",
		Code:    "IR-1059",
	},
	SprintStartOrEndDateMissingError: ErrorResponse{
		Message: "Sprint doesn't have a start and/or end date, please provide the start date and end date " +
			"or set them in the task tracker",
		Code: "IR-1060",
	},
	CreateSprintError: ErrorResponse{
		Message: "Failed to create sprint",
		Code:    "IR-1061",
	},
	GetSprintSummaryError: ErrorResponse{
		Message: "Failed to get sprint summary",
		Code:    "IR-1062",
	},
	GetTaskDetailsError: ErrorResponse{
		Message: "Failed to get task info",
		Code:    "IR-1063",
	},
	UpdateSprintError: ErrorResponse{
		Message: "Sprint couldn't be updated",
		Code:    "IR-1064",
	},
	DeleteSprintError: ErrorResponse{
		Message: "This sprint can not be deleted",
		Code:    "IR-1065",
	},
	SomethingWentWrong: ErrorResponse{
		Message: "Something went wrong, please retry after some time",
		Code:    "IR-1066",
	},
	ActivateSprintError: ErrorResponse{
		Message: "Sprint couldn't be activated",
		Code:    "IR-1067",
	},
	InvalidDraftSprintAcivationError: ErrorResponse{
		Message: "Cannot activate an invalid draft sprint",
		Code:    "IR-1068",
	},
	InvalidSprintTaskListError: ErrorResponse{
		Message: "Error in fetching invalid sprint task lists",
		Code:    "IR-1069",
	},
	FrozenSprintError: ErrorResponse{
		Message: "Cannot Freeze Sprint",
		Code:    "IR-1070",
	},
	FreezeInvalidActiveSprintError: ErrorResponse{
		Message: "Cannot freeze a invalid active sprint",
		Code:    "IR-1071",
	},
	GetSprintTrailsError: ErrorResponse{
		Message: "Failed to get Activity Log!",
		Code:    "IR-1072",
	},
	UserCanAccessSprintTaskError: ErrorResponse{
		Message: "You do not have permission to access the sprint task",
		Code:    "IR-1073",
	},
	UserCanEditSprintTaskError: ErrorResponse{
		Message: "You do not have permission to edit the sprint task",
		Code:    "IR-1074",
	},
	MemberAlreadyInSprintTaskError: ErrorResponse{
		Message: "Member is already a part of the sprint task",
		Code:    "IR-1075",
	},
	InvalidTaskIDError: ErrorResponse{
		Message: "Invalid task id",
		Code:    "IR-1076",
	},
	TaskMemberNotFoundError: ErrorResponse{
		Message: "Task member not found",
		Code:    "IR-1077",
	},
	UpdateTaskMemberError: ErrorResponse{
		Message: "Failed to update task member",
		Code:    "IR-1078",
	},
	GetSprintIssuesError: ErrorResponse{
		Message: "Failed to get sprint issues",
		Code:    "IR-1079",
	},
	InvalidRetrospectiveError: ErrorResponse{
		Message: "Invalid retrospective",
		Code:    "IR-1080",
	},
	GetSprintIssueError: ErrorResponse{
		Message: "Failed to get sprint issue",
		Code:    "IR-1081",
	},
	SprintTaskNotFoundError: ErrorResponse{
		Message: "Sprint task not found",
		Code:    "IR-1082",
	},
	GetSprintTaskError: ErrorResponse{
		Message: "Failed to get sprint task",
		Code:    "IR-1083",
	},
	UpdateSprintTaskError: ErrorResponse{
		Message: "Failed to update sprint task",
		Code:    "IR-1084",
	},
	MarkDoneSprintTaskError: ErrorResponse{
		Message: "Failed to mark this task as done",
		Code:    "IR-1085",
	},
	IssueNotFoundError: ErrorResponse{
		Message: "Issue not found",
		Code:    "IR-1086",
	},
	MarkUndoneSprintTaskError: ErrorResponse{
		Message: "Failed to mark this task as undone",
		Code:    "IR-1087",
	},
	DeleteSprintTaskError: ErrorResponse{
		Message: "Failed to delete sprint task",
		Code:    "IR-1088",
	},
	TeamNotFoundError: ErrorResponse{
		Message: "Team not found",
		Code:    "IR-1089",
	},
	GetTimeProviderOptionError: ErrorResponse{
		Message: "Cannot get time provider options!",
		Code:    "IR-1090",
	},
	InvalidEmailOrPassword: ErrorResponse{
		Message: "Invalid email or password",
		Code:    "IR-1091",
	},
	InternalServerError: ErrorResponse{
		Message: "Internal server error",
		Code:    "IR-1092",
	},
	IReflectAccountNotFoundError: ErrorResponse{
		Message: "We couldn't find a iReflect account associated with ",
		Code:    "IR-1093",
	},
	GeneratedOtpNotFoundError: ErrorResponse{
		Message: "We didn't find any generated OTP for this email. Please generate a new one",
		Code:    "IR-1094",
	},
	GeneratedOtpExpiredError: ErrorResponse{
		Message: "OTP is expired. Please generate a new one",
		Code:    "IR-1095",
	},
	OtpReGeneratedError: ErrorResponse{
		Message: "You just generated a OTP. Please try again after sometime",
		Code:    "IR-1096",
	},
	InvalidOtpError: ErrorResponse{
		Message: "Invalid otp",
		Code:    "IR-1097",
	},
	UserNotFoundError: ErrorResponse{
		Message: "User not found",
		Code:    "IR-1098",
	},
	UnauthorizedError: ErrorResponse{
		Message: "Unauthorized",
		Code:    "IR-1099",
	},
	InvalidSprintError: ErrorResponse{
		Message: "Invalid sprint",
		Code:    "IR-1100",
	},
	GetRetrospectiveMembersError: ErrorResponse{
		Message: "Cannot get Retrospective Members List!",
		Code:    "IR-1101",
	},
	AdSprintIssueMemberError: ErrorResponse{
		Message: "Failed to add the member to this Issue!",
		Code:    "IR-1102",
	},
	ResyncSprintError: ErrorResponse{
		Message: "Failed to resync sprint details",
		Code:    "IR-1103",
	},
}

// constants for error messages
const (
	InvalidEmailOrPassword                = "invalidEmailOrPassword"
	TaskTrackerNameIsMustError            = "taskTrackerNameIsMustError"
	TeamIDIsMustError                     = "teamIDIsMustError"
	InvalidRequestDataError               = "invalidRequestDataError"
	UserCanAccessRetroError               = "userCanAccessRetroError"
	RetrospectiveListError                = "retrospectiveListError"
	RetrospectiveNotFoundError            = "retrospectiveNotFoundError"
	RetrospectiveDetailsError             = "retrospectiveDetailsError"
	NotTeamMemberError                    = "notTeamMemberError"
	MemberListNotFoundError               = "memberListNotFoundError"
	GetSprintIssueMemberSummaryError      = "getSprintIssueMemberSummaryError"
	RetrospectiveNoSprintError            = "retrospectiveNoSprintError"
	RetrospectiveLatestSprintError        = "retrospectiveLatestSprintError"
	CreateRetrospectivePermissionError    = "createRetrospectivePermissionError"
	CreateRetrospectiveError              = "createRetrospectiveError"
	GetUserTeamListError                  = "getUserTeamListError"
	RetrospectiveFeedbackAccessError      = "retrospectiveFeedbackAccessError"
	UserCanEditSprintError                = "userCanEditSprintError"
	InvalidRetrospectiveIDError           = "invalidRetrospectiveIDError"
	SprintNotFoundError                   = "sprintNotFoundError"
	UnableToGetSprintError                = "unableToGetSprintError"
	AddRetrospectiveFeedbackHighligtError = "addRetrospectiveFeedbackHighligtError"
	AddRetrospectiveFeedbackNoteError     = "addRetrospectiveFeedbackNoteError"
	AddRetrospectiveFeedbackGoalError     = "addRetrospectiveFeedbackGoalError"
	RetroFeedbackNotFoundError            = "retroFeedbackNotFoundError"
	GetRetroFeedbackError                 = "getRetroFeedbackError"
	UpdateResolvedGoalError               = "updateResolvedGoalError"
	FeedbackExpectedAtUpdationError       = "feedbackExpectedAtUpdationError"
	UpdateRetroFeedbackHighligtError      = "updateRetroFeedbackHighligtError"
	RetroFeedbackGoalNotFoundError        = "retroFeedbackGoalNotFoundError"
	RetroFeedbackResolvedAtUpdationError  = "retroFeedbackResolvedAtUpdationError"
	FailedToResolveFeedbackGoalError      = "failedToResolveFeedbackGoalError"
	InvalidGoalTypeError                  = "invalidGoalTypeError"
	UpdateRetroFeedbackNoteError          = "updateRetroFeedbackNoteError"
	UpdateRetroFeedbackGoalError          = "updateRetroFeedbackGoalError"
	DeleteRetroFeedbackGoalError          = "deleteRetroFeedbackGoalError"
	DeleteRetroFeedbackHighlightError     = "deleteRetroFeedbackHighlightError"
	DeleteRetroFeedbackNoteError          = "deleteRetroFeedbackNoteError"
	UserCanAccessSprintError              = "userCanAccessSprintError"
	GetRetroFeedbackNoteListError         = "getRetroFeedbackNoteListError"
	GetRetroFeedbackHighlightListError    = "getRetroFeedbackHighlightListError"
	FailedToUnResolveFeedbackGoalError    = "failedToUnResolveFeedbackGoalError"
	GetRetroFeedbackGoalError             = "getRetroFeedbackGoalError"
	GetRetroFeedbackAddedGoalsError       = "getRetroFeedbackAddedGoalsError"
	GetRetroFeedbackCompletedGoalsError   = "getRetroFeedbackCompletedGoalsError"
	GetRetroFeedbackPendingGoalsError     = "getRetroFeedbackPendingGoalsError"
	MemberAlreadyInSprintError            = "memberAlreadyInSprintError"
	NotRetroTeamMemberError               = "notRetroTeamMemberError"
	UnableToAddMemberError                = "unableToAddMemberError"
	MemberNotInSprintError                = "memberNotInSprintError"
	GetMemberSummaryError                 = "getMemberSummaryError"
	GetSprintMemberListError              = "getSprintMemberListError"
	SprintMemberNotFoundError             = "sprintMemberNotFoundError"
	RemoveSprintMemberError               = "removeSprintMemberError"
	GetSprintMemberError                  = "getSprintMemberError"
	UpdateSprintMemberError               = "updateSprintMemberError"
	GetSprintListError                    = "getSprintListError"
	GetTaskProviderConfigError            = "getTaskProviderConfigError"
	InvalidConnectionConfigError          = "invalidConnectionConfigError"
	SprintNotFoundInTaskTrackerError      = "sprintNotFoundInTaskTrackerError"
	SprintStartOrEndDateMissingError      = "sprintStartOrEndDateMissingError"
	CreateSprintError                     = "createSprintError"
	GetSprintSummaryError                 = "getSprintSummaryError"
	GetTaskDetailsError                   = "getTaskDetailsError"
	UpdateSprintError                     = "updateSprintError"
	DeleteSprintError                     = "deleteSprintError"
	SomethingWentWrong                    = "somethingWentWrong"
	ActivateSprintError                   = "activateSprintError"
	InvalidDraftSprintAcivationError      = "invalidDraftSprintAcivationError"
	InvalidSprintTaskListError            = "invalidSprintTaskListError"
	FrozenSprintError                     = "frozenSprintError"
	FreezeInvalidActiveSprintError        = "freezeInvalidActiveSprintError"
	GetSprintTrailsError                  = "getSprintTrailsError"
	UserCanAccessSprintTaskError          = "userCanAccessSprintTaskError"
	UserCanEditSprintTaskError            = "userCanEditSprintTaskError"
	MemberAlreadyInSprintTaskError        = "memberAlreadyInSprintTaskError"
	InvalidTaskIDError                    = "invalidTaskIDError"
	TaskMemberNotFoundError               = "taskMemberNotFoundError"
	UpdateTaskMemberError                 = "updateTaskMemberError"
	GetSprintIssuesError                  = "getSprintIssuesError"
	InvalidRetrospectiveError             = "invalidRetrospectiveError"
	GetSprintIssueError                   = "getSprintIssueError"
	SprintTaskNotFoundError               = "sprintTaskNotFoundError"
	GetSprintTaskError                    = "getSprintTaskError"
	UpdateSprintTaskError                 = "updateSprintTaskError"
	MarkDoneSprintTaskError               = "markDoneSprintTaskError"
	IssueNotFoundError                    = "issueNotFoundError"
	MarkUndoneSprintTaskError             = "markUndoneSprintTaskError"
	DeleteSprintTaskError                 = "deleteSprintTaskError"
	TeamNotFoundError                     = "teamNotFoundError"
	GetTimeProviderOptionError            = "getTimeProviderOptionError"
	InternalServerError                   = "internalServerError"
	IReflectAccountNotFoundError          = "iReflectAccountNotFoundError"
	GeneratedOtpNotFoundError             = "generatedOtpNotFoundError"
	GeneratedOtpExpiredError              = "generatedOtpExpiredError"
	OtpReGeneratedError                   = "otpReGeneratedError"
	InvalidOtpError                       = "invalidOtpError"
	UserNotFoundError                     = "userNotFoundError"
	UnauthorizedError                     = "unauthorizedError"
	InvalidSprintError                    = "invalidSprintError"
	GetRetrospectiveMembersError          = "getRetrospectiveMembersError"
	AdSprintIssueMemberError              = "addSprintIssueMemberError"
	ResyncSprintError                     = "resyncSprintError"
)

// <----------- constants for email --------------->

// OTPEmailSubject ...
const OTPEmailSubject = "One Time Password"

// EmailMIME ...
const EmailMIME = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

// <------------- time tracker constants ------------>

// GenericTimeTrackersList is list of generic time providers which can be used for any task provider.
var GenericTimeTrackersList = []string{"gsheet"}

// <-------------------- end ------------------------->
