package request

type GeneratePreview struct {
	Link   string  `json:"link"`
	QRCode *QRCode `json:"qr_code"`
}
