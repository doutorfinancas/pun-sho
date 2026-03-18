package request

type Login struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

type TOTPVerify struct {
	Code         string `json:"code" form:"code"`
	SessionToken string `json:"session_token" form:"session_token"`
}

type CreateUser struct {
	Username string `json:"username" form:"username"`
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
	Role     string `json:"role" form:"role"`
}
