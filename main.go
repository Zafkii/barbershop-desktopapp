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
)

var (
	//go:embed assets/icoda.png
	iconBytes []byte
	//go:embed assets/lucid-icons/wifi.svg
	wifiSVG []byte
	//go:embed assets/lucid-icons/wifi-off.svg
	wifiOffSVG []byte
)

func main() {
	a := app.NewWithID("com.zafkiev.barbershop")
	a.SetIcon(fyne.NewStaticResource("icoda.png", iconBytes))
	w := a.NewWindow(appTitle)
	w.Resize(fyne.NewSize(960, 640))

	// Inicializar managers
	themeManager := theme.NewThemeManager(a, w)
	toastManager := toast.NewToastManager(w)
	apiClient := api.NewClient(serverURL)
	connChecker := connectivity.NewChecker(serverURL, 3*time.Second)

	// Callback de login
	var onLogin func(username, password string)
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

		// Éxito - cambiar a UI principal
		fyne.Do(func() {
			mainUI := ui.MakeMainUI(w, resp.Username, themeManager, toastManager, connChecker, func() {
				w.SetContent(ui.MakeLoginUI(w, themeManager, toastManager, connChecker, onLogin))
			})
			w.SetContent(mainUI)
			toastManager.Show("¡Bienvenido, "+resp.Username+"!", toast.ToastSuccess, 4*time.Second)
		})
	}

	// Pantalla inicial: login
	w.SetContent(ui.MakeLoginUI(w, themeManager, toastManager, connChecker, onLogin))
	w.ShowAndRun()

	// Limpiar al salir
	connChecker.Stop()
}