package model

import "gorm.io/gorm"

type Session struct {
	*gorm.Model
	Uuid string `gorm:"column:uuid;uniqueIndex;not null;type:char(255);comment:'会话唯一uuid'"`
	// PairKey 用于单聊强制规范化 ID 组合 (如 min_max)，确保物理唯一；群聊可直接存 group_uuid 或特定前缀
	PairKey string `gorm:"column:pair_key;uniqueIndex;not null;type:char(255);comment:'对端唯一标识键'"`
	// 为了兼容现有代码，暂时保留 SendId/ReceiveId 记录创建时的方向，但逻辑不再依赖它们做唯一性
	SendId      string `gorm:"column:send_id;not null;type:char(255);comment:'发起者id'"`
	ReceiveId   string `gorm:"column:receive_id;not null;type:char(255);comment:'接收者id'"`
	ReceiveName string `gorm:"column:receive_name;type:varchar(20);comment:'名称'"`
	Avatar      string `gorm:"column:avatar;type:varchar(255);comment:'会话头像'"`
}

func (Session) TableName() string {
	return "session"
}
