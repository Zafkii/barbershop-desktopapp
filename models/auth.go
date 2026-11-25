package models

// LoginResponse representa la respuesta del endpoint /auth/login
type LoginResponse struct {
	AdminID  int    `json:"admin_id"`
	Username string `json:"username"`
	Token    string `json:"token"`
	Message  string `json:"message"`
	Error    string `json:"error"`
}