package request

type Login struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

type TOTPVerify struct {
	Code         string `json:"code" form:"code" binding:"required"`
	SessionToken string `json:"session_token" form:"session_token" binding:"required"`
}

type CreateUser struct {
	Username string `json:"username" form:"username" binding:"required"`
	Email    string `json:"email" form:"email" binding:"required,email"`
	Password string `json:"password" form:"password" binding:"required,min=8"`
	Role     string `json:"role" form:"role"`
}
