package auth

type signupDTO struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signinDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
