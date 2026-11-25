package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	appTheme "barbershop-desktop/theme"
	"barbershop-desktop/toast"
)

const appTitle = "Barber Shop"

// MakeLoginUI crea la pantalla de login
func MakeLoginUI(w fyne.Window, tm *appTheme.ThemeManager, toastMgr *toast.ToastManager, onLogin func(username, password string)) fyne.CanvasObject {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Usuario")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Contraseña")

	statusLabel := widget.NewLabel("")
	statusLabel.TextStyle = fyne.TextStyle{Italic: true}

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
			onLogin(username, password)
			// El callback manejará el resultado y cambiará de pantalla
			fyne.Do(func() {
				loginButton.Enable()
			})
		}()
	})
	loginButton.Importance = widget.HighImportance

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

// MakeMainUI crea la pantalla principal
func MakeMainUI(w fyne.Window, username string, tm *appTheme.ThemeManager, toastMgr *toast.ToastManager, onLogout func()) fyne.CanvasObject {
	titleLabel := widget.NewLabelWithStyle(appTitle, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Botón toggle tema
	var themeBtn *widget.Button
	themeBtn = widget.NewButtonWithIcon("", theme.ColorPaletteIcon(), func() {
		tm.Toggle()
		themeBtn.SetText(tm.ThemeButtonText())
	})
	themeBtn.SetText(tm.ThemeButtonText())
	themeBtn.Importance = widget.LowImportance

	logoutBtn := widget.NewButtonWithIcon("Cerrar sesión", theme.LogoutIcon(), onLogout)
	logoutBtn.Importance = widget.LowImportance

	headerRight := container.NewHBox(themeBtn, logoutBtn)
	header := container.NewBorder(
		titleLabel,
		nil, nil, headerRight,
		container.NewPadded(widget.NewLabel("")),
	)

	content := container.NewVBox(
		widget.NewLabel(""),
		widget.NewLabelWithStyle("Panel de Administración", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel(""),
		widget.NewLabel("Módulos disponibles:"),
		widget.NewButton("Gestionar Barberos", func() { 
			toastMgr.Show("Módulo Barberos - Próximamente", toast.ToastInfo, 3*time.Second)
		}),
		widget.NewButton("Gestionar Clientes", func() { 
			toastMgr.Show("Módulo Clientes - Próximamente", toast.ToastInfo, 3*time.Second)
		}),
		widget.NewButton("Gestionar Servicios", func() { 
			toastMgr.Show("Módulo Servicios - Próximamente", toast.ToastInfo, 3*time.Second)
		}),
		widget.NewButton("Ver Ventas", func() { 
			toastMgr.Show("Módulo Ventas - Próximamente", toast.ToastInfo, 3*time.Second)
		}),
	)

	mainContent := container.NewStack(
		container.NewPadded(content),
		toastMgr.Overlay(),
	)

	return container.NewBorder(header, nil, nil, nil, mainContent)
}