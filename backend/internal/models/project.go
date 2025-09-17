package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Project 项目模型
type Project struct {
	ID               string         `json:"id" gorm:"primaryKey;type:varchar(50);default:public.generate_table_id('PROJ', 'public.projects_id_num_seq')"`
	Name             string         `json:"name" gorm:"size:100;not null"`
	Description      string         `json:"description" gorm:"type:text"`
	Requirements     string         `json:"requirements" gorm:"type:text;not null"`
	Status           string         `json:"status" gorm:"size:20;not null;default:'draft'"` // draft, in_progress, completed, failed
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

// 开发子状态常量
const (
	DevStatusPending               = "pending"                // 等待开始
	DevStatusEnvironmentProcessing = "environment_processing" // 环境处理中
	DevStatusEnvironmentDone       = "environment_done"       // 环境就绪
	DevStatusPRDGenerating         = "prd_generating"         // PRD生成中
	DevStatusPRDCompleted          = "prd_completed"          // PRD完成
	DevStatusUXDefining            = "ux_defining"            // UX标准定义中
	DevStatusUXCompleted           = "ux_completed"           // UX标准完成
	DevStatusArchDesigning         = "arch_designing"         // 架构设计中
	DevStatusArchCompleted         = "arch_completed"         // 架构设计完成
	DevStatusDataModeling          = "data_modeling"          // 数据模型定义中
	DevStatusDataCompleted         = "data_completed"         // 数据模型完成
	DevStatusAPIDefining           = "api_defining"           // API接口定义中
	DevStatusAPICompleted          = "api_completed"          // API接口完成
	DevStatusEpicPlanning          = "epic_planning"          // Epic和Story划分中
	DevStatusEpicCompleted         = "epic_completed"         // Epic和Story完成
	DevStatusStoryDeveloping       = "story_developing"       // Story开发中
	DevStatusStoryCompleted        = "story_completed"        // Story开发完成
	DevStatusBugFixing             = "bug_fixing"             // 问题修复中
	DevStatusBugFixed              = "bug_fixed"              // 问题修复完成
	DevStatusTesting               = "testing"                // 自动测试中
	DevStatusTestCompleted         = "test_completed"         // 自动测试完成
	DevStatusPackaging             = "packaging"              // 打包中
	DevStatusPackaged              = "packaged"               // 打包完成
	DevStatusCompleted             = "completed"              // 开发完成
	DevStatusFailed                = "failed"                 // 开发失败
)

// 获取开发阶段进度
func (p *Project) GetDevStageProgress() int {
	switch p.DevStatus {
	case DevStatusPending:
		return 0
	case DevStatusEnvironmentProcessing:
		return 2
	case DevStatusEnvironmentDone:
		return 5
	case DevStatusPRDGenerating:
		return 8
	case DevStatusPRDCompleted:
		return 12
	case DevStatusUXDefining:
		return 16
	case DevStatusUXCompleted:
		return 20
	case DevStatusArchDesigning:
		return 24
	case DevStatusArchCompleted:
		return 28
	case DevStatusDataModeling:
		return 32
	case DevStatusDataCompleted:
		return 36
	case DevStatusAPIDefining:
		return 40
	case DevStatusAPICompleted:
		return 44
	case DevStatusEpicPlanning:
		return 48
	case DevStatusEpicCompleted:
		return 52
	case DevStatusStoryDeveloping:
		return 56
	case DevStatusStoryCompleted:
		return 60
	case DevStatusBugFixing:
		return 64
	case DevStatusBugFixed:
		return 68
	case DevStatusTesting:
		return 72
	case DevStatusTestCompleted:
		return 76
	case DevStatusPackaging:
		return 80
	case DevStatusPackaged:
		return 85
	case DevStatusCompleted:
		return 100
	case DevStatusFailed:
		return 0
	default:
		return 0
	}
}

// 获取开发阶段描述
func (p *Project) GetDevStageDescription() string {
	switch p.DevStatus {
	case DevStatusPending:
		return "等待开始开发"
	case DevStatusEnvironmentProcessing:
		return "正在初始化开发环境"
	case DevStatusEnvironmentDone:
		return "开发环境准备就绪"
	case DevStatusPRDGenerating:
		return "正在生成产品需求文档"
	case DevStatusPRDCompleted:
		return "产品需求文档已完成"
	case DevStatusUXDefining:
		return "正在定义用户体验标准"
	case DevStatusUXCompleted:
		return "用户体验标准已定义"
	case DevStatusArchDesigning:
		return "正在进行系统架构设计"
	case DevStatusArchCompleted:
		return "系统架构设计已完成"
	case DevStatusDataModeling:
		return "正在定义数据模型"
	case DevStatusDataCompleted:
		return "数据模型已定义"
	case DevStatusAPIDefining:
		return "正在定义API接口"
	case DevStatusAPICompleted:
		return "API接口已定义"
	case DevStatusEpicPlanning:
		return "正在划分Epic和Story"
	case DevStatusEpicCompleted:
		return "Epic和Story划分完成"
	case DevStatusStoryDeveloping:
		return "正在开发Story功能"
	case DevStatusStoryCompleted:
		return "Story功能开发完成"
	case DevStatusBugFixing:
		return "正在修复开发问题"
	case DevStatusBugFixed:
		return "开发问题修复完成"
	case DevStatusTesting:
		return "正在进行自动测试"
	case DevStatusTestCompleted:
		return "自动测试完成"
	case DevStatusPackaging:
		return "正在打包项目"
	case DevStatusPackaged:
		return "项目打包完成"
	case DevStatusCompleted:
		return "项目开发完成"
	case DevStatusFailed:
		return "项目开发失败"
	default:
		return "未知状态"
	}
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
	return nil
}
