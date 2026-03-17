package response

import "encoding/json"

type LoadMyJoinedGroupsResponse struct {
	GroupId   string `json:"group_id"`
	GroupName string `json:"group_name"`
	Avatar    string `json:"avatar,omitempty"`
}

type UserListResponse struct {
	UserId   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar,omitempty"`
}

type GetContactInfoResponse struct {
	ContactId        string          `json:"contact_id"`
	ContactName      string          `json:"contact_name"`
	ContactAvatar    string          `json:"contact_avatar"`
	ContactPhone     string          `json:"contact_phone"`
	ContactEmail     string          `json:"contact_email"`
	ContactGender    int8            `json:"contact_gender"`
	ContactSignature string          `json:"contact_signature"`
	ContactBirthday  string          `json:"contact_birthday"`
	ContactNotice    string          `json:"contact_notice"`
	ContactMembers   json.RawMessage `json:"contact_members"`
	ContactMemberCnt int             `json:"contact_member_cnt"`
	ContactOwnerId   string          `json:"contact_owner_id"`
	ContactAddMode   int8            `json:"contact_add_mode"`
}

type ContactApplyListResponse struct {
	ApplyId     string `json:"apply_id"`
	ApplicantId string `json:"applicant_id"`
	Applicant   string `json:"applicant"`
	Avatar      string `json:"avatar,omitempty"`
	Message     string `json:"message,omitempty"`
	ContactType int8   `json:"contact_type"` // 0: 好友, 1: 群聊
	TargetName  string `json:"target_name"`  // 群名或"个人"
	CreatedAt   string `json:"created_at"`
}
