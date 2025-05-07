package schema

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Authenticated bool `json:"authenticated"`
}

type StatusLoginResponse struct {
	Authenticated bool   `json:"authenticated"`
	Username      string `json:"username"`
}

type LogoutResponse struct{}
