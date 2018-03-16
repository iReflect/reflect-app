package constants

const (
	InternalServerError = "internal server error"
	UserNotFound = "could not find user"

	InvalidUserState   = "invalid user state"
	InvalidRequestData = "invalid request data"

	UnAuthorizedUser     = "unauthorized user"
	NoUserWithEmailFound = "user with email %s not found"

	FeedbackNotFound        = "no feedback(s) found"
	FeedbackUpdateError     = "feedback couldn't be updated"
	InvalidFeedbackTemplate = "no feedback form content found"

	QuestionNotFound            = "no question(s) found"
	QuestionResponseUpdateError = "question response couldn't be updated"

	NoUserTeamFound    = "no team(s) found for the user"
	UserNotATeamMember = "must be a member of this team"
	NoTeamMemberFound  = "team has no members"

	RetrospectiveNotFound       = "no retrospective(s) found"
	RetrospectiveCreateError    = "retrospective couldn't be created"
	RetroCreatePermissionDenied = "user doesn't have the permission to create the retro"

	InvalidProviderConfigError   = "couldn't validate task provider configuration, please ensure that you have provided the correct information"
	TaskProviderConfigParseError = "couldn't parse task provider configuration, please ensure that you have provided the correct information"

	TimeProviderConfigParseError = "couldn't parse time provider configuration, please ensure that your account has the correct information"

	InvalidTaskID   = "invalid task id"
	TaskUpdateError = "unable to update task"
	TaskNotFound    = "no task(s) found"

	InvalidSprintID     = "invalid sprint id"
	SprintNotFound      = "no sprint(s) found"
	SprintCreateError   = "sprint couldn't be created"
	SprintDeleteError   = "sprint couldn't be deleted"
	SprintActivateError = "sprint couldn't be activated"
	SprintFreezeError   = "sprint couldn't be frozen"
	SprintUpdateError   = "sprint couldn't be updated"
	SprintSyncError     = "sprint couldn't be synced"
	SprintTaskListError = "couldn't get sprint tasks list"
	SprintStartEndDateError = "sprint doesn't have any start and/or end date"
	SprintNotFoundInTaskTracker = "couldn't find the sprint in the task tracker"

	AlreadySprintMemberError = "member already a part of the sprint"
	NotASprintMemberError    = "member is not a part of the sprint"
	SprintMemberCreateError  = "sprint member couldn't be created"
	SprintMemberNotFound     = "no sprint member(s) found"
	SprintMemberDeleteError  = "sprint member(s) couldn't be deleted"
	SprintMemberUpdateError  = "sprint member(s) couldn't be updated"
	SprintMemberAddError     = "sprint member(s) couldn't be added"
	SprintMemberSummaryError = "couldn't get sprint member summary"

	SprintTaskMemberNotFound     = "no sprint task member(s) found"
	SprintMemberTaskNotFound     = "no sprint member task(s) found"
	SprintMemberTaskAddError     = "sprint task member couldn't be added"
	SprintMemberTaskUpdateError  = "sprint member task(s) couldn't be updated"
	SprintMemberTaskDeleteError  = "sprint member task(s) couldn't be deleted"
	AlreadySprintMemberTaskError = "member already a part of the sprint task"
)
