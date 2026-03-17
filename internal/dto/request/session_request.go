package request

type CreateNewSessionRequest struct {
	SendId    string `json:"send_id" binding:"required"`
	ReceiveId string `json:"receive_id" binding:"required"`
}
