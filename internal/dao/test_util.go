package dao

import (
	"gorm.io/gorm"
)

// SetDB 用于测试时注入模拟的 DB 实例
func SetDB(mockDB *gorm.DB) {
	db = mockDB
}
