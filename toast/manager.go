package toast

import (
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
	
	// Fondo con esquinas redondeadas (radio 8px)
	bgRect := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	bgRect.SetMinSize(fyne.NewSize(400, 50))
	
	// Stack: fondo + contenido
	toastCore := container.NewStack(bgRect, innerContent)
	
	// BORDE 1px con color separator - usando contenedor con padding mínimo
	// Para 1px usamos padding de 1px en NewPadded, pero NewPadded usa theme.Padding() que es más grande
	// Así que creamos nuestro propio contenedor con padding manual
	borderContainer := container.NewStack(
		// Fondo del borde (separator color)
		canvas.NewRectangle(theme.Color(theme.ColorNameSeparator)),
		// Contenido centrado con 1px de margen
		container.NewCenter(container.NewPadded(toastCore)),
	)
	
	// El padding del NewPadded interno crea el espacio para el borde
	// Pero para control exacto de 1px, usamos un contenedor personalizado
	borderContainer.Resize(fyne.NewSize(402, 52)) // 400+2, 50+2 = 1px cada lado
	
	outer := container.NewVBox(
		container.NewHBox(
			widget.NewLabel(""),
			borderContainer,
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