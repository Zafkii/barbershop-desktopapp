package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	"barbershop-desktop/connectivity"
)

// ConnectivityIndicator muestra el estado de conexión con un icono SVG
type ConnectivityIndicator struct {
	widget.BaseWidget
	status   connectivity.Status
	wifiSVG  []byte
	wifiOffSVG []byte
}

func NewConnectivityIndicator(wifiSVG, wifiOffSVG []byte) *ConnectivityIndicator {
	ci := &ConnectivityIndicator{
		status:     connectivity.StatusOffline,
		wifiSVG:    wifiSVG,
		wifiOffSVG: wifiOffSVG,
	}
	ci.ExtendBaseWidget(ci)
	return ci
}

func (ci *ConnectivityIndicator) CreateRenderer() fyne.WidgetRenderer {
	img := ci.createImage()
	r := &connectivityRenderer{
		img: img,
		ind: ci,
	}
	return r
}

func (ci *ConnectivityIndicator) createImage() *canvas.Image {
	var svgData []byte
	var color string
	var name string

	switch ci.status {
	case connectivity.StatusOnline:
		svgData = ci.wifiSVG
		color = "#00C800" // Verde
		name = "wifi-green.svg"
	case connectivity.StatusServerDown:
		svgData = ci.wifiSVG
		color = "#FFC800" // Amarillo
		name = "wifi-yellow.svg"
	default:
		svgData = ci.wifiOffSVG
		color = "#C80000" // Rojo
		name = "wifi-red.svg"
	}

	// Reemplazar el color del stroke en el SVG
	coloredSVG := strings.ReplaceAll(string(svgData), `stroke="#000000"`, fmt.Sprintf(`stroke="%s"`, color))

	// Usar un nombre único por color para evitar cache de Fyne
	img := canvas.NewImageFromResource(fyne.NewStaticResource(name, []byte(coloredSVG)))
	img.SetMinSize(fyne.NewSize(24, 24))
	img.FillMode = canvas.ImageFillContain
	return img
}

type connectivityRenderer struct {
	img *canvas.Image
	ind *ConnectivityIndicator
}

func (r *connectivityRenderer) Layout(size fyne.Size) {
	r.img.Resize(size)
	r.img.Move(fyne.NewPos(0, 0))
}

func (r *connectivityRenderer) MinSize() fyne.Size {
	return fyne.NewSize(24, 24)
}

func (r *connectivityRenderer) Refresh() {
	// Recrear la imagen con el color apropiado
	// Usar un nombre único para evitar el cache de SVG de Fyne
	newImg := r.ind.createImage()
	r.img.Resource = newImg.Resource
	r.img.Refresh()
}

func (r *connectivityRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.img}
}

func (r *connectivityRenderer) Destroy() {}

func (ci *ConnectivityIndicator) UpdateStatus(s connectivity.Status) {
	ci.status = s
	ci.Refresh()
}

func (ci *ConnectivityIndicator) Status() connectivity.Status {
	return ci.status
}