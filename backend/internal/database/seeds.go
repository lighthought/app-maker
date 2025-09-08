package database

import (
	"log"

	"autocodeweb-backend/internal/models"

	"gorm.io/gorm"
)

// Seed 执行数据库种子数据
func Seed(db *gorm.DB) error {
	log.Println("开始执行数据库种子数据...")

	// 创建默认管理员用户
	if err := createDefaultAdmin(db); err != nil {
		return err
	}

	log.Println("数据库种子数据执行完成")
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
			Username: "系统管理员",
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
