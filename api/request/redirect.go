package request

type Redirect struct {
	UserAgent string
	IP        string
	Meta      map[string][]string
	Extra     string
}
