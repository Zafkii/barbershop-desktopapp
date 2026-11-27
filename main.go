package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	_ "embed"

	"barbershop-desktop/api"
	"barbershop-desktop/connectivity"
	"barbershop-desktop/theme"
	"barbershop-desktop/toast"
	"barbershop-desktop/ui"
)

const (
	appTitle  = "Barber Shop"
	serverURL = "http://localhost:8080"

	prefToken   = "session_token"
	prefUser    = "session_username"
)

var (
	//go:embed assets/icoda.png
	iconBytes []byte
	//go:embed assets/lucid-icons/wifi.svg
	wifiSVG []byte
	//go:embed assets/lucid-icons/wifi-off.svg
	wifiOffSVG []byte
	//go:embed assets/lucid-icons/moon.svg
	moonSVG []byte
	//go:embed assets/lucid-icons/sun.svg
	sunSVG []byte
	//go:embed assets/lucid-icons/paintbrush-vertical.svg
	customSVG []byte
	//go:embed assets/lucid-icons/user.svg
	userSVG []byte
)

func main() {
	a := app.NewWithID("com.zafkiev.barbershop")
	a.SetIcon(fyne.NewStaticResource("icoda.png", iconBytes))
	w := a.NewWindow(appTitle)
	w.Resize(fyne.NewSize(960, 640))

	// Inicializar managers
	themeManager := theme.NewThemeManager(a, w, moonSVG, sunSVG, customSVG)
	toastManager := toast.NewToastManager(w)
	apiClient := api.NewClient(serverURL)
	connChecker := connectivity.NewChecker(serverURL, 3*time.Second)
	prefs := a.Preferences()

	// Callback de login
	var onLogin func(username, password string)
	var startSessionRefresh func()

	// Mostrar la UI principal con el usuario dado
	showMainUI := func(username string) {
		fyne.Do(func() {
			mainUI := ui.MakeMainUI(w, username, themeManager, toastManager, connChecker, wifiSVG, wifiOffSVG, userSVG, func() {
				// Logout: limpiar token en memoria y en disco
				apiClient.Logout()
				prefs.SetString(prefToken, "")
				prefs.SetString(prefUser, "")
				w.SetContent(ui.MakeLoginUI(w, themeManager, toastManager, connChecker, wifiSVG, wifiOffSVG, onLogin))
			})
			w.SetContent(mainUI)
			toastManager.Show("¡Bienvenido, "+username+"!", toast.ToastSuccess, 4*time.Second)
		})
		startSessionRefresh()
	}

	onLogin = func(username, password string) {
		resp, err := apiClient.Login(username, password)
		if err != nil {
			fyne.Do(func() {
				toastManager.Show("Error de conexión: "+err.Error(), toast.ToastError, 5*time.Second)
			})
			return
		}
		if resp.Error != "" {
			fyne.Do(func() {
				toastManager.Show(resp.Error, toast.ToastError, 4*time.Second)
			})
			return
		}

		// Persistir token y usuario en disco (Fyne preferences)
		prefs.SetString(prefToken, resp.Token)
		prefs.SetString(prefUser, resp.Username)

		showMainUI(resp.Username)
	}

	// Refresh periódico de sesión (cada 1 hora) para mantener la sesión viva
	// más allá de los 7 días mientras la app esté abierta
	startSessionRefresh = func() {
		go func() {
			ticker := time.NewTicker(1 * time.Hour)
			defer ticker.Stop()
			for range ticker.C {
				valid, err := apiClient.VerifySession()
				if err != nil || !valid {
					// Sesión expirada o inválida - volver a login
					fyne.Do(func() {
						apiClient.Logout()
						prefs.SetString(prefToken, "")
						prefs.SetString(prefUser, "")
						w.SetContent(ui.MakeLoginUI(w, themeManager, toastManager, connChecker, wifiSVG, wifiOffSVG, onLogin))
						toastManager.Show("Sesión expirada. Por favor inicia sesión nuevamente.", toast.ToastWarning, 5*time.Second)
					})
					return
				}
			}
		}()
	}

	// Al arrancar: si hay token guardado, validar contra el server.
	// Si es válido (dentro de los 7 días), entrar directo sin pedir login.
	token := prefs.String(prefToken)
	username := prefs.String(prefUser)
	if token != "" {
		apiClient.SetToken(token)
		go func() {
			valid, err := apiClient.VerifySession()
			if err == nil && valid {
				showMainUI(username)
			} else {
				// Token expirado o inválido - limpiar y mostrar login
				prefs.SetString(prefToken, "")
				prefs.SetString(prefUser, "")
				apiClient.Logout()
				fyne.Do(func() {
					w.SetContent(ui.MakeLoginUI(w, themeManager, toastManager, connChecker, wifiSVG, wifiOffSVG, onLogin))
				})
			}
		}()
		// Mostrar login mientras se verifica en segundo plano
		w.SetContent(ui.MakeLoginUI(w, themeManager, toastManager, connChecker, wifiSVG, wifiOffSVG, onLogin))
	} else {
		w.SetContent(ui.MakeLoginUI(w, themeManager, toastManager, connChecker, wifiSVG, wifiOffSVG, onLogin))
	}

	w.ShowAndRun()

	// Limpiar al salir. NO cerramos sesión: el token persiste en disco
	// para no pedir login al reabrir dentro de los 7 días.
	connChecker.Stop()
}