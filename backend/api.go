package backend

type EmailVerifyRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=8"`
}

type EmailSendRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ConfigResponse struct {
	Title   string `json:"title"`
	Brand   string `json:"brand"`
	Support string `json:"support"`
}
