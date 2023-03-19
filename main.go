package main

import (
	"os"
	"time"

	"github.com/Netflix/go-env"
	"github.com/subosito/gotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/doutorfinancas/pun-sho/api"
	"github.com/doutorfinancas/pun-sho/database"
	"github.com/doutorfinancas/pun-sho/entity"
	"github.com/doutorfinancas/pun-sho/service"
)

const Timestamp = "timestamp"

func main() {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.EncoderConfig.TimeKey = Timestamp
	loggerConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	log, _ := loggerConfig.Build()
	cfg := &api.Config{}
	handleEnv(log, cfg)
	g, err := database.Connect(cfg.GetDatabaseConfig())
	if err != nil {
		log.Fatal("can't connect to database")
	}
	db := database.NewDatabase(g)

	shortyRepo := entity.NewShortyRepository(db, log)
	shortyAccessRepo := entity.NewShortyAccessRepository(db, log)
	shortySvc := service.NewShortyService(cfg.HostName, log, shortyRepo, shortyAccessRepo)

	a := api.NewAPI(log, cfg, shortySvc)

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
