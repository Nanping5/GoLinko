package request

type SendEmailCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}
