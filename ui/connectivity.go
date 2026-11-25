package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	"barbershop-desktop/connectivity"
)

// ConnectivityIndicator muestra el estado de conexión con un círculo coloreado
type ConnectivityIndicator struct {
	widget.BaseWidget
	status connectivity.Status
	circle *canvas.Circle
}

func NewConnectivityIndicator() *ConnectivityIndicator {
	ci := &ConnectivityIndicator{
		status: connectivity.StatusOffline,
	}
	ci.ExtendBaseWidget(ci)
	return ci
}

func (ci *ConnectivityIndicator) CreateRenderer() fyne.WidgetRenderer {
	circle := canvas.NewCircle(colorForStatus(ci.status))
	circle.StrokeWidth = 0
	
	r := &connectivityRenderer{
		circle: circle,
		ind:    ci,
	}
	return r
}

func colorForStatus(s connectivity.Status) color.Color {
	switch s {
	case connectivity.StatusOnline:
		return color.RGBA{R: 0, G: 200, B: 0, A: 255} // Verde
	case connectivity.StatusServerDown:
		return color.RGBA{R: 255, G: 200, B: 0, A: 255} // Amarillo
	default:
		return color.RGBA{R: 200, G: 0, B: 0, A: 255} // Rojo
	}
}

type connectivityRenderer struct {
	circle *canvas.Circle
	ind    *ConnectivityIndicator
}

func (r *connectivityRenderer) Layout(size fyne.Size) {
	r.circle.Resize(size)
	r.circle.Move(fyne.NewPos(0, 0))
}

func (r *connectivityRenderer) MinSize() fyne.Size {
	return fyne.NewSize(20, 20)
}

func (r *connectivityRenderer) Refresh() {
	r.circle.FillColor = colorForStatus(r.ind.status)
	r.circle.Refresh()
}

func (r *connectivityRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.circle}
}

func (r *connectivityRenderer) Destroy() {}

// UpdateStatus actualiza el estado y refresca el widget
func (ci *ConnectivityIndicator) UpdateStatus(s connectivity.Status) {
	ci.status = s
	ci.Refresh()
}