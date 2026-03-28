package model

import (
	"encoding/json"

	"gorm.io/gorm"
)

type GroupInfo struct {
	*gorm.Model
	Uuid      string          `gorm:"column:uuid;uniqueIndex;type:char(255);comment:'群组唯一id'"`
	Name      string          `gorm:"column:name;type:varchar(20);comment:'群组名称'"`
	Members   json.RawMessage `gorm:"column:members;type:json;comment:'群成员列表'"`
	Notice    string          `gorm:"column:notice;type:varchar(500);comment:'群公告'"`
	MemberCnt int             `gorm:"column:member_cnt;default:1;type:int;comment:'群成员数量'"` //默认群主1人
	OwnerId   string          `gorm:"column:owner_id;type:char(255);comment:'群主uuid'"`
	AddMode   int8            `gorm:"column:add_mode;default:0;type:tinyint;comment:'加群方式 0-公开 1-验证 '"`
	Avatar    string          `gorm:"column:avatar;type:varchar(255);comment:'群头像'"`
	Status    int8            `gorm:"column:status;type:tinyint;comment:'群状态 0-正常 1-解散'"`
}

func (GroupInfo) TableName() string {
	return "group_info"
}
