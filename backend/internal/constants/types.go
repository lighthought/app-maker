package constants

const (
	WebSocketMessageTypePing                 = "ping"
	WebSocketMessageTypePong                 = "pong"
	WebSocketMessageTypeJoinProject          = "join_project"
	WebSocketMessageTypeLeaveProject         = "leave_project"
	WebSocketMessageTypeProjectStageUpdate   = "project_stage_update"
	WebSocketMessageTypeProjectMessage       = "project_message"
	WebSocketMessageTypeProjectInfoUpdate    = "project_info_update"
	WebSocketMessageTypeAgentMessage         = "agent_message"
	WebSocketMessageTypeUserFeedback         = "user_feedback"
	WebSocketMessageTypeUserFeedbackResponse = "user_feedback_response"
	WebSocketMessageTypeError                = "error"
)
