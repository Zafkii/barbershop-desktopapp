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
	message   string
	toastType ToastType
	duration  time.Duration
	progress  *canvas.Rectangle
	container *fyne.Container
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
	
	progress := canvas.NewRectangle(theme.Color(barColor))
	progress.SetMinSize(fyne.NewSize(400, 3))
	
	innerContent := container.NewBorder(nil, progress, nil, nil, container.NewPadded(label))
	
	// Usar RoundedRect para fondo con esquinas redondeadas (radio 8px)
	bgRect := NewRoundedRect(theme.Color(theme.ColorNameBackground), 8)
	bgRect.SetMinSize(fyne.NewSize(400, 50))
	
	// Stack: fondo redondeado + contenido
	toastCore := container.NewStack(bgRect, innerContent)
	
	// BORDE 1px redondeado: otro RoundedRect ligeramente más grande detrás
	borderRect := NewRoundedRect(theme.Color(theme.ColorNameSeparator), 9) // radio 9 = 8+1 para que se vea el borde
	borderRect.SetMinSize(fyne.NewSize(402, 52)) // 1px cada lado
	
	borderedToast := container.NewStack(borderRect, toastCore)
	
	outer := container.NewVBox(
		container.NewHBox(
			widget.NewLabel(""),
			borderedToast,
			widget.NewLabel(""),
		),
	)
	
	return &Toast{
		message:   message,
		toastType: toastType,
		duration:  duration,
		progress:  progress,
		container: outer,
	}
}

// RoundedRect es un widget que dibuja un rectángulo con esquinas redondeadas usando canvas.Raster
type RoundedRect struct {
	widget.BaseWidget
	fillColor  color.Color
	radius     float32
	minSize    fyne.Size
}

func NewRoundedRect(c color.Color, radius float32) *RoundedRect {
	r := &RoundedRect{
		fillColor: c,
		radius:    radius,
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
	radius := float32(r.radius)
	
	// Convertir color a RGBA
	rgba := color.RGBAModel.Convert(r.fillColor).(color.RGBA)
	
	radiusInt := int(radius)
	
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// Calcular distancia a la esquina más cercana
			dx := float32(0)
			dy := float32(0)
			
			// Esquina superior izquierda
			if x < radiusInt && y < radiusInt {
				dx = float32(radiusInt - x)
				dy = float32(radiusInt - y)
			}
			// Esquina superior derecha
			if x >= w-radiusInt && y < radiusInt {
				dx = float32(x - (w - radiusInt))
				dy = float32(radiusInt - y)
			}
			// Esquina inferior izquierda
			if x < radiusInt && y >= h-radiusInt {
				dx = float32(radiusInt - x)
				dy = float32(y - (h - radiusInt))
			}
			// Esquina inferior derecha
			if x >= w-radiusInt && y >= h-radiusInt {
				dx = float32(x - (w - radiusInt))
				dy = float32(y - (h - radiusInt))
			}
			
			// Si está en zona de esquinas, verificar distancia
			if dx > 0 || dy > 0 {
				dist := dx*dx + dy*dy
				if dist > radius*radius {
					continue // fuera del radio, pixel transparente
				}
			}
			
			img.SetRGBA(x, y, rgba)
		}
	}
	
	return img
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
				newWidth := progressRatio * 400
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