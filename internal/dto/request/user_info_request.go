package request

type UserUpdateInfoRequest struct {
	Uuid      string  `json:"user_id"` // 去掉 binding:"required"，由 Token 提供
	Nickname  *string `json:"nickname" binding:"omitempty,min=2,max=20"`
	Avatar    *string `json:"avatar" binding:"omitempty,url"`
	Birthday  *string `json:"birthday" binding:"omitempty,datetime=2006-01-02"`
	Gender    *int8   `json:"gender" binding:"omitempty,oneof=0 1"`
	Signature *string `json:"signature" binding:"omitempty,max=100"`

	Telephone *string `json:"telephone" binding:"omitempty,len=11"`
}

type UserRegisterRequest struct {
	Nickname  string `json:"nickname" binding:"required,min=2,max=20"`
	Telephone string `json:"telephone" binding:"required,len=11,numeric"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6,max=20"`
	Code      string `json:"code" binding:"required,len=6"`
}

type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=20"`
}
type UserLoginByCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

type GetUserInfoRequest struct {
	Uuid string `json:"user_id" binding:"required,uuid4"`
}
