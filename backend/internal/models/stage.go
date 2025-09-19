package models

import (
	"autocodeweb-backend/internal/constants"
	"time"

	"gorm.io/gorm"
)

// DevStage 开发阶段模型
type DevStage struct {
	ID        string `json:"id" gorm:"primaryKey;type:varchar(50);default:public.generate_table_id('STAGE', 'public.dev_stages_id_num_seq')"`
	ProjectID string `json:"project_id" gorm:"type:varchar(50);not null"`
	//Project     Project        `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Name        string         `json:"name" gorm:"size:100;not null"`
	Status      string         `json:"status" gorm:"size:20;not null;default:'pending'"` // pending, in_progress, completed, failed
	Progress    int            `json:"progress" gorm:"default:0"`                        // 0-100
	Description string         `json:"description" gorm:"type:text"`
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (DevStage) TableName() string {
	return "dev_stages"
}

func NewDevStage(projectId, name, status string) *DevStage {
	return &DevStage{
		ProjectID:   projectId,
		Name:        name,
		Status:      status,
		Progress:    constants.GetProgressByCommandStatus(status),
		Description: constants.GetDevStageDescription(name),
	}
}

func (ds *DevStage) SetStatus(status string) {
	ds.Status = status
	ds.Progress = constants.GetProgressByCommandStatus(status)
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
