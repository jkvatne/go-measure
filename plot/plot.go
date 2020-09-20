package plot

import (
	"fmt"
	"image"
	"time"

	"golang.org/x/image/colornames"

	"golang.org/x/image/font/gofont/goregular"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"github.com/goki/freetype/truetype"
	"golang.org/x/image/font"
)

type Frame struct {
	Canvas   fyne.CanvasObject
	ScopeImg *image.RGBA
	face     font.Face
	Data     [][]float64
	n        int
}

func (f *Frame) Redraw(window fyne.Window) {
	f.n++
	f.ScopeImg = image.NewRGBA(image.Rect(0, 0, f.Canvas.Size().Width, f.Canvas.Size().Height))
	plot(f.ScopeImg, f.Data)
	Label(f.ScopeImg, 100, 100, fmt.Sprintf("n=%d", f.n), colornames.Orange, Regular10)
	window.Canvas().Refresh(f.Canvas)

}

func (f *Frame) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if size.Width != f.Canvas.Size().Width || size.Height != f.Canvas.Size().Height {
		f.Canvas.Resize(size)
		f.n++
		f.ScopeImg = image.NewRGBA(image.Rect(0, 0, f.Canvas.Size().Width, f.Canvas.Size().Height))
		plot(f.ScopeImg, f.Data)
		Label(f.ScopeImg, 100, 100, fmt.Sprintf("n=%d", f.n), colornames.Orange, Regular10)
		f.Canvas.Refresh() //events <- evRedraw
	}
}

func (f *Frame) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(640, 480)
}

func (f *Frame) SetData(data [][]float64) {
	f.Data = data
}

func NewFrame() *Frame {
	fnt, err := truetype.Parse(goregular.TTF)
	f := &Frame{}
	time.Sleep(300 * time.Millisecond)
	f.ScopeImg = image.NewRGBA(image.Rect(0, 0, 640, 480))
	//draw.Draw(scopeFrame.scopeImg, scopeFrame.scopeImg.Bounds(), image.NewUniform(colornames.Black), image.Pt(0, 0), draw.Src)
	f.Canvas = canvas.NewRaster(func(w, h int) image.Image { return f.ScopeImg })
	f.face = truetype.NewFace(fnt, &truetype.Options{Size: 10, DPI: 96})
	if err != nil {
		return nil
	}
	return f
}
