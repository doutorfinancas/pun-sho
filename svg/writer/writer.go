package writer

import (
	"errors"
	"fmt"
	"io"
	"os"

	svg "github.com/ajstarks/svgo"
	"github.com/doutorfinancas/pun-sho/str"
	"github.com/yeqown/go-qrcode/v2"
)

const (
	finderSize        = 7
	alignmentSize     = 5
	shapeCircle       = "circle"
	shapeCircleProper = "circle-proper"
	shapeRect         = "rect"
)

type CustomSVGWriter struct {
	w io.WriteCloser

	o *Options

	c *svg.SVG
}

func NewWithWriter(writer io.WriteCloser, opt *Options) qrcode.Writer {
	return &CustomSVGWriter{w: writer, c: svg.New(writer), o: opt}
}

func (w CustomSVGWriter) Write(matrix qrcode.Matrix) error {
	moduleSize := 10

	width, height := matrix.Width(), matrix.Height()
	style := ""

	if w.o.BGColor != "none" {
		style = "background-color:" + w.o.BGColor
	}

	w.c.Start(
		width*moduleSize+moduleSize,
		height*moduleSize+moduleSize,
		fmt.Sprintf("style=\"%s\"", style))

	fgFill := "fill:" + w.o.FGColor

	switch w.o.Shape {
	case shapeCircleProper:
		w.drawCircleProperFinder(width, height, moduleSize)

		matrix.Iterate(
			qrcode.IterDirection_COLUMN,
			func(x int, y int, v qrcode.QRValue) {
				if v.IsSet() && w.shouldDraw(x, y, width, height) {
					w.c.Circle(x*moduleSize+moduleSize, y*moduleSize+moduleSize, moduleSize/2, fgFill)
				}
			},
		)
	case shapeCircle:
		matrix.Iterate(
			qrcode.IterDirection_COLUMN,
			func(x int, y int, v qrcode.QRValue) {
				if v.IsSet() {
					w.c.Circle(x*moduleSize+moduleSize, y*moduleSize+moduleSize, moduleSize/2, fgFill)
				}
			},
		)
	default:
		matrix.Iterate(
			qrcode.IterDirection_COLUMN,
			func(x int, y int, v qrcode.QRValue) {
				if v.IsSet() {
					w.c.Rect(
						x*moduleSize+moduleSize,
						y*moduleSize+moduleSize,
						moduleSize,
						moduleSize,
						fgFill)
				}
			},
		)
	}

	if w.o.Logo != "" {
		wd := (((width + 1) * moduleSize) - (width * (moduleSize) / 5)) / 2
		ht := (((height + 1) * moduleSize) - (height * (moduleSize) / 5)) / 2
		w.c.Image(
			wd,
			ht,
			width*(moduleSize)/5,
			height*(moduleSize)/5,
			"data:image/svg+xml;base64,"+w.o.Logo)
	}

	return nil
}

func (w CustomSVGWriter) Close() error {
	w.c.End()

	if w.w == nil {
		return nil
	}

	if err := w.w.Close(); !errors.Is(err, os.ErrClosed) {
		return err
	}

	return nil
}

func (w CustomSVGWriter) drawCircleProperFinder(width, height, moduleSize int) {
	w.c.Circle(
		(finderSize+1)*moduleSize/2,
		(finderSize+1)*moduleSize/2,
		(finderSize-1)*moduleSize/2,
		fmt.Sprintf("fill:none;stroke:%s;stroke-width:%s", w.o.FGColor, str.ToString(moduleSize)))
	w.c.Circle(
		(finderSize+1)*moduleSize/2,
		(finderSize+1)*moduleSize/2,
		(finderSize-1)*moduleSize/4,
		fmt.Sprintf("fill:%s", w.o.FGColor))
	w.c.Circle(
		(finderSize+1)*moduleSize/2,
		height*moduleSize-((finderSize-1)*moduleSize/2),
		(finderSize-1)*moduleSize/2,
		fmt.Sprintf("fill:none;stroke:%s;stroke-width:%s", w.o.FGColor, str.ToString(moduleSize)))
	w.c.Circle(
		(finderSize+1)*moduleSize/2,
		height*moduleSize-((finderSize-1)*moduleSize/2),
		(finderSize-1)*moduleSize/4,
		fmt.Sprintf("fill:%s", w.o.FGColor))
	w.c.Circle(
		width*moduleSize-((finderSize-1)*moduleSize/2),
		(finderSize+1)*moduleSize/2,
		(finderSize-1)*moduleSize/2,
		fmt.Sprintf("fill:none;stroke:%s;stroke-width:%s", w.o.FGColor, str.ToString(moduleSize)))
	w.c.Circle(
		width*moduleSize-((finderSize-1)*moduleSize/2),
		(finderSize+1)*moduleSize/2,
		(finderSize-1)*moduleSize/4,
		fmt.Sprintf("fill:%s", w.o.FGColor))
}

func (w CustomSVGWriter) shouldDraw(x, y, width, height int) bool {
	if x < finderSize && y < finderSize {
		return false
	}

	if x < finderSize && y > height-finderSize-1 {
		return false
	}

	if x > width-finderSize-1 && y < finderSize {
		return false
	}

	return true
}

type Options struct {
	BGColor     string
	FGColor     string
	Width       int
	BorderWidth int
	Shape       string
	Logo        string
}
