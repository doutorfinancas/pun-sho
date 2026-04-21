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
		logoURI := detectLogoDataURI(w.o.Logo)
		w.c.Image(
			wd,
			ht,
			width*(moduleSize)/5,
			height*(moduleSize)/5,
			logoURI)
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

// detectLogoDataURI determines the correct data URI prefix for a base64-encoded logo.
// It decodes the first few bytes to check for PNG/SVG signatures.
func detectLogoDataURI(b64Logo string) string {
	// If it already starts with "data:", it's a full data URI
	if len(b64Logo) > 5 && b64Logo[:5] == "data:" {
		return b64Logo
	}

	// Try to detect format from base64 content
	// PNG starts with \x89PNG, base64 of which starts with "iVBOR"
	// SVG starts with "<" or "<?xml", base64 of which starts with "PD" or "PHN2"
	if len(b64Logo) >= 5 {
		if b64Logo[:5] == "iVBOR" {
			return "data:image/png;base64," + b64Logo
		}
		if b64Logo[:2] == "PD" || b64Logo[:4] == "PHN2" {
			return "data:image/svg+xml;base64," + b64Logo
		}
	}

	// Default to PNG for backwards compatibility
	return "data:image/png;base64," + b64Logo
}

type Options struct {
	BGColor     string
	FGColor     string
	Width       int
	BorderWidth int
	Shape       string
	Logo        string
}
