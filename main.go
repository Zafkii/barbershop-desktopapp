package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	fynetheme "fyne.io/fyne/v2/theme"

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
	//go:embed assets/bg-ui/oooscillate.svg
	bgSVG []byte
)

func main() {
	a := app.NewWithID("com.zafkiev.barbershop")
	a.SetIcon(fyne.NewStaticResource("icoda.png", iconBytes))
	w := a.NewWindow(appTitle)
	w.Resize(fyne.NewSize(960, 640))

	// Inicializar managers
	themeManager := theme.NewThemeManager(a, w, moonSVG, sunSVG, customSVG)
	paletteManager := theme.NewPaletteManager(a.Preferences())
	toastManager := toast.NewToastManager(w)
	apiClient := api.NewClient(serverURL)
	connChecker := connectivity.NewChecker(serverURL, 3*time.Second)
	prefs := a.Preferences()

	// Background lazy: la ventana se muestra de inmediato con un placeholder
	// (color de fondo) y el SVG pesado se renderiza en segundo plano.
	bgLayer := container.NewStack(canvas.NewRectangle(fynetheme.Color(fynetheme.ColorNameBackground)))
	contentLayer := container.NewStack()
	root := container.NewStack(bgLayer, contentLayer)
	w.SetContent(root)

	// setContent cambia la pantalla actual (login / main) sin tocar el fondo.
	setContent := func(content fyne.CanvasObject) {
		contentLayer.RemoveAll()
		contentLayer.Add(content)
	}

	// Carga/actualiza el background rasterizando el SVG FUERA del hilo principal.
	// El hilo principal solo dibuja el bitmap resultante (rápido).
	var bgImg *canvas.Image
	refreshBg := func() {
		w1, w2 := paletteManager.Wave1(), paletteManager.Wave2()
		colored := ui.ColorizeBackground(bgSVG, w1, w2)
		go func() {
			size := w.Canvas().Size()
			img := rasterizeBackground(colored, int(size.Width), int(size.Height))
			if img == nil {
				return
			}
			fyne.Do(func() {
				if bgImg == nil {
					bgImg = canvas.NewImageFromImage(img)
					bgImg.FillMode = canvas.ImageFillCover
					bgLayer.RemoveAll()
					bgLayer.Add(bgImg)
				} else {
					bgImg.Image = img
					bgImg.Refresh()
				}
			})
		}()
	}

	// Dispara la carga del background en segundo plano tras mostrar la ventana.
	go func() {
		time.Sleep(150 * time.Millisecond)
		refreshBg()
	}()

	// Callback de login
	var onLogin func(username, password string)
	var startSessionRefresh func()

	// Mostrar la UI principal con el usuario dado
	showMainUI := func(username string) {
		fyne.Do(func() {
			mainUI := ui.MakeMainUI(w, username, themeManager, paletteManager, toastManager, connChecker, wifiSVG, wifiOffSVG, userSVG, refreshBg, func() {
				// Logout: limpiar token en memoria y en disco
				apiClient.Logout()
				prefs.SetString(prefToken, "")
				prefs.SetString(prefUser, "")
				setContent(ui.MakeLoginUI(w, themeManager, toastManager, connChecker, wifiSVG, wifiOffSVG, onLogin))
			})
			setContent(mainUI)
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
						setContent(ui.MakeLoginUI(w, themeManager, toastManager, connChecker, wifiSVG, wifiOffSVG, onLogin))
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
					setContent(ui.MakeLoginUI(w, themeManager, toastManager, connChecker, wifiSVG, wifiOffSVG, onLogin))
				})
			}
		}()
		// Mostrar login mientras se verifica en segundo plano
		setContent(ui.MakeLoginUI(w, themeManager, toastManager, connChecker, wifiSVG, wifiOffSVG, onLogin))
	} else {
		setContent(ui.MakeLoginUI(w, themeManager, toastManager, connChecker, wifiSVG, wifiOffSVG, onLogin))
	}

	w.ShowAndRun()

	// Limpiar al salir. NO cerramos sesión: el token persiste en disco
	// para no pedir login al reabrir dentro de los 7 días.
	connChecker.Stop()
}