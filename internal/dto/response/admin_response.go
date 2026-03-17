package response

type GetUserListResponse struct {
	Uuid      string `json:"uuid"`
	Nickname  string `json:"nickname"`
	Telephone string `json:"telephone"`
	Email     string `json:"email"`
	IsAdmin   bool   `json:"is_admin"`
	Status    int8   `json:"status"`
	Avatar    string `json:"avatar"`
	CreatedAt string `json:"created_at"`
	Birthday  string `json:"birthday"`
	Gender    int8   `json:"gender"`
	Signature string `json:"signature"`
}
