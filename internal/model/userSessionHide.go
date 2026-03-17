package model

import "gorm.io/gorm"

type UserSessionHide struct {
	*gorm.Model
	UserId    string `gorm:"column:user_id;not null;type:char(255);uniqueIndex:idx_user_session_hide;comment:'用户id'"`
	SessionId string `gorm:"column:session_id;not null;type:char(255);uniqueIndex:idx_user_session_hide;comment:'会话id'"`
}

func (UserSessionHide) TableName() string {
	return "user_session_hide"
}
