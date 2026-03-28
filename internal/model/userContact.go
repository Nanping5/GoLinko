package model

import "gorm.io/gorm"

type UserContact struct {
	*gorm.Model
	UserId      string `gorm:"column:user_id;index;not null;type:char(255);comment:用户id"`
	ContactId   string `gorm:"column:contact_id;index;not null;type:char(255);comment:联系人id"`
	ContactType int8   `gorm:"column:contact_type;not null;comment:联系人类型，0.好友，1.群聊"`
	Status      int8   `gorm:"column:status;not null;default:0;comment:状态，0.正常，1.拉黑，2.被拉黑，3.删除，4.被删除，5.被禁言，6.退出群聊，7.被移出群聊"`
}

// 建表表名单数
func (UserContact) TableName() string {
	return "user_contact"
}
