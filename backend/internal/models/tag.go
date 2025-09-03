package models

import (
	"time"

	"gorm.io/gorm"
)

// Tag 标签模型
type Tag struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(50);default:public.generate_table_id('TAGS', 'public.tags_id_num_seq')"`
	Name      string         `json:"name" gorm:"uniqueIndex;not null"`
	Color     string         `json:"color" gorm:"default:'#666666'"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Projects []Project `json:"projects,omitempty" gorm:"many2many:project_tags;"`
}

// TableName 指定表名
func (Tag) TableName() string {
	return "tags"
}

// BeforeCreate 创建前的钩子
func (t *Tag) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		var result string
		err := tx.Raw("SELECT generate_table_id('TAGS', 'tags_id_num_seq')").Scan(&result).Error
		if err != nil {
			return err
		}
		t.ID = result
	}
	return nil
}

// ProjectTag 项目标签关联表
type ProjectTag struct {
	ProjectID string    `json:"project_id" gorm:"primaryKey;type:varchar(50)"`
	TagID     string    `json:"tag_id" gorm:"primaryKey;type:varchar(50)"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (ProjectTag) TableName() string {
	return "project_tags"
}
