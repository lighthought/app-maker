package models

type Agent struct {
	Name        string `json:"name" gorm:"size:100;not null"`
	Role        string `json:"role" gorm:"size:20;not null"`
	ChineseRole string `json:"chinese_role" gorm:"size:100;not null"`
}
