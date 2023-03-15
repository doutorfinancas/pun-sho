package api

type Config struct {
	Port        int    `env:"API_PORT"`
	Token       string `env:"AUTH_TOKEN"`
	UnknownPage string `env:"UNKNOWN_PAGE"`
}
