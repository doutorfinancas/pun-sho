package request

type QRCode struct {
	Create       bool   `json:"create"`
	Width        int    `json:"width"`
	BorderWidth  int    `json:"border_width"`
	FgColor      string `json:"foreground_color"`
	BgColor      string `json:"background_color"`
	Shape        string `json:"shape"`
	LogoImage    string `json:"logo"`
	OutputFormat string `json:"output_format" example:"svg" default:"png"`
}
