package ui

import (
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

const (
	svSquareHeight = 180
	hueBarHeight   = 18
)

// ColorPicker es un selector HSV: cuadrado saturación×valor + barra de hue arcoíris.
// onChange recibe el color seleccionado en formato hex (#RRGGBB).
type ColorPicker struct {
	widget.BaseWidget
	hue      float64 // 0..360
	sat      float64 // 0..1
	val      float64 // 0..1
	onChange func(hex string)
}

// NewColorPicker crea un selector inicializado con el color hex dado.
func NewColorPicker(hex string, onChange func(string)) *ColorPicker {
	h, s, v := hexToHSV(hex)
	cp := &ColorPicker{
		hue:      h,
		sat:      s,
		val:      v,
		onChange: onChange,
	}
	cp.ExtendBaseWidget(cp)
	return cp
}

func (cp *ColorPicker) CreateRenderer() fyne.WidgetRenderer {
	square := canvas.NewRaster(cp.renderSVSquare)
	square.SetMinSize(fyne.NewSize(10, svSquareHeight))

	hueBar := canvas.NewRaster(cp.renderHueBar)
	hueBar.SetMinSize(fyne.NewSize(10, hueBarHeight))

	preview := canvas.NewRectangle(cp.currentColor())
	preview.SetMinSize(fyne.NewSize(10, hueBarHeight))

	// Indicadores circulares (uno en el cuadrado SV, otro en la barra de hue)
	svMarker := canvas.NewCircle(color.White)
	svMarker.StrokeColor = color.Black
	svMarker.StrokeWidth = 1

	hueMarker := canvas.NewCircle(color.White)
	hueMarker.StrokeColor = color.Black
	hueMarker.StrokeWidth = 1

	r := &colorPickerRenderer{
		cp:        cp,
		square:    square,
		hueBar:    hueBar,
		preview:   preview,
		svMarker:  svMarker,
		hueMarker: hueMarker,
	}
	return r
}

// Tapped maneja el clic: coordenadas relativas al widget.
func (cp *ColorPicker) Tapped(e *fyne.PointEvent) {
	cp.updateFromPosition(e.Position.X, e.Position.Y)
}

// Dragged permite mover el indicador arrastrando.
func (cp *ColorPicker) Dragged(e *fyne.DragEvent) {
	cp.updateFromPosition(e.Position.X, e.Position.Y)
}

// DragEnd no-op (la actualización ya se hizo durante el arrastre).
func (cp *ColorPicker) DragEnd() {}

func (cp *ColorPicker) updateFromPosition(px, py float32) {
	size := cp.Size()
	sqHeight := float32(svSquareHeight)

	if py <= sqHeight {
		// Clic/arrastre en el cuadrado SV
		cp.sat = clamp01(float64(px) / float64(size.Width))
		cp.val = clamp01(1 - float64(py)/float64(sqHeight))
	} else {
		// Clic/arrastre en la barra de hue
		cp.hue = clamp360(float64(px) / float64(size.Width) * 360)
	}
	cp.Refresh()
	if cp.onChange != nil {
		cp.onChange(cp.Hex())
	}
}

// Hex devuelve el color actual en formato #RRGGBB.
func (cp *ColorPicker) Hex() string {
	r, g, b := hsvToRGB(cp.hue, cp.sat, cp.val)
	return rgbToHex(r, g, b)
}

func (cp *ColorPicker) currentColor() color.Color {
	r, g, b := hsvToRGB(cp.hue, cp.sat, cp.val)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

func (cp *ColorPicker) renderSVSquare(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		v := 1 - float64(y)/float64(h)
		for x := 0; x < w; x++ {
			s := float64(x) / float64(w)
			r, g, b := hsvToRGB(cp.hue, s, v)
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

func (cp *ColorPicker) renderHueBar(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		hue := float64(x) / float64(w) * 360
		r, g, b := hsvToRGB(hue, 1, 1)
		for y := 0; y < h; y++ {
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

type colorPickerRenderer struct {
	cp        *ColorPicker
	square    *canvas.Raster
	hueBar    *canvas.Raster
	preview   *canvas.Rectangle
	svMarker  *canvas.Circle
	hueMarker *canvas.Circle
}

func (r *colorPickerRenderer) Layout(size fyne.Size) {
	sqHeight := float32(svSquareHeight)
	r.square.Resize(fyne.NewSize(size.Width, sqHeight))
	r.square.Move(fyne.NewPos(0, 0))

	r.hueBar.Resize(fyne.NewSize(size.Width, hueBarHeight))
	r.hueBar.Move(fyne.NewPos(0, sqHeight))

	r.preview.Resize(fyne.NewSize(size.Width, hueBarHeight))
	r.preview.Move(fyne.NewPos(0, sqHeight))

	r.layoutMarkers(size)
}

func (r *colorPickerRenderer) layoutMarkers(size fyne.Size) {
	sqHeight := float32(svSquareHeight)
	// Marcador del cuadrado SV
	svX := float32(r.cp.sat) * size.Width
	svY := (1 - float32(r.cp.val)) * sqHeight
	r.svMarker.Resize(fyne.NewSize(12, 12))
	r.svMarker.Move(fyne.NewPos(svX-6, svY-6))

	// Marcador de la barra de hue
	hueX := float32(r.cp.hue) / 360 * size.Width
	hueY := sqHeight + hueBarHeight/2
	r.hueMarker.Resize(fyne.NewSize(12, 12))
	r.hueMarker.Move(fyne.NewPos(hueX-6, hueY-6))
}

func (r *colorPickerRenderer) MinSize() fyne.Size {
	return fyne.NewSize(220, svSquareHeight+hueBarHeight)
}

func (r *colorPickerRenderer) Refresh() {
	r.square.Refresh()
	r.hueBar.Refresh()
	r.preview.FillColor = r.cp.currentColor()
	r.preview.Refresh()
	r.layoutMarkers(r.cp.Size())
}

func (r *colorPickerRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.square, r.hueBar, r.svMarker, r.hueMarker}
}

func (r *colorPickerRenderer) Destroy() {}

// ---- utilidades HSV ----

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func clamp360(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 360 {
		return 360
	}
	return v
}

// hsvToRGB convierte HSV (h en grados, s/v en 0..1) a RGB 0..255.
func hsvToRGB(h, s, v float64) (uint8, uint8, uint8) {
	c := v * s
	sec := math.Mod(h/60, 2)
	x := c * (1 - math.Abs(sec-1))
	m := v - c

	var r1, g1, b1 float64
	switch {
	case h < 60:
		r1, g1, b1 = c, x, 0
	case h < 120:
		r1, g1, b1 = x, c, 0
	case h < 180:
		r1, g1, b1 = 0, c, x
	case h < 240:
		r1, g1, b1 = 0, x, c
	case h < 300:
		r1, g1, b1 = x, 0, c
	default:
		r1, g1, b1 = c, 0, x
	}
	r := uint8((r1 + m) * 255)
	g := uint8((g1 + m) * 255)
	b := uint8((b1 + m) * 255)
	return r, g, b
}

// hexToHSV convierte #RRGGBB a HSV (h grados, s/v 0..1).
func hexToHSV(hex string) (float64, float64, float64) {
	r, g, b := hexToRGB(hex)
	return rgbToHSV(r, g, b)
}

func rgbToHSV(r, g, b uint8) (float64, float64, float64) {
	rf := float64(r) / 255
	gf := float64(g) / 255
	bf := float64(b) / 255

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	delta := max - min

	var h float64
	switch {
	case delta == 0:
		h = 0
	case max == rf:
		h = math.Mod(60*(gf-bf)/delta, 360)
	case max == gf:
		h = 60*(bf-rf)/delta + 120
	default:
		h = 60*(rf-gf)/delta + 240
	}
	if h < 0 {
		h += 360
	}

	s := 0.0
	if max != 0 {
		s = delta / max
	}
	return h, s, max
}

func hexToRGB(hex string) (uint8, uint8, uint8) {
	hex = trimHash(hex)
	if len(hex) != 6 {
		return 0, 0, 0
	}
	var r, g, b uint8
	for i := 0; i < 6; i += 2 {
		v := hexNibble(hex[i])<<4 | hexNibble(hex[i+1])
		switch i {
		case 0:
			r = v
		case 2:
			g = v
		case 4:
			b = v
		}
	}
	return r, g, b
}

func hexNibble(c byte) uint8 {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	default:
		return 0
	}
}

func trimHash(s string) string {
	if len(s) > 0 && s[0] == '#' {
		return s[1:]
	}
	return s
}

func rgbToHex(r, g, b uint8) string {
	const hex = "0123456789abcdef"
	buf := []byte{'#', 0, 0, 0, 0, 0, 0}
	buf[1] = hex[r>>4]
	buf[2] = hex[r&0xf]
	buf[3] = hex[g>>4]
	buf[4] = hex[g&0xf]
	buf[5] = hex[b>>4]
	buf[6] = hex[b&0xf]
	return string(buf)
}

// hexToColor convierte #RRGGBB a color.RGBA.
func hexToColor(hex string) color.Color {
	r, g, b := hexToRGB(hex)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

// Swatch es un cuadrado pequeño que muestra un color y es clickeable.
type Swatch struct {
	widget.BaseWidget
	rect  *canvas.Rectangle
	onTap func()
}

// NewSwatch crea un swatch del color hex dado. onTap se dispara al tocarlo.
func NewSwatch(hex string, onTap func()) *Swatch {
	s := &Swatch{
		rect:  canvas.NewRectangle(hexToColor(hex)),
		onTap: onTap,
	}
	s.ExtendBaseWidget(s)
	return s
}

// SetColor actualiza el color mostrado.
func (s *Swatch) SetColor(hex string) {
	s.rect.FillColor = hexToColor(hex)
	s.rect.Refresh()
}

func (s *Swatch) CreateRenderer() fyne.WidgetRenderer {
	return &swatchRenderer{s: s}
}

func (s *Swatch) Tapped(_ *fyne.PointEvent) {
	if s.onTap != nil {
		s.onTap()
	}
}

type swatchRenderer struct {
	s *Swatch
}

func (r *swatchRenderer) Layout(size fyne.Size) {
	r.s.rect.Resize(size)
	r.s.rect.Move(fyne.NewPos(0, 0))
}

func (r *swatchRenderer) MinSize() fyne.Size {
	return fyne.NewSize(48, 24)
}

func (r *swatchRenderer) Refresh() {
	r.s.rect.Refresh()
}

func (r *swatchRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.s.rect}
}

func (r *swatchRenderer) Destroy() {}