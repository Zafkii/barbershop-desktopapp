package theme

import (
	"fyne.io/fyne/v2"
)

// PaletteManager maneja los colores de la paleta "background" (wave lines).
// Los colores se guardan como hex (#RRGGBB) y persisten en Fyne preferences.
type PaletteManager struct {
	prefs fyne.Preferences
	wave1 string // hex
	wave2 string // hex
}

const (
	prefWave1 = "palette_wave1"
	prefWave2 = "palette_wave2"

	// Defaults: verde y azul
	defaultWave1 = "#2ECC71"
	defaultWave2 = "#3498DB"
)

// NewPaletteManager crea el gestor leyendo (o inicializando) la paleta persistida.
func NewPaletteManager(prefs fyne.Preferences) *PaletteManager {
	m := &PaletteManager{prefs: prefs}

	m.wave1 = prefs.String(prefWave1)
	if m.wave1 == "" {
		m.wave1 = defaultWave1
	}
	m.wave2 = prefs.String(prefWave2)
	if m.wave2 == "" {
		m.wave2 = defaultWave2
	}
	return m
}

// Wave1 devuelve el color de la línea de onda 1 (hex).
func (m *PaletteManager) Wave1() string { return m.wave1 }

// Wave2 devuelve el color de la línea de onda 2 (hex).
func (m *PaletteManager) Wave2() string { return m.wave2 }

// SetWave1 fija y persiste el color de la línea 1.
func (m *PaletteManager) SetWave1(hex string) {
	m.wave1 = hex
	m.prefs.SetString(prefWave1, hex)
}

// SetWave2 fija y persiste el color de la línea 2.
func (m *PaletteManager) SetWave2(hex string) {
	m.wave2 = hex
	m.prefs.SetString(prefWave2, hex)
}