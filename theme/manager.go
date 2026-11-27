package theme

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ThemeMode representa el modo de tema actual.
type ThemeMode int

const (
	ModeDark ThemeMode = iota
	ModeLight
	ModeCustom // "personalizado" - solo se setea desde settings, nunca por el botón toggle
)

// ThemeManager maneja el cambio entre tema claro/oscuro/personalizado.
type ThemeManager struct {
	app  fyne.App
	win  fyne.Window
	mode ThemeMode

	moonSVG   []byte
	sunSVG    []byte
	customSVG []byte
}

// NewThemeManager crea el manejador con los SVGs de iconos embebidos.
func NewThemeManager(a fyne.App, w fyne.Window, moonSVG, sunSVG, customSVG []byte) *ThemeManager {
	tm := &ThemeManager{
		app:       a,
		win:       w,
		mode:      ModeDark, // Por defecto tema oscuro
		moonSVG:   moonSVG,
		sunSVG:    sunSVG,
		customSVG: customSVG,
	}
	a.Settings().SetTheme(tm.currentTheme())
	return tm
}

func (tm *ThemeManager) currentTheme() fyne.Theme {
	if tm.mode == ModeLight {
		return theme.LightTheme()
	}
	// Dark y Custom usan la base oscura por ahora (personalizado aún no implementado)
	return theme.DarkTheme()
}

// Toggle alterna SOLO entre dark y light. Nunca entra a custom.
// Si está en custom, pasa a dark.
func (tm *ThemeManager) Toggle() {
	switch tm.mode {
	case ModeDark:
		tm.mode = ModeLight
	case ModeLight:
		tm.mode = ModeDark
	case ModeCustom:
		tm.mode = ModeDark
	}
	tm.app.Settings().SetTheme(tm.currentTheme())
}

// SetMode cambia el modo (usado por settings para "personalizado").
func (tm *ThemeManager) SetMode(m ThemeMode) {
	tm.mode = m
	tm.app.Settings().SetTheme(tm.currentTheme())
}

// Mode devuelve el modo actual.
func (tm *ThemeManager) Mode() ThemeMode {
	return tm.mode
}

// IsDark reporta si el tema base es oscuro.
func (tm *ThemeManager) IsDark() bool {
	return tm.mode != ModeLight
}

// IconResource devuelve el recurso SVG coloreado según el modo actual.
// dark -> moon, light -> sun, custom -> paintbrush.
func (tm *ThemeManager) IconResource() fyne.Resource {
	var svg []byte
	var color string
	var name string

	switch tm.mode {
	case ModeLight:
		// En modo claro se muestra la luna (azul)
		svg = tm.moonSVG
		color = "#2196F3" // azul
		name = "theme-moon.svg"
	case ModeCustom:
		svg = tm.customSVG
		color = "#9C27B0" // acento violeta
		name = "theme-custom.svg"
	default: // ModeDark
		// En modo oscuro se muestra el sol (amarillo)
		svg = tm.sunSVG
		color = "#FFC800" // amarillo
		name = "theme-sun.svg"
	}

	colored := strings.ReplaceAll(string(svg), `stroke="#000000"`, fmt.Sprintf(`stroke="%s"`, color))
	return fyne.NewStaticResource(name, []byte(colored))
}