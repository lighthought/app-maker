package models

import (
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/utils"

	"gorm.io/gorm"
)

type ProjectInfoUpdate struct {
	ID          string `json:"id" gorm:"primaryKey;type:varchar(50);default:public.generate_table_id('PROJ', 'public.projects_id_num_seq')"`
	GUID        string `json:"guid" gorm:"size:50;not null"`
	Name        string `json:"name" gorm:"size:100;not null"`
	Status      string `json:"status" gorm:"size:20;not null;default:'pending'"` // pending, in_progress, done, failed
	Description string `json:"description" gorm:"type:text"`
	PreviewUrl  string `json:"preview_url" gorm:"size:500"`
}

func (p *ProjectInfoUpdate) Copy(other *ProjectInfoUpdate) {
	p.ID = other.ID
	p.GUID = other.GUID
	p.Name = other.Name
	p.Status = other.Status
	p.Description = other.Description
	p.PreviewUrl = other.PreviewUrl
}

// Project 项目模型
type Project struct {
	ID                    string         `json:"id" gorm:"primaryKey;type:varchar(50);default:public.generate_table_id('PROJ', 'public.projects_id_num_seq')"`
	GUID                  string         `json:"guid" gorm:"size:50;not null"`
	Name                  string         `json:"name" gorm:"size:100;not null"`
	Description           string         `json:"description" gorm:"type:text"`
	Requirements          string         `json:"requirements" gorm:"type:text;not null"`
	Status                string         `json:"status" gorm:"size:20;not null;default:'pending'"` // pending, in_progress, done, failed
	DevStatus             string         `json:"dev_status" gorm:"size:50;default:'pending'"`      // 开发子状态
	DevProgress           int            `json:"dev_progress" gorm:"default:0"`                    // 开发进度 0-100
	CurrentTaskID         string         `json:"current_task_id" gorm:"type:varchar(50)"`          // 当前执行的任务ID
	BackendPort           int            `json:"backend_port" gorm:"not null;default:9501"`
	FrontendPort          int            `json:"frontend_port" gorm:"not null;default:3501"`
	RedisPort             int            `json:"redis_port" gorm:"not null;default:7501"`
	PostgresPort          int            `json:"postgres_port" gorm:"not null;default:5501"`
	ApiBaseUrl            string         `json:"api_base_url" gorm:"size:200"`
	AppSecretKey          string         `json:"app_secret_key" gorm:"size:100"`
	DatabasePassword      string         `json:"database_password" gorm:"size:100"`
	RedisPassword         string         `json:"redis_password" gorm:"size:100"`
	JwtSecretKey          string         `json:"jwt_secret_key" gorm:"size:100"`
	Subnetwork            string         `json:"subnetwork" gorm:"size:50"`
	PreviewUrl            string         `json:"preview_url" gorm:"size:500"`
	ProjectPath           string         `json:"project_path" gorm:"size:500;not null"`
	UserID                string         `json:"user_id" gorm:"type:varchar(50);not null"`
	GitlabRepoURL         string         `json:"gitlab_repo_url" gorm:"size:500"`
	CliTool               string         `json:"cli_tool" gorm:"size:50"`
	AiModel               string         `json:"ai_model" gorm:"size:100"`
	ModelProvider         string         `json:"model_provider" gorm:"size:50"`
	ModelApiUrl           string         `json:"model_api_url" gorm:"size:500"`
	ApiToken              string         `json:"api_token,omitempty" gorm:"size:500"`           // API Token，敏感信息
	WaitingForUserConfirm bool           `json:"waiting_for_user_confirm" gorm:"default:false"` // 是否等待用户确认
	ConfirmStage          string         `json:"confirm_stage" gorm:"size:50"`                  // 等待确认的阶段
	AutoGoNext            bool           `json:"auto_go_next" gorm:"default:false"`             // 项目级自动进入下一阶段配置
	User                  User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CreatedAt             time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt             gorm.DeletedAt `json:"-" gorm:"index"`
}

func GetDefaultProject(userID, requirements string) *Project {
	guid := utils.GenerateUUID()
	filePath := filepath.Join("/app/data/projects", userID, guid)
	newProject := &Project{
		GUID:         guid,
		Requirements: requirements,
		UserID:       userID,
		Status:       common.CommonStatusPending,
		ProjectPath:  filePath,
		BackendPort:  9501,
		FrontendPort: 3501,
		RedisPort:    7501,
		PostgresPort: 5501,
	}
	return newProject
}

func (p *Project) GetUpdateInfo() *ProjectInfoUpdate {
	return &ProjectInfoUpdate{
		ID:          p.ID,
		GUID:        p.GUID,
		Name:        p.Name,
		Status:      p.Status,
		Description: p.Description,
		PreviewUrl:  p.PreviewUrl,
	}
}

func (p *Project) SetDevStatus(stage common.DevStatus) {
	p.DevStatus = string(stage)
	p.DevProgress = common.GetDevStageProgress(stage)
}

// 转换为 []byte
func (p *Project) ToBytes() ([]byte, error) {
	return json.Marshal(p)
}

// TableName 指定表名
func (Project) TableName() string {
	return "projects"
}

// BeforeCreate 创建前的钩子
func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.Status == "" {
		p.Status = "draft"
	}
	if p.BackendPort == 0 {
		p.BackendPort = 9501
	}
	if p.FrontendPort == 0 {
		p.FrontendPort = 3501
	}
	if p.RedisPort == 0 {
		p.RedisPort = 7501
	}
	if p.PostgresPort == 0 {
		p.PostgresPort = 5501
	}
	if p.ID == "" {
		var result string
		err := tx.Raw("SELECT generate_table_id('PROJ', 'projects_id_num_seq')").Scan(&result).Error
		if err != nil {
			return err
		}
		p.ID = result
	}
	if p.GUID == "" {
		p.GUID = utils.GenerateUUID()
	}
	return nil
}

type ProjectShareInfo struct {
	Token     string    `json:"token" gorm:"size:100;uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	ShareLink string    `json:"share_link" gorm:"size:500"`
}
