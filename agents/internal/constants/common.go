package constants

import "app-maker-agents/internal/models"

const (
	TaskStatusPending   models.TaskStatus = "pending"
	TaskStatusRunning   models.TaskStatus = "running"
	TaskStatusSuccess   models.TaskStatus = "success"
	TaskStatusFailed    models.TaskStatus = "failed"
	TaskStatusCancelled models.TaskStatus = "cancelled"
)

const (
	StageProjectBrief       models.DevStage = "project_brief"
	StageGeneratePrd        models.DevStage = "generate_prd"
	StageDefineUxStandard   models.DevStage = "define_ux_standard"
	StageGeneratePagePrompt models.DevStage = "generate_page_prompt"
	StageDesignArchitecture models.DevStage = "design_architecture"
	StageDefineApi          models.DevStage = "define_api"
	StageDefineDataModel    models.DevStage = "define_data_model"
	StagePlanEpicAndStory   models.DevStage = "plan_epic_and_story"
	StageDevelopStory       models.DevStage = "develop_story"
	StageFixBug             models.DevStage = "fix_bug"
	StageRunTest            models.DevStage = "run_test"
	StageDeploy             models.DevStage = "deploy"
)
