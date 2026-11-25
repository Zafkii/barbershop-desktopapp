package theme

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ThemeManager maneja el cambio entre tema claro/oscuro
type ThemeManager struct {
	app    fyne.App
	window fyne.Window
	isDark bool
}

func NewThemeManager(a fyne.App, w fyne.Window) *ThemeManager {
	tm := &ThemeManager{
		app:    a,
		window: w,
		isDark: true, // Por defecto tema oscuro
	}
	a.Settings().SetTheme(tm.currentTheme())
	return tm
}

func (tm *ThemeManager) currentTheme() fyne.Theme {
	if tm.isDark {
		return theme.DarkTheme()
	}
	return theme.LightTheme()
}

func (tm *ThemeManager) Toggle() {
	tm.isDark = !tm.isDark
	tm.app.Settings().SetTheme(tm.currentTheme())
}

func (tm *ThemeManager) IsDark() bool {
	return tm.isDark
}

func (tm *ThemeManager) ThemeButtonText() string {
	if tm.IsDark() {
		return "☀️ Claro"
	}
	return "🌙 Oscuro"
}