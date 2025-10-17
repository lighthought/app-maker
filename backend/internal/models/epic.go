package models

import (
	"time"

	"gorm.io/gorm"
)

// Epic 项目史诗模型
type Epic struct {
	ID            string         `json:"id" gorm:"primaryKey;type:varchar(50);default:public.generate_table_id('EPIC', 'public.project_epics_id_num_seq')"`
	ProjectID     string         `json:"project_id" gorm:"type:varchar(50);not null"`
	ProjectGuid   string         `json:"project_guid" gorm:"type:varchar(50);not null"`
	EpicNumber    int            `json:"epic_number" gorm:"not null"`
	Name          string         `json:"name" gorm:"size:200;not null"`
	Description   string         `json:"description" gorm:"type:text"`
	Priority      string         `json:"priority" gorm:"size:20;not null"`
	EstimatedDays int            `json:"estimated_days"`
	Status        string         `json:"status" gorm:"size:20;default:'pending'"`
	FilePath      string         `json:"file_path" gorm:"size:500"`
	DisplayOrder  int            `json:"display_order" gorm:"default:0"` // 显示顺序，用于前端拖拽排序
	Stories       []Story        `json:"stories,omitempty" gorm:"foreignKey:EpicID"`
	CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Epic) TableName() string {
	return "project_epics"
}

// BeforeCreate 创建前的钩子
func (e *Epic) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		var result string
		err := tx.Raw("SELECT generate_table_id('EPIC', 'project_epics_id_num_seq')").Scan(&result).Error
		if err != nil {
			return err
		}
		e.ID = result
	}
	return nil
}

// Story 用户故事模型
type Story struct {
	ID                 string         `json:"id" gorm:"primaryKey;type:varchar(50);default:public.generate_table_id('STORY', 'public.epic_stories_id_num_seq')"`
	EpicID             string         `json:"epic_id" gorm:"type:varchar(50);not null"`
	StoryNumber        string         `json:"story_number" gorm:"size:20;not null"`
	Title              string         `json:"title" gorm:"size:200;not null"`
	Description        string         `json:"description" gorm:"type:text"`
	Priority           string         `json:"priority" gorm:"size:20;not null"`
	EstimatedDays      int            `json:"estimated_days"`
	Status             string         `json:"status" gorm:"size:20;default:'pending'"`
	FilePath           string         `json:"file_path" gorm:"size:500"`
	Depends            string         `json:"depends" gorm:"type:text"`
	Techs              string         `json:"techs" gorm:"type:text"`
	Content            string         `json:"content" gorm:"type:text"`
	AcceptanceCriteria string         `json:"acceptance_criteria" gorm:"type:text"`
	DisplayOrder       int            `json:"display_order" gorm:"default:0"` // 显示顺序，用于前端拖拽排序
	CreatedAt          time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt          gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Story) TableName() string {
	return "epic_stories"
}

// BeforeCreate 创建前的钩子
func (s *Story) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		var result string
		err := tx.Raw("SELECT generate_table_id('STORY', 'epic_stories_id_num_seq')").Scan(&result).Error
		if err != nil {
			return err
		}
		s.ID = result
	}
	return nil
}

// MvpEpicsData PO Agent 返回的 MVP Epics JSON 数据结构
type MvpEpicsData struct {
	MvpEpics []MvpEpicItem `json:"mvp_epics"`
}

type MvpEpicItem struct {
	EpicNumber    int            `json:"epic_number"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Priority      string         `json:"priority"`
	EstimatedDays int            `json:"estimated_days"`
	FilePath      string         `json:"file_path"`
	Stories       []MvpStoryItem `json:"stories"`
}

type MvpStoryItem struct {
	StoryNumber   string `json:"story_number"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Priority      string `json:"priority"`
	EstimatedDays int    `json:"estimated_days"`
	Depends       string `json:"depends"`
	Techs         string `json:"techs"`
}
