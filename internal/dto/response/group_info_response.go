package response

type CreateGroupResponse struct {
	GroupId   string `json:"group_id"`
	GroupName string `json:"group_name"`
	Avatar    string `json:"avatar,omitempty"`
	Notice    string `json:"notice,omitempty"`
	AddMode   int8   `json:"add_mode"`
	OwnerId   string `json:"owner_id"`
	CreatedAt string `json:"created_at"`
}

type LoadMyGroupsResponse struct {
	GroupId   string `json:"group_id"`
	GroupName string `json:"group_name"`
	Avatar    string `json:"avatar,omitempty"`
	Notice    string `json:"notice,omitempty"`
	AddMode   int8   `json:"add_mode"`
	OwnerId   string `json:"owner_id"`
	CreatedAt string `json:"created_at"`
}

type GetGroupInfoResponse struct {
	GroupId   string `json:"group_id"`
	GroupName string `json:"group_name"`
	Avatar    string `json:"avatar,omitempty"`
	Notice    string `json:"notice,omitempty"`
	AddMode   int8   `json:"add_mode"`
	OwnerId   string `json:"owner_id"`
	Status    int8   `json:"status"`
	MemberCnt int    `json:"member_cnt"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
type GetGroupMembersResponse struct {
	UserId   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar,omitempty"`
}
