package response

type GetMessageListResponse struct {
	SendId        string `json:"send_id"`
	SendName      string `json:"send_name"`
	SendAvatar    string `json:"send_avatar"`
	ReceiveId     string `json:"receive_id"`
	ReceiveName   string `json:"receive_name"`
	ReceiveAvatar string `json:"receive_avatar"`
	Content       string `json:"content"`
	Url           string `json:"url"`
	Type          int8   `json:"type"`
	FileType      string `json:"file_type"`
	FileName      string `json:"file_name"`
	FileSize      int64  `json:"file_size"`
	AVdata        string `json:"av_data"`
}

type GetGroupMessageListResponse struct {
	SendId      string `json:"send_id"`
	SendName    string `json:"send_name"`
	SendAvatar  string `json:"send_avatar"`
	GroupId     string `json:"group_id"`
	GroupName   string `json:"group_name"`
	GroupAvatar string `json:"group_avatar"`
	Content     string `json:"content"`
	Url         string `json:"url"`
	Type        int8   `json:"type"`
	FileType    string `json:"file_type"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	AVdata      string `json:"av_data"`
}
