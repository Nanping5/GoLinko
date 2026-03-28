package model

import "gorm.io/gorm"

type Message struct {
	*gorm.Model
	Uuid       string `gorm:"column:uuid;uniqueIndex;not null;type:char(255);comment:'消息唯一id'"`
	SessionId  string `gorm:"column:session_id;index;not null;type:char(255);comment:'会话id'"`
	Type       int8   `gorm:"column:type;not null;comment:'消息类型，0.文本，1.语音，2.文件，3.通话'"`
	Content    string `gorm:"column:content;type:text;comment:'消息内容，文本消息为文本内容'"`
	Url        string `gorm:"column:url;type:varchar(255);comment:'消息url'"`
	SendId     string `gorm:"column:send_id;index;not null;type:char(255);comment:'发送者id'"`
	SendName   string `gorm:"column:send_name;type:varchar(20);comment:'发送者名称'"`
	SendAvatar string `gorm:"column:send_avatar;type:varchar(255);comment:'发送者头像'"`
	ReceiveId  string `gorm:"column:receive_id;index;not null;type:char(255);comment:'接收者id'"`
	FileType   string `gorm:"column:file_type;type:varchar(50);comment:'文件类型'"`
	FileName   string `gorm:"column:file_name;type:varchar(255);comment:'文件名称'"`
	FileSize   int64  `gorm:"column:file_size;type:bigint;comment:'文件大小'"`
	Status     int8   `gorm:"column:status;not null;default:0;comment:'消息状态，0.未发送，1.已发送'"`
	AVdata     string `gorm:"column:av_data;type:varchar(255);comment:'通话传递数据'"`
}

func (Message) TableName() string {
	return "message"
}
