package models

import (
	"autocodeweb-backend/internal/constants"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Project 项目模型
type Project struct {
	ID               string         `json:"id" gorm:"primaryKey;type:varchar(50);default:public.generate_table_id('PROJ', 'public.projects_id_num_seq')"`
	GUID             string         `json:"guid" gorm:"size:50;not null"`
	Name             string         `json:"name" gorm:"size:100;not null"`
	Description      string         `json:"description" gorm:"type:text"`
	Requirements     string         `json:"requirements" gorm:"type:text;not null"`
	Status           string         `json:"status" gorm:"size:20;not null;default:'draft'"` // draft, in_progress, done, failed
	DevStatus        string         `json:"dev_status" gorm:"size:50;default:'pending'"`    // 开发子状态
	DevProgress      int            `json:"dev_progress" gorm:"default:0"`                  // 开发进度 0-100
	CurrentTaskID    string         `json:"current_task_id" gorm:"type:varchar(50)"`        // 当前执行的任务ID
	BackendPort      int            `json:"backend_port" gorm:"not null;default:9501"`
	FrontendPort     int            `json:"frontend_port" gorm:"not null;default:3501"`
	RedisPort        int            `json:"redis_port" gorm:"not null;default:7501"`
	PostgresPort     int            `json:"postgres_port" gorm:"not null;default:5501"`
	ApiBaseUrl       string         `json:"api_base_url" gorm:"size:200"`
	AppSecretKey     string         `json:"app_secret_key" gorm:"size:100"`
	DatabasePassword string         `json:"database_password" gorm:"size:100"`
	RedisPassword    string         `json:"redis_password" gorm:"size:100"`
	JwtSecretKey     string         `json:"jwt_secret_key" gorm:"size:100"`
	Subnetwork       string         `json:"subnetwork" gorm:"size:50"`
	PreviewUrl       string         `json:"preview_url" gorm:"size:500"`
	ProjectPath      string         `json:"project_path" gorm:"size:500;not null"`
	UserID           string         `json:"user_id" gorm:"type:varchar(50);not null"`
	GitlabRepoURL    string         `json:"gitlab_repo_url" gorm:"size:500"`
	User             User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CreatedAt        time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

func (p *Project) SetDevStatus(status string) {
	p.DevStatus = status
	p.DevProgress = constants.GetDevStageProgress(status)
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
		p.GUID = uuid.New().String()
		p.GUID = strings.ReplaceAll(p.GUID, "-", "")
	}
	return nil
}
