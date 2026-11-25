package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	_ "embed"

	"barbershop-desktop/api"
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

	// Callback de login
	var onLogin func(username, password string)
	onLogin = func(username, password string) {
		resp, err := apiClient.Login(username, password)
		if err != nil {
			// Mostrar error en la UI de login - necesitamos acceso al statusLabel
			// Por simplicidad, mostramos toast de error
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
			w.SetContent(ui.MakeMainUI(w, resp.Username, themeManager, toastManager, func() {
				w.SetContent(ui.MakeLoginUI(w, themeManager, toastManager, onLogin))
			}))
			toastManager.Show("¡Bienvenido, "+resp.Username+"!", toast.ToastSuccess, 4*time.Second)
		})
	}

	// Pantalla inicial: login
	w.SetContent(ui.MakeLoginUI(w, themeManager, toastManager, onLogin))
	w.ShowAndRun()
}