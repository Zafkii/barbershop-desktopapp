package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	appTitle   = "Barber Shop"
	serverURL  = "http://localhost:8080" // configurable via env if needed
)

var (
	//go:embed assets/icoda.png
	iconBytes []byte
)

type loginResponse struct {
	AdminID  int    `json:"admin_id"`
	Username string `json:"username"`
	Message  string `json:"message"`
	Error    string `json:"error"`
}

func main() {
	a := app.NewWithID("com.zafkiev.barbershop")
	a.SetIcon(fyne.NewStaticResource("icoda.png", iconBytes))
	w := a.NewWindow(appTitle)
	w.Resize(fyne.NewSize(960, 640))

	// Start with login screen
	w.SetContent(makeLoginUI(w))
	w.ShowAndRun()
}

func makeLoginUI(w fyne.Window) fyne.CanvasObject {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Usuario")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Contraseña")

	statusLabel := widget.NewLabel("")
	statusLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Define loginButton first so callbacks can reference it
	var loginButton *widget.Button
	loginButton = widget.NewButtonWithIcon("Iniciar sesión", theme.LoginIcon(), func() {
		username := usernameEntry.Text
		password := passwordEntry.Text

		if username == "" || password == "" {
			statusLabel.SetText("Usuario y contraseña requeridos")
			return
		}

		statusLabel.SetText("Conectando...")
		loginButton.Disable()

		go func() {
			resp, err := doLogin(username, password)
			fyne.Do(func() {
				loginButton.Enable()
				if err != nil {
					statusLabel.SetText("Error de conexión: " + err.Error())
					return
				}
				if resp.Error != "" {
					statusLabel.SetText(resp.Error)
					return
				}
				// Success - switch to main UI
				w.SetContent(makeMainUI(w, resp.Username))
			})
		}()
	})
	loginButton.Importance = widget.HighImportance

	// Enter key triggers login
	usernameEntry.OnSubmitted = func(_ string) { passwordEntry.FocusGained() }
	passwordEntry.OnSubmitted = func(_ string) { loginButton.OnTapped() }

	header := widget.NewLabelWithStyle(appTitle, fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Monospace: true})
	subHeader := widget.NewLabelWithStyle("Panel de administración", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	form := container.NewVBox(
		widget.NewLabel(""),
		usernameEntry,
		passwordEntry,
		widget.NewLabel(""),
		loginButton,
		statusLabel,
	)

	centered := container.NewCenter(form)
	return container.NewBorder(
		container.NewVBox(header, subHeader, widget.NewSeparator()),
		nil, nil, nil,
		container.NewPadded(centered),
	)
}

func makeMainUI(w fyne.Window, username string) fyne.CanvasObject {
	welcome := widget.NewLabelWithStyle(
		fmt.Sprintf("Bienvenido, %s", username),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	logoutBtn := widget.NewButtonWithIcon("Cerrar sesión", theme.LogoutIcon(), func() {
		w.SetContent(makeLoginUI(w))
	})
	logoutBtn.Importance = widget.LowImportance

	header := container.NewBorder(
		widget.NewLabelWithStyle(appTitle, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil, nil, logoutBtn,
		container.NewPadded(welcome),
	)

	// Placeholder for main functionality
	content := container.NewVBox(
		widget.NewLabel(""),
		widget.NewLabelWithStyle("Panel de Administración", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel(""),
		widget.NewLabel("Módulos disponibles:"),
		widget.NewButton("Gestionar Barberos", func() { dialog.ShowInformation("Próximamente", "Gestión de barberos", w) }),
		widget.NewButton("Gestionar Clientes", func() { dialog.ShowInformation("Próximamente", "Gestión de clientes", w) }),
		widget.NewButton("Gestionar Servicios", func() { dialog.ShowInformation("Próximamente", "Gestión de servicios", w) }),
		widget.NewButton("Ver Ventas", func() { dialog.ShowInformation("Próximamente", "Ventas y compras", w) }),
	)

	return container.NewBorder(header, nil, nil, nil, container.NewPadded(content))
}

func doLogin(username, password string) (loginResponse, error) {
	var resp loginResponse

	payload := map[string]string{
		"username": username,
		"password": password,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", serverURL+"/auth/login", bytes.NewBuffer(body))
	if err != nil {
		return resp, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	httpResp, err := client.Do(req)
	if err != nil {
		return resp, err
	}
	defer httpResp.Body.Close()

	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return resp, err
	}

	return resp, nil
}