package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	appTheme "barbershop-desktop/theme"
	"barbershop-desktop/toast"

	"barbershop-desktop/connectivity"
)

const appTitle = "Barber Shop"

// makeThemeButton crea un botón de tema con icono dinámico (moon/sun/custom).
// Al tocarlo alterna solo entre dark y light; nunca entra a custom.
func makeThemeButton(tm *appTheme.ThemeManager) *widget.Button {
	var btn *widget.Button
	btn = widget.NewButtonWithIcon("", tm.IconResource(), func() {
		tm.Toggle()
		btn.SetIcon(tm.IconResource())
	})
	btn.Importance = widget.LowImportance
	return btn
}

// MakeLoginUI crea la pantalla de login
func MakeLoginUI(w fyne.Window, tm *appTheme.ThemeManager, toastMgr *toast.ToastManager, connChecker *connectivity.Checker, wifiSVG, wifiOffSVG []byte, onLogin func(username, password string)) fyne.CanvasObject {
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

	// Indicador de conectividad en el header del login
	connIndicator := NewConnectivityIndicator(wifiSVG, wifiOffSVG)
	connChecker.OnStatusChange(func(s connectivity.Status) {
		fyne.Do(func() {
			connIndicator.UpdateStatus(s)
		})
	})
	connChecker.Start()
	// Forzar actualización inicial
	connIndicator.UpdateStatus(connChecker.Status())

	headerRight := container.NewHBox(connIndicator, makeThemeButton(tm))
	headerBar := container.NewBorder(header, nil, nil, headerRight, nil)

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
		container.NewVBox(headerBar, subHeader, widget.NewSeparator()),
		nil, nil, nil,
		container.NewPadded(centered),
	)
}

