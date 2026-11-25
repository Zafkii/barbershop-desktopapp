package connectivity

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// Status representa el estado de conectividad
type Status int

const (
	StatusOffline Status = iota // Sin internet
	StatusServerDown           // Internet OK, server no responde
	StatusOnline               // Internet OK + server responde
)

// String devuelve el nombre del estado
func (s Status) String() string {
	switch s {
	case StatusOffline:
		return "offline"
	case StatusServerDown:
		return "server_down"
	case StatusOnline:
		return "online"
	default:
		return "unknown"
	}
}

// Checker verifica periódicamente la conectividad
type Checker struct {
	serverURL  string
	interval   time.Duration
	stopCh     chan struct{}
	status     Status
	mu         sync.RWMutex
	callbacks  []func(Status)
	cbMu       sync.Mutex
}

func NewChecker(serverURL string, interval time.Duration) *Checker {
	if interval == 0 {
		interval = 3 * time.Second
	}
	return &Checker{
		serverURL: serverURL,
		interval:  interval,
		stopCh:    make(chan struct{}),
		status:    StatusOffline,
	}
}

// Start inicia el loop de verificación
func (c *Checker) Start() {
	go func() {
		c.check()
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.check()
			case <-c.stopCh:
				return
			}
		}
	}()
}

// Stop detiene el loop
func (c *Checker) Stop() {
	close(c.stopCh)
}

// Status devuelve el último estado conocido
func (c *Checker) Status() Status {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

// OnStatusChange registra un callback que se llama cuando cambia el estado
func (c *Checker) OnStatusChange(cb func(Status)) {
	c.cbMu.Lock()
	defer c.cbMu.Unlock()
	c.callbacks = append(c.callbacks, cb)
}

func (c *Checker) check() {
	status := c.determineStatus()
	fmt.Printf("[connectivity] status: %v\n", status)
	
	c.mu.Lock()
	c.status = status
	c.mu.Unlock()
	
	// Always call callbacks to ensure UI updates
	c.cbMu.Lock()
	cbs := make([]func(Status), len(c.callbacks))
	copy(cbs, c.callbacks)
	c.cbMu.Unlock()
	
	for _, cb := range cbs {
		cb(status)
	}
}

func (c *Checker) determineStatus() Status {
	if !hasInternet() {
		return StatusOffline
	}
	if !serverReachable(c.serverURL) {
		return StatusServerDown
	}
	return StatusOnline
}

func hasInternet() bool {
	// Intentar múltiples endpoints HTTP para confirmar conectividad
	endpoints := []string{
		"http://www.google.com/generate_204",
		"http://www.cloudflare.com/cdn-cgi/trace",
		"http://detectportal.firefox.com/success.txt",
	}
	
	client := &http.Client{Timeout: 3 * time.Second}
	
	for _, url := range endpoints {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 204 || resp.StatusCode == 200 {
				return true
			}
		}
	}
	
	// Fallback: TCP a DNS
	for _, addr := range []string{"8.8.8.8:53", "1.1.1.1:53"} {
		conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
		if err == nil {
			conn.Close()
			return true
		}
	}
	
	return false
}

func serverReachable(serverURL string) bool {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(serverURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}