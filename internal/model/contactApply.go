package model

import (
	"time"

	"gorm.io/gorm"
)

type ContactApply struct {
	*gorm.Model
	Uuid        string    `gorm:"column:uuid;uniqueIndex;not null;type:char(255);comment:'申请唯一id'"`
	UserId      string    `gorm:"column:user_id;index;not null;type:char(255);comment:'申请人id'"`
	ContactId   string    `gorm:"column:contact_id;index;not null;type:char(255);comment:'被申请id'"`
	ContactType int8      `gorm:"column:contact_type;not null;comment:'被申请类型，0.好友，1.群聊'"`
	Message     string    `gorm:"column:message;type:varchar(255);comment:'申请消息'"`
	Status      int8      `gorm:"column:status;not null;default:0;comment:'状态，0.待处理，1.同意，2.拒绝，3.拉黑'"`
	LastApplyAt time.Time `gorm:"column:last_apply_at;type:datetime;not null;comment:'最后一次申请时间'"`
}

func (ContactApply) TableName() string {
	return "contact_apply"
}
