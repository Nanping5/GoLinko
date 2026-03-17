package response

type GetSessionListResponse struct {
	SessionId   string `json:"session_id"`
	ReceiveId   string `json:"receive_id"`
	ReceiveName string `json:"receive_name"`
	Avatar      string `json:"avatar"`
}

type GetGroupSessionListResponse struct {
	SessionId string `json:"session_id"`
	GroupId   string `json:"group_id"`
	GroupName string `json:"group_name"`
	Avatar    string `json:"avatar"`
}

type CheckOpenSessionAllowedResponse struct {
	Allowed bool `json:"allowed"`
}
