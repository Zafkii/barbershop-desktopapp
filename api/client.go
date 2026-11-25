package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"

	"barbershop-desktop/models"
)

const defaultServerURL = "http://localhost:8080"

type Client struct {
	baseURL string
	http    *http.Client
	token   string
	mu      sync.RWMutex
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

// SetToken stores the session token in memory
func (c *Client) SetToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
}

// Token returns the current session token
func (c *Client) Token() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.token
}

// HasSession returns true if a token is stored
func (c *Client) HasSession() bool {
	return c.Token() != ""
}

// clearToken removes the session token
func (c *Client) clearToken() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = ""
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

	if resp.Token != "" {
		c.SetToken(resp.Token)
	}

	return resp, nil
}

// VerifySession checks if the current session is still valid
func (c *Client) VerifySession() (bool, error) {
	token := c.Token()
	if token == "" {
		return false, nil
	}

	req, err := http.NewRequest("GET", c.baseURL+"/auth/verify", nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	httpResp, err := c.http.Do(req)
	if err != nil {
		return false, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		c.clearToken()
		return false, nil
	}

	return true, nil
}

// Logout invalidates the session on the server
func (c *Client) Logout() {
	token := c.Token()
	if token == "" {
		return
	}

	req, _ := http.NewRequest("POST", c.baseURL+"/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	c.http.Do(req)

	c.clearToken()
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