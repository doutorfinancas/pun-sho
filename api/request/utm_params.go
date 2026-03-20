package request

type UTMParams struct {
	Source   string `json:"utm_source" form:"utm_source"`
	Medium   string `json:"utm_medium" form:"utm_medium"`
	Campaign string `json:"utm_campaign" form:"utm_campaign"`
	Term     string `json:"utm_term" form:"utm_term"`
	Content  string `json:"utm_content" form:"utm_content"`
}

func (u *UTMParams) IsEmpty() bool {
	if u == nil {
		return true
	}
	return u.Source == "" && u.Medium == "" && u.Campaign == "" && u.Term == "" && u.Content == ""
}
