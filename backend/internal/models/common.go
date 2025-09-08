package models

import (
	"time"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(50);default:public.generate_table_id('BASE', 'public.base_id_num_seq')"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at,omitempty"`
}
