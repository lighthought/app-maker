package database

import (
	"log"

	"autocodeweb-backend/internal/models"

	"gorm.io/gorm"
)

// Seed 执行数据库种子数据
func Seed(db *gorm.DB) error {
	log.Println("开始执行数据库种子数据...")

	// 创建默认标签
	if err := createDefaultTags(db); err != nil {
		return err
	}

	// 创建默认管理员用户
	if err := createDefaultAdmin(db); err != nil {
		return err
	}

	log.Println("数据库种子数据执行完成")
	return nil
}

// createDefaultTags 创建默认标签
func createDefaultTags(db *gorm.DB) error {
	defaultTags := []models.Tag{
		{Name: "Web应用", Color: "#3B82F6"},
		{Name: "移动应用", Color: "#10B981"},
		{Name: "桌面应用", Color: "#F59E0B"},
		{Name: "API服务", Color: "#8B5CF6"},
		{Name: "数据库", Color: "#EF4444"},
		{Name: "机器学习", Color: "#06B6D4"},
		{Name: "区块链", Color: "#84CC16"},
		{Name: "游戏开发", Color: "#F97316"},
	}

	for _, tag := range defaultTags {
		var existingTag models.Tag
		if err := db.Where("name = ?", tag.Name).First(&existingTag).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&tag).Error; err != nil {
					return err
				}
				log.Printf("创建标签: %s", tag.Name)
			} else {
				return err
			}
		}
	}

	return nil
}

// createDefaultAdmin 创建默认管理员用户
func createDefaultAdmin(db *gorm.DB) error {
	// 检查是否已存在管理员用户
	var adminCount int64
	if err := db.Model(&models.User{}).Where("role = ?", "admin").Count(&adminCount).Error; err != nil {
		return err
	}

	if adminCount == 0 {
		// 创建默认管理员用户
		adminUser := models.User{
			Email:    "admin@autocodeweb.com",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
			Name:     "系统管理员",
			Role:     "admin",
			Status:   "active",
		}

		if err := db.Create(&adminUser).Error; err != nil {
			return err
		}

		log.Println("创建默认管理员用户: admin@autocodeweb.com (密码: password)")
	}

	return nil
}
