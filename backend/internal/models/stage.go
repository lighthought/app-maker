package models

import (
	"context"
	"time"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/tasks"

	"gorm.io/gorm"
)

// DevStage 开发阶段模型
type DevStage struct {
	ID           string         `json:"id" gorm:"primaryKey;type:varchar(50);default:public.generate_table_id('STAGE', 'public.dev_stages_id_num_seq')"`
	ProjectID    string         `json:"project_id" gorm:"type:varchar(50);not null"`
	ProjectGuid  string         `json:"project_guid" gorm:"type:varchar(50);"`
	Name         string         `json:"name" gorm:"size:100;not null"`
	Status       string         `json:"status" gorm:"size:20;not null;default:'pending'"` // pending, in_progress, completed, failed
	Progress     int            `json:"progress" gorm:"default:0"`                        // 0-100
	Description  string         `json:"description" gorm:"type:text"`
	FailedReason string         `json:"failed_reason" gorm:"type:text"`
	TaskID       string         `json:"task_id" gorm:"type:varchar(50)"`
	StartedAt    *time.Time     `json:"started_at"`
	CompletedAt  *time.Time     `json:"completed_at"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type DevStageInfo struct {
	ID           string `json:"id" gorm:"primaryKey;type:varchar(50);default:public.generate_table_id('STAGE', 'public.dev_stages_id_num_seq')"`
	ProjectID    string `json:"project_id" gorm:"type:varchar(50);not null"`
	ProjectGuid  string `json:"project_guid" gorm:"type:varchar(50);"`
	Name         string `json:"name" gorm:"size:100;not null"`
	Status       string `json:"status" gorm:"size:20;not null;default:'pending'"` // pending, in_progress, completed, failed
	Progress     int    `json:"progress" gorm:"default:0"`                        // 0-100
	Description  string `json:"description" gorm:"type:text"`
	FailedReason string `json:"failed_reason" gorm:"type:text"`
	TaskID       string `json:"task_id" gorm:"type:varchar(50)"`
}

func (ds *DevStageInfo) CopyFromDevStage(other *DevStage) {
	ds.ID = other.ID
	ds.ProjectID = other.ProjectID
	ds.ProjectGuid = other.ProjectGuid
	ds.Name = other.Name
	ds.Status = other.Status
	ds.Progress = other.Progress
	ds.Description = other.Description
	ds.FailedReason = other.FailedReason
	ds.TaskID = other.TaskID
}

func (ds *DevStageInfo) Copy(other *DevStageInfo) {
	ds.ID = other.ID
	ds.ProjectID = other.ProjectID
	ds.ProjectGuid = other.ProjectGuid
	ds.Name = other.Name
	ds.Status = other.Status
	ds.Progress = other.Progress
	ds.Description = other.Description
	ds.FailedReason = other.FailedReason
	ds.TaskID = other.TaskID
}

func (DevStage) TableName() string {
	return "dev_stages"
}

func NewDevStage(project *Project, stageName common.DevStatus, status string) *DevStage {
	if status == "" {
		status = common.CommonStatusInProgress
	}
	return &DevStage{
		ProjectID:   project.ID,
		ProjectGuid: project.GUID,
		Name:        string(stageName),
		Status:      status,
		TaskID:      project.CurrentTaskID,
		Progress:    common.GetProgressByCommonStatus(status),
		Description: common.GetDevStageDescription(stageName),
	}
}

func (ds *DevStage) SetStatus(status string) {
	ds.Status = status
	ds.Progress = common.GetProgressByCommonStatus(status)
	ds.FailedReason = ""
}

// BeforeCreate 创建前的钩子
func (ds *DevStage) BeforeCreate(tx *gorm.DB) error {
	if ds.ID == "" {
		var result string
		err := tx.Raw("SELECT generate_table_id('STAGE', 'dev_stages_id_num_seq')").Scan(&result).Error
		if err != nil {
			return err
		}
		ds.ID = result
	}
	return nil
}

// DevStageItem 开发阶段项
type DevStageItem struct {
	Name          common.DevStatus                                                              // 阶段名称
	Desc          string                                                                        // 阶段描述
	NeedConfirm   bool                                                                          // 是否需要确认
	SkipInDevMode bool                                                                          // 在开发模式下是否跳过，跳过则不执行
	ReqHandler    func(context.Context, *Project) (string, error)                               // 阶段执行器
	RespHandler   func(context.Context, *agent.AgentTaskStatusMessage, *tasks.TaskResult) error // 阶段响应处理器
}
