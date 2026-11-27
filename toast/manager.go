package toast

import (
	"image"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ToastType define el tipo de notificación
type ToastType int

const (
	ToastInfo ToastType = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// Toast representa una notificación individual
type Toast struct {
	message      string
	toastType    ToastType
	duration     time.Duration
	progress     *canvas.Rectangle
	container    *fyne.Container
	barMaxWidth  float32
}

// ToastManager maneja la visualización de toasts
type ToastManager struct {
	window       fyne.Window
	toastStack   *fyne.Container
	currentToast *Toast
}

func NewToastManager(w fyne.Window) *ToastManager {
	stack := container.NewStack()
	stack.Hide()
	
	return &ToastManager{
		window:     w,
		toastStack: stack,
	}
}

// Show muestra un toast con animación de barra de progreso
func (tm *ToastManager) Show(message string, toastType ToastType, duration time.Duration) {
	if tm.currentToast != nil {
		tm.dismissCurrent()
	}
	
	toast := tm.createToast(message, toastType, duration)
	tm.currentToast = toast
	
	tm.toastStack.RemoveAll()
	tm.toastStack.Add(toast.container)
	tm.toastStack.Show()
	tm.toastStack.Refresh()
	
	go toast.animateProgress(func() {
		fyne.Do(func() {
			tm.dismissCurrent()
		})
	})
}

func (tm *ToastManager) createToast(message string, toastType ToastType, duration time.Duration) *Toast {
	var barColor fyne.ThemeColorName
	switch toastType {
	case ToastSuccess:
		barColor = theme.ColorNameSuccess
	case ToastWarning:
		barColor = theme.ColorNameWarning
	case ToastError:
		barColor = theme.ColorNameError
	default:
		barColor = theme.ColorNamePrimary
	}
	
	label := widget.NewLabel(message)
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.Wrapping = fyne.TextWrapWord

	barTint := theme.Color(barColor)

	// Ancho útil de la barra: ancho del toast (400) menos 8px de margen a cada lado
	// para que no sobresalga en las esquinas redondeadas.
	const toastWidth = 400
	const cornerInset = 8
	barMaxWidth := float32(toastWidth - 2*cornerInset)

	progress := canvas.NewRectangle(barTint)
	progress.SetMinSize(fyne.NewSize(barMaxWidth, 3))

	// Insets transparentes a los lados para mantener la barra dentro del área redondeada
	leftInset := canvas.NewRectangle(color.Transparent)
	leftInset.SetMinSize(fyne.NewSize(cornerInset, 3))
	rightInset := canvas.NewRectangle(color.Transparent)
	rightInset.SetMinSize(fyne.NewSize(cornerInset, 3))

	bottomBar := container.NewHBox(leftInset, progress, rightInset)

	innerContent := container.NewBorder(nil, bottomBar, nil, nil, container.NewPadded(label))

	// Un solo RoundedRect que dibuja fondo + borde en el mismo raster.
	// Evita el problema de capas donde el fondo tapa el borde.
	bgRect := NewRoundedRectBordered(
		theme.Color(theme.ColorNameBackground),
		barTint,
		8,        // radio
		1,        // grosor del borde
	)
	bgRect.SetMinSize(fyne.NewSize(400, 50))

	outer := container.NewVBox(
		container.NewHBox(
			widget.NewLabel(""),
			container.NewStack(bgRect, innerContent),
			widget.NewLabel(""),
		),
	)
	
	return &Toast{
		message:     message,
		toastType:   toastType,
		duration:    duration,
		progress:    progress,
		container:   outer,
		barMaxWidth: barMaxWidth,
	}
}

// RoundedRect es un widget que dibuja un rectángulo con esquinas redondeadas usando canvas.Raster.
// Puede dibujar un borde del mismo radio (borderWidth > 0) en el mismo raster.
type RoundedRect struct {
	widget.BaseWidget
	fillColor   color.Color
	borderColor color.Color
	radius      float32
	borderWidth float32
	minSize     fyne.Size
}

func NewRoundedRect(c color.Color, radius float32) *RoundedRect {
	r := &RoundedRect{
		fillColor: c,
		radius:    radius,
	}
	r.ExtendBaseWidget(r)
	return r
}

// NewRoundedRectBordered crea un RoundedRect con borde dibujado en el raster.
func NewRoundedRectBordered(fill, border color.Color, radius, borderWidth float32) *RoundedRect {
	r := &RoundedRect{
		fillColor:   fill,
		borderColor: border,
		radius:      radius,
		borderWidth: borderWidth,
	}
	r.ExtendBaseWidget(r)
	return r
}

func (r *RoundedRect) SetMinSize(size fyne.Size) {
	r.minSize = size
}

func (r *RoundedRect) MinSize() fyne.Size {
	if r.minSize != fyne.NewSize(0, 0) {
		return r.minSize
	}
	return fyne.NewSize(2*r.radius, 2*r.radius)
}

func (r *RoundedRect) CreateRenderer() fyne.WidgetRenderer {
	raster := canvas.NewRaster(r.renderRoundedRect)
	raster.SetMinSize(r.MinSize())
	return &roundedRectRenderer{raster: raster, rect: r}
}

func (r *RoundedRect) renderRoundedRect(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	radius := r.radius
	borderWidth := r.borderWidth

	fillRGBA := color.RGBAModel.Convert(r.fillColor).(color.RGBA)
	var borderRGBA color.RGBA
	if r.borderColor != nil {
		borderRGBA = color.RGBAModel.Convert(r.borderColor).(color.RGBA)
	}

	// Rectángulo interior (dentro del borde), inset por borderWidth
	inset := int(borderWidth)
	innerRadius := radius - borderWidth
	if innerRadius < 0 {
		innerRadius = 0
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// ¿Está el pixel dentro del rectángulo exterior redondeado?
			if !pointInRoundedRect(x, y, w, h, radius) {
				continue // transparente
			}

			// Si hay borde, verificar si el pixel está en la zona del borde
			if borderWidth > 0 && !pointInRoundedRect(x-inset, y-inset, w-2*inset, h-2*inset, innerRadius) {
				img.SetRGBA(x, y, borderRGBA)
				continue
			}

			img.SetRGBA(x, y, fillRGBA)
		}
	}

	return img
}

// pointInRoundedRect reporta si (x,y) está dentro de un rectángulo redondeado
// de tamaño w×h con esquinas de radio r, en coordenadas locales (0,0 = esquina sup-izq).
func pointInRoundedRect(x, y, w, h int, r float32) bool {
	if x < 0 || y < 0 || x >= w || y >= h {
		return false
	}
	radiusInt := int(r)
	dx := float32(0)
	dy := float32(0)

	// Esquina superior izquierda
	if x < radiusInt && y < radiusInt {
		dx = float32(radiusInt - x)
		dy = float32(radiusInt - y)
	} else if x >= w-radiusInt && y < radiusInt {
		// Esquina superior derecha
		dx = float32(x - (w - radiusInt))
		dy = float32(radiusInt - y)
	} else if x < radiusInt && y >= h-radiusInt {
		// Esquina inferior izquierda
		dx = float32(radiusInt - x)
		dy = float32(y - (h - radiusInt))
	} else if x >= w-radiusInt && y >= h-radiusInt {
		// Esquina inferior derecha
		dx = float32(x - (w - radiusInt))
		dy = float32(y - (h - radiusInt))
	}

	if dx > 0 || dy > 0 {
		dist := dx*dx + dy*dy
		if dist > r*r {
			return false
		}
	}
	return true
}

type roundedRectRenderer struct {
	raster *canvas.Raster
	rect   *RoundedRect
}

func (r *roundedRectRenderer) Layout(size fyne.Size) {
	r.raster.Resize(size)
}

func (r *roundedRectRenderer) MinSize() fyne.Size {
	return r.rect.MinSize()
}

func (r *roundedRectRenderer) Refresh() {
	r.raster.Refresh()
}

func (r *roundedRectRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.raster}
}

func (r *roundedRectRenderer) Destroy() {}

func (t *Toast) animateProgress(onComplete func()) {
	steps := 60
	stepDuration := t.duration / time.Duration(steps)
	
	for i := steps; i >= 0; i-- {
		time.Sleep(stepDuration)
		
		progressRatio := float32(i) / float32(steps)

		fyne.Do(func() {
			if t.progress != nil {
				newWidth := progressRatio * t.barMaxWidth
				t.progress.Resize(fyne.NewSize(newWidth, 3))
				t.progress.Refresh()
			}
		})
	}
	
	fyne.Do(onComplete)
}

func (tm *ToastManager) dismissCurrent() {
	tm.currentToast = nil
	tm.toastStack.Hide()
	tm.toastStack.Refresh()
}

func (tm *ToastManager) Overlay() *fyne.Container {
	return tm.toastStack
}