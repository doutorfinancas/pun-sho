package service

import (
	"bytes"
	"encoding/base64"
	"image/png"

	"github.com/doutorfinancas/pun-sho/api/request"
	"github.com/doutorfinancas/pun-sho/buf"
	"github.com/doutorfinancas/pun-sho/str"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
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

	bgColor := standard.WithBgColorRGBHex("#ffffff")
	fgColor := standard.WithFgColorRGBHex("#000000")

	if str.SubString(qr.BgColor, 0, 1) == "#" {
		bgColor = standard.WithBgColorRGBHex(qr.BgColor)
	}

	if qr.BgColor == TransparentBackground {
		bgColor = standard.WithBgTransparent()
	}

	if str.SubString(qr.FgColor, 0, 1) == "#" {
		fgColor = standard.WithFgColorRGBHex(qr.FgColor)
	}

	options := []standard.ImageOption{
		bgColor,
		fgColor,
		standard.WithBuiltinImageEncoder(standard.PNG_FORMAT),
	}

	if qr.LogoImage != "" {
		x, err := base64.StdEncoding.DecodeString(qr.LogoImage)
		if err != nil {
			return "", err
		}

		img, err := png.Decode(bytes.NewReader(x))
		if err != nil {
			return "", err
		}

		options = append(options, standard.WithLogoImage(img))
	} else if q.logo != "" {
		options = append(options, standard.WithLogoImageFilePNG(q.logo))
	}

	if qr.Width > 0 {
		options = append(options, standard.WithQRWidth(uint8(qr.Width)))
	}

	if qr.BorderWidth > 0 {
		options = append(options, standard.WithBorderWidth(qr.BorderWidth))
	}

	if qr.Shape == "circle" {
		options = append(options, standard.WithCircleShape())
	}

	var b []byte
	x := bytes.NewBuffer(b)
	w := buf.NewWriteCloser(x)
	wr := standard.NewWithWriter(w, options...)

	err = qrc.Save(wr)
	if err != nil {
		return "", err
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(x.Bytes()), nil
}
