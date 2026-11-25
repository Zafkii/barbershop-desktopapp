package api

import (
	"bytes"
	"encoding/json"
	"net/http"

	"barbershop-desktop/models"
)

const defaultServerURL = "http://localhost:8080"

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultServerURL
	}
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{},
	}
}

// Login realiza la autenticación contra el server
func (c *Client) Login(username, password string) (models.LoginResponse, error) {
	var resp models.LoginResponse

	payload := map[string]string{
		"username": username,
		"password": password,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", c.baseURL+"/auth/login", bytes.NewBuffer(body))
	if err != nil {
		return resp, err
	}
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := c.http.Do(req)
	if err != nil {
		return resp, err
	}
	defer httpResp.Body.Close()

	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return resp, err
	}

	return resp, nil
}

// HealthCheck verifica si el server está disponible
func (c *Client) HealthCheck() error {
	resp, err := c.http.Get(c.baseURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}