package service

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"os"

	"github.com/doutorfinancas/pun-sho/api/request"
	"github.com/doutorfinancas/pun-sho/buf"
	"github.com/doutorfinancas/pun-sho/str"
	"github.com/doutorfinancas/pun-sho/svg/writer"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

const (
	finderSize        = 7
	alignmentSize     = 5
	shapeCircle       = "circle"
	shapeCircleProper = "circle-proper"
	shapeRect         = "rect"
)

type QRCodeService struct {
	logo string
}

func NewQRCodeService(logo string) *QRCodeService {
	return &QRCodeService{
		logo: logo,
	}
}

func (q *QRCodeService) Generate(qr *request.QRCode, link string) (string, error) {
	qrc, err := qrcode.New(link)
	if err != nil {
		return "", err
	}

	opt := &writer.Options{
		BGColor:     "#ffffff",
		FGColor:     "#000000",
		Width:       10,
		BorderWidth: 0,
		Shape:       shapeRect,
		Logo:        qr.LogoImage,
	}

	if str.SubString(qr.BgColor, 0, 1) == "#" {
		opt.BGColor = qr.BgColor
	}

	if str.SubString(qr.FgColor, 0, 1) == "#" {
		opt.FGColor = qr.FgColor
	}

	if qr.BgColor == TransparentBackground {
		opt.BGColor = "none"
	}

	if qr.Shape != "" {
		opt.Shape = qr.Shape
	}

	if qr.Width > 0 {
		opt.Width = qr.Width
	}

	if qr.BorderWidth > 0 {
		opt.BorderWidth = qr.BorderWidth
	}

	if qr.LogoImage == "" && q.logo != "" {
		x, err := os.ReadFile(q.logo)
		if err != nil {
			return "", err
		}
		opt.Logo = base64.StdEncoding.EncodeToString(x)
	}

	var b []byte
	x := bytes.NewBuffer(b)
	w := buf.NewWriteCloser(x)
	outputFormat := "data:image/png;base64,"
	var wr qrcode.Writer

	switch qr.OutputFormat {
	case "svg":
		wr = writer.NewWithWriter(w, opt)
		outputFormat = "data:image/svg+xml;base64,"
	default:
		bgColor := standard.WithBgColorRGBHex(opt.BGColor)

		if qr.BgColor == TransparentBackground {
			bgColor = standard.WithBgTransparent()
		}

		stdOpt := []standard.ImageOption{
			bgColor,
			standard.WithFgColorRGBHex(opt.FGColor),
			standard.WithQRWidth(uint8(opt.Width)),
			standard.WithBorderWidth(opt.BorderWidth),
			standard.WithBuiltinImageEncoder(standard.PNG_FORMAT),
		}

		if qr.Shape == shapeCircle || qr.Shape == shapeCircleProper {
			stdOpt = append(stdOpt, standard.WithCircleShape())
		}

		if opt.Logo != "" {
			a, err := base64.StdEncoding.DecodeString(opt.Logo)
			if err != nil {
				return "", err
			}

			img, err := png.Decode(bytes.NewReader(a))
			if err != nil {
				return "", err
			}

			stdOpt = append(stdOpt, standard.WithLogoImage(img))
		}

		wr = standard.NewWithWriter(w, stdOpt...)
	}

	err = qrc.Save(wr)
	if err != nil {
		return "", err
	}

	return outputFormat + base64.StdEncoding.EncodeToString(x.Bytes()), nil
}
