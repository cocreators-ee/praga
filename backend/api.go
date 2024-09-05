package backend

type emailVerifyRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=8"`
}

type emailSendRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type configResponse struct {
	Title   string `json:"title"`
	Brand   string `json:"brand"`
	Support string `json:"support"`
}
