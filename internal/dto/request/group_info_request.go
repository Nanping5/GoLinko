package request

type CreateGroupRequest struct {
	Notice    string `json:"notice" binding:"omitempty,max=100"`
	AddMode   *int8  `json:"add_mode" binding:"required,oneof=0 1 2"`
	GroupName string `json:"group_name" binding:"required,min=2,max=20"`
	Avatar    string `json:"avatar" binding:"omitempty,url"`
}

type LoadMyGroupsRequest struct {
}

type EnterGroupDirectlyRequest struct {
	GroupId string `json:"group_id" binding:"required,uuid4"`
}

type LeaveGroupRequest struct {
	GroupId string `json:"group_id" binding:"required,uuid4"`
}

type DismissGroupRequest struct {
	GroupId string `json:"group_id" binding:"required,uuid4"`
}

type UpdateGroupInfoRequest struct {
	GroupId   string  `json:"group_id" binding:"required,uuid4"`
	GroupName *string `json:"group_name" binding:"omitempty,min=2,max=20"`
	Avatar    *string `json:"avatar" binding:"omitempty,url"`
	Notice    *string `json:"notice" binding:"omitempty,max=100"`
	AddMode   *int8   `json:"add_mode" binding:"omitempty,oneof=0 1 2"`
}
type RemoveGroupMemberRequest struct {
	GroupId string `json:"group_id" form:"group_id" binding:"required,uuid4"`
	UserId  string `json:"user_id" form:"user_id" binding:"required,uuid4"`
}
