package main

import (
	"os"
	"time"

	"github.com/Netflix/go-env"
	"github.com/doutorfinancas/pun-sho/api"
	"github.com/subosito/gotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const Timestamp = "timestamp"

func main() {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.EncoderConfig.TimeKey = Timestamp
	loggerConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	log, _ := loggerConfig.Build()
	cfg := &api.Config{}
	handleEnv(log, cfg)

	a := api.NewAPI(log, cfg)

	a.Run()
}

func handleEnv(log *zap.Logger, cfg *api.Config) {
	if _, err := os.Stat(".env"); err == nil {
		err := gotenv.Load(".env")
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	if _, err := env.UnmarshalFromEnviron(cfg); err != nil {
		log.Fatal(err.Error())
	}
}
