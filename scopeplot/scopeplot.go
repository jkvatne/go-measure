
	/*package scopeplot

import (
	"image"
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

type ScopePlot struct {
	widget.BaseWidget
	img  image.RGBA
	Icon fyne.Resource
}

type ScopePlotRenderer struct {
	img     *canvas.Image
	objects []fyne.CanvasObject
}

type scopeRenderer struct {
	widget.BaseRenderer
	image *Icon
}

func (i *scopeRenderer) MinSize() fyne.Size {
	size := theme.IconInlineSize()
	return fyne.NewSize(size, size)
}

func (i *scopeRenderer) Layout(size fyne.Size) {
	if len(i.Objects()) == 0 {
		return
	}

	i.Objects()[0].Resize(size)
}

func (i *scopeRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

func (i *scopeRenderer) Refresh() {
	i.image.propertyLock.RLock()
	i.updateObjects()
	i.image.propertyLock.RUnlock()
	i.Layout(i.image.Size())
	canvas.Refresh(i.image.super())
}

func (i *scopeRenderer) updateObjects() {
	var objects []fyne.CanvasObject
	objects = append(objects, i.img)
	i.SetObjects(objects)
}
*/