package models

// LoginResponse representa la respuesta del endpoint /auth/login
type LoginResponse struct {
	AdminID  int    `json:"admin_id"`
	Username string `json:"username"`
	Message  string `json:"message"`
	Error    string `json:"error"`
}