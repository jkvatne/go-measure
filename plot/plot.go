package plot

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/goki/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/widget"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
)

// Frame is a scope display in a frame
type Frame struct {
	widget.BaseWidget
	ScopeImg *image.RGBA
	face     font.Face
	Data     [][]float64
	n        int
}

type frameRender struct {
	frame   *Frame
	scope   *canvas.Image
	n       int
	objects []fyne.CanvasObject
}

// MinSize() is minimum sized
func (r *frameRender) MinSize() fyne.Size {
	return fyne.NewSize(640, 480)
}

func (r *frameRender) doPlot() {
	if img, ok := r.scope.Image.(draw.Image); ok {
		draw.Draw(img, img.Bounds(), image.NewUniform(colornames.Cyan), image.Pt(0, 0), draw.Src)
		plot(img, r.frame.Data)
		r.n++
		Label(img, 35, h10+2, fmt.Sprintf("n=%d", r.n), colornames.Orange, Regular12)
	}
}

// Layout resizes image
func (r *frameRender) Layout(size fyne.Size) {
	r.scope.Image = image.NewRGBA(image.Rect(0, 0, size.Width, size.Height))
	r.doPlot()
	r.scope.Resize(size)
}

// ApplyTheme is not used
func (r *frameRender) ApplyTheme() {
}

// BackgroundColor is background
func (r *frameRender) BackgroundColor() color.Color {
	return colornames.Green
}

// Refresh do refresh
func (r *frameRender) Refresh() {
	r.doPlot()
	canvas.Refresh(r.frame)
}

// Object returns objects
func (r *frameRender) Objects() []fyne.CanvasObject {
	return r.objects
}

// Destroy is not used
func (r *frameRender) Destroy() {
}

// MinSize is the minimum size
func (f *Frame) MinSize() fyne.Size {
	return fyne.NewSize(640, 480)
}

// SetData will update the plot source data
func (f *Frame) SetData(data [][]float64) {
	f.Data = data
}

// CreateRenderer gets the widget renderer for this table - internal use only
func (f *Frame) CreateRenderer() fyne.WidgetRenderer {
	r := &frameRender{}
	r.frame = f
	r.scope = &canvas.Image{}
	r.objects = []fyne.CanvasObject{r.scope}
	r.Refresh()
	return r
}

// NewFrame makes a scope frame
func NewFrame() *Frame {
	f := &Frame{}
	f.ExtendBaseWidget(f)
	fnt, _ := truetype.Parse(goregular.TTF)
	f.face = truetype.NewFace(fnt, &truetype.Options{Size: 16, DPI: 120})
	return f
}
