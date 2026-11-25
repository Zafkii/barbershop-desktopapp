package ui

import (
	"fmt"
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
	circle.StrokeWidth = 2
	circle.StrokeColor = color.RGBA{R: 80, G: 80, B: 80, A: 255}

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
	return fyne.NewSize(18, 18)
}

func (r *connectivityRenderer) Refresh() {
	fmt.Printf("[renderer] Refresh, status=%v\n", r.ind.status)
	r.circle.FillColor = colorForStatus(r.ind.status)
	r.circle.Refresh()
	fmt.Printf("[renderer] Refresh done, color=%v\n", r.circle.FillColor)
}

func (r *connectivityRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.circle}
}

func (r *connectivityRenderer) Destroy() {}

func (ci *ConnectivityIndicator) UpdateStatus(s connectivity.Status) {
	fmt.Printf("[indicator] UpdateStatus, s=%v, current=%v\n", s, ci.status)
	ci.status = s
	ci.Refresh()
	fmt.Printf("[indicator] UpdateStatus done\n")
}

func (ci *ConnectivityIndicator) Status() connectivity.Status {
	return ci.status
}