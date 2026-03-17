package model

import (
	"GoLinko/pkg/zlog"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserInfo struct {
	*gorm.Model
	Uuid      string `gorm:"column:uuid;uniqueIndex;type:char(255);comment:用户唯一id"`
	Nickname  string `gorm:"column:nickname;type:varchar(20);not null;comment:昵称"`
	Telephone string `gorm:"column:telephone;not null;type:char(11);comment:电话"`
	Email     string `gorm:"column:email;index;not null;type:varchar(30);comment:邮箱"`
	Avatar    string `gorm:"column:avatar;type:varchar(255);default:https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png;not null;comment:头像"`
	Gender    int8   `gorm:"column:gender;default:0;comment:性别，0.男，1.女"`
	Signature string `gorm:"column:signature;type:varchar(100);comment:个性签名"`
	Password  string `gorm:"column:password;type:char(64);not null;comment:密码"` // 使用 char(64) 存储加密后的密码
	Birthday  string `gorm:"column:birthday;type:char(20);comment:生日"`
	IsAdmin   int8   `gorm:"column:is_admin;not null;default:0;comment:是否是管理员，0.不是，1.是"`
	Status    int8   `gorm:"column:status;not null;default:0;comment:状态，0.正常，1.禁用"`
}

func (UserInfo) TableName() string {
	return "user_info"
}

// GeneratePassword 生成加密后的密码
func GeneratePassword(pwdDigest string) string {
	pwd, err := bcrypt.GenerateFromPassword([]byte(pwdDigest), bcrypt.DefaultCost)
	if err != nil {
		zlog.GetLogger().Error("密码加密失败", zap.Error(err))
	}
	return string(pwd)
}

// CheckPassword 验证密码
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		zlog.GetLogger().Error("密码验证失败", zap.Error(err))
		return false
	}
	return true
}
