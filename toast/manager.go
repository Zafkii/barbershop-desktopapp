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
	window      fyne.Window
	toastStack  *fyne.Container
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
	// Descartar toast actual si existe
	if tm.currentToast != nil {
		tm.dismissCurrent()
	}
	
	toast := tm.createToast(message, toastType, duration)
	tm.currentToast = toast
	
	// Mostrar
	tm.toastStack.RemoveAll()
	tm.toastStack.Add(toast.container)
	tm.toastStack.Show()
	tm.toastStack.Refresh()
	
	// Animar barra de progreso
	go toast.animateProgress(func() {
		fyne.Do(func() {
			tm.dismissCurrent()
		})
	})
}

func (tm *ToastManager) createToast(message string, toastType ToastType, duration time.Duration) *Toast {
	var bgColor, barColor fyne.ThemeColorName
	switch toastType {
	case ToastSuccess:
		bgColor = theme.ColorNameSuccess
		barColor = theme.ColorNameSuccess
	case ToastWarning:
		bgColor = theme.ColorNameWarning
		barColor = theme.ColorNameWarning
	case ToastError:
		bgColor = theme.ColorNameError
		barColor = theme.ColorNameError
	default:
		bgColor = theme.ColorNamePrimary
		barColor = theme.ColorNamePrimary
	}
	
	label := widget.NewLabel(message)
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.Wrapping = fyne.TextWrapWord
	
	bgRect := canvas.NewRectangle(theme.Color(bgColor))
	bgRect.SetMinSize(fyne.NewSize(400, 50))
	
	progress := canvas.NewRectangle(theme.Color(barColor))
	progress.SetMinSize(fyne.NewSize(400, 3))
	
	content := container.NewBorder(nil, progress, nil, nil, container.NewPadded(label))
	toastContainer := container.NewStack(bgRect, content)
	
	outer := container.NewVBox(
		container.NewHBox(
			widget.NewLabel(""),
			container.NewPadded(toastContainer),
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

// Overlay devuelve el contenedor para superponer en la UI principal
func (tm *ToastManager) Overlay() *fyne.Container {
	return tm.toastStack
}