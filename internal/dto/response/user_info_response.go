package response

type UserUpdateUserInfoResponse struct {
	Uuid      string  `json:"user_id"`
	Nickname  *string `json:"nickname,omitempty"`
	Avatar    *string `json:"avatar,omitempty"`
	Birthday  *string `json:"birthday,omitempty"`
	Gender    *int8   `json:"gender,omitempty"`
	Signature *string `json:"signature,omitempty"`
	Telephone *string `json:"telephone,omitempty"`
}

type UserRegisterResponse struct {
	Uuid      string `json:"user_id"`
	Nickname  string `json:"nickname"`
	Telephone string `json:"telephone"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
	IsAdmin   int8   `json:"is_admin"`
	Status    int8   `json:"status"`
	CreatedAt string `json:"created_at"`
}

type UserLoginResponse struct {
	Uuid      string `json:"user_id"`
	Nickname  string `json:"nickname"`
	Telephone string `json:"telephone"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
	Gender    int8   `json:"gender"`
	Signature string `json:"signature"`
	Birthday  string `json:"birthday"`
	IsAdmin   int8   `json:"is_admin"`
	Status    int8   `json:"status"`
	CreatedAt string `json:"created_at"`
	Token     string `json:"token,omitempty"`
}

type GetUserByIdResponse struct {
	Uuid      string `json:"user_id"`
	Telephone string `json:"telephone,omitempty"`
	Email     string `json:"email,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
	Avatar    string `json:"avatar,omitempty"`
	Birthday  string `json:"birthday,omitempty"`
	Gender    int8   `json:"gender,omitempty"`
	Signature string `json:"signature,omitempty"`
	IsAdmin   int8   `json:"is_admin,omitempty"`
	Status    int8   `json:"status,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}