// MakeMainUI crea la pantalla principal con 3 paneles:
// - Panel izquierdo: nombre del app + módulos + utilidades abajo
// - Panel superior: título del módulo + búsqueda + usuario con dropdown
// - Contenido: área del módulo seleccionado (vacío por ahora)
// bgSVG es el SVG del background con placeholders __WAVE_LINE_1__ y __WAVE_LINE_2__.
func MakeMainUI(w fyne.Window, username string, tm *appTheme.ThemeManager, pm *appTheme.PaletteManager, toastMgr *toast.ToastManager, connChecker *connectivity.Checker, wifiSVG, wifiOffSVG, userSVG, bgSVG []byte, onLogout func()) fyne.CanvasObject {

	// ---- Indicador de conectividad ----
	connIndicator := NewConnectivityIndicator(wifiSVG, wifiOffSVG)
	connChecker.OnStatusChange(func(s connectivity.Status) {
		fyne.Do(func() {
			connIndicator.UpdateStatus(s)
		})
	})
	connChecker.Start()
	connIndicator.UpdateStatus(connChecker.Status())

	// ---- Panel izquierdo ----
	appName := widget.NewLabelWithStyle("barber shop", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Lista de módulos
	type module struct {
		name string
	}
	modules := []module{
		{"Barberos"},
		{"Clientes"},
		{"Servicios"},
		{"Ventas"},
		{"Ajustes"},
	}

	// Título del módulo en el panel superior (se actualiza al seleccionar)
	moduleTitle := widget.NewLabelWithStyle("Barberos", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Background SVG colorizado con la paleta actual (hex)
	bgColored := colorizeBackground(bgSVG, pm.Wave1(), pm.Wave2())
	bgImg := canvas.NewImageFromResource(fyne.NewStaticResource("bg.svg", bgColored))
	bgImg.FillMode = canvas.ImageFillCover

	// Reconstruye el background con los colores actuales de la paleta
	refreshBg := func() {
		colored := colorizeBackground(bgSVG, pm.Wave1(), pm.Wave2())
		bgImg.Resource = fyne.NewStaticResource("bg.svg", colored)
		bgImg.Refresh()
	}

	// Contenido del módulo (vacío por ahora; Ajustes muestra el selector de paleta)
	moduleContent := container.NewStack()

	// Botones de módulo
	var moduleButtons []*widget.Button
	for _, m := range modules {
		m := m
		btn := widget.NewButton(m.name, func() {
			moduleTitle.SetText(m.name)
			moduleContent.RemoveAll()
			if m.name == "Ajustes" {
				moduleContent.Add(makeSettingsContent(pm, refreshBg, w.Canvas()))
			} else {
				moduleContent.Add(widget.NewLabel(""))
			}
		})
		btn.Importance = widget.LowImportance
		moduleButtons = append(moduleButtons, btn)
	}

	modulesList := container.NewVBox()
	for _, btn := range moduleButtons {
		modulesList.Add(btn)
	}

	// Utilidades abajo del panel izquierdo
	themeBtn := makeThemeButton(tm)
	bottomUtilities := container.NewHBox(connIndicator, themeBtn)

	leftPanel := container.NewBorder(
		appName,         // top
		bottomUtilities, // bottom
		nil, nil,
		modulesList, // center (módulos + espacio vacío abajo)
	)

	// ---- Panel superior ----
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Buscar...")

	// Sección de usuario: icono + nombre + dropdown
	userIconRes := fyne.NewStaticResource("user.svg", colorizeSVG(userSVG, "#E0E0E0"))
	userIcon := canvas.NewImageFromResource(userIconRes)
	userIcon.SetMinSize(fyne.NewSize(24, 24))
	userIcon.FillMode = canvas.ImageFillContain

	userLabel := widget.NewLabel(username)

	var userMenuBtn *widget.Button
	userMenuBtn = widget.NewButton("▼", func() {
		menu := fyne.NewMenu("",
			fyne.NewMenuItem("Cerrar sesión", onLogout),
		)
		popup := widget.NewPopUpMenu(menu, w.Canvas())
		popup.ShowAtPosition(fyne.NewPos(w.Canvas().Size().Width-180, 40))
	})
	userMenuBtn.Importance = widget.LowImportance

	userSection := container.NewHBox(userIcon, userLabel, userMenuBtn)

	topPanel := container.NewBorder(
		nil, nil,
		moduleTitle,  // left
		userSection,  // right
		searchEntry,  // center
	)

	// ---- Layout principal ----
	rightSide := container.NewBorder(topPanel, nil, nil, nil, moduleContent)
	appLayout := container.NewBorder(nil, nil, leftPanel, nil, rightSide)

	mainContent := container.NewStack(
		bgImg,
		appLayout,
		toastMgr.Overlay(),
	)

	return mainContent
}

// colorizeBackground reemplaza los placeholders del SVG de background con los colores dados.
func colorizeBackground(svg []byte, wave1, wave2 string) []byte {
	s := string(svg)
	s = strings.ReplaceAll(s, "__WAVE_LINE_1__", wave1)
	s = strings.ReplaceAll(s, "__WAVE_LINE_2__", wave2)
	return []byte(s)
}

// makeSettingsContent construye el contenido del módulo Ajustes:
// solo filas con etiqueta + cuadrado; el picker grande aparece como popup centrado al tocar.
func makeSettingsContent(pm *appTheme.PaletteManager, refreshBg func(), canvas fyne.Canvas) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Tema / Paleta de colores", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Abre el picker grande centrado en el contenedor para editar un color.
	// onChange se dispara en cada selección y actualiza swatch + paleta + background.
	openPicker := func(currentHex string, onChange func(string)) {
		picker := NewColorPicker(currentHex, onChange)
		popup := widget.NewPopUp(container.NewPadded(picker), canvas)
		canvasSize := canvas.Size()
		popupSize := popup.MinSize()
		x := (canvasSize.Width - popupSize.Width) / 2
		y := (canvasSize.Height - popupSize.Height) / 2
		if x < 0 {
			x = 0
		}
		if y < 0 {
			y = 0
		}
		popup.ShowAtPosition(fyne.NewPos(float32(x), float32(y)))
	}

	// --- Wave line 1 ---
	swatch1 := NewSwatch(pm.Wave1(), nil)
	swatch1.onTap = func() {
		openPicker(pm.Wave1(), func(hex string) {
			pm.SetWave1(hex)
			swatch1.SetColor(hex)
			refreshBg()
		})
	}
	wave1Row := container.NewBorder(nil, nil, widget.NewLabel("Línea de onda 1"), swatch1, nil)

	// --- Wave line 2 ---
	swatch2 := NewSwatch(pm.Wave2(), nil)
	swatch2.onTap = func() {
		openPicker(pm.Wave2(), func(hex string) {
			pm.SetWave2(hex)
			swatch2.SetColor(hex)
			refreshBg()
		})
	}
	wave2Row := container.NewBorder(nil, nil, widget.NewLabel("Línea de onda 2"), swatch2, nil)

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		wave1Row,
		widget.NewLabel(""),
		wave2Row,
	)

	return container.NewPadded(content)
}

// colorizeSVG reemplaza el stroke #000000 por el color dado.
func colorizeSVG(svg []byte, hexColor string) []byte {
	return []byte(strings.ReplaceAll(string(svg), `stroke="#000000"`, `stroke="`+hexColor+`"`))
}