package main

import (
	"bytes"
	"image"

	"github.com/fyne-io/oksvg"
	"github.com/srwiley/rasterx"
)

// rasterizeBackground convierte el SVG (ya colorizado) en un bitmap image.Image.
// Se ejecuta fuera del hilo principal para no bloquear la UI.
func rasterizeBackground(svgData []byte, w, h int) image.Image {
	if w <= 0 || h <= 0 {
		return nil
	}
	icon, err := oksvg.ReadIconStream(bytes.NewReader(svgData))
	if err != nil {
		return nil
	}
	icon.SetTarget(0, 0, float64(w), float64(h))

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())
	dasher := rasterx.NewDasher(w, h, scanner)
	icon.Draw(dasher, 1.0)
	return img
}