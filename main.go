package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"omo-msa-activity/config"
	"omo-msa-activity/handler"
	"omo-msa-activity/model"
	"omo-msa-activity/subscriber"
	"os"
	"path/filepath"
	"time"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/logger"
	proto "github.com/xtech-cloud/omo-msp-activity/proto/activity"
)

func main() {
	config.Setup()
	model.Setup()
	model.AutoMigrateDatabase()
	var err error
	model.DefaultConn, err = model.OpenSqlDB()
	if nil != err {
		panic(err)
	}
	defer model.CloseSqlDB(model.DefaultConn)

	// New Service
	service := micro.NewService(
		micro.Name("omo.msa.activity"),
		micro.Version(BuildVersion),
		micro.RegisterTTL(time.Second*time.Duration(config.Schema.Service.TTL)),
		micro.RegisterInterval(time.Second*time.Duration(config.Schema.Service.Interval)),
		micro.Address(config.Schema.Service.Address),
	)

	// Initialise service
	service.Init()

	// Register subscriber
	subscriber.Setup(service.Server())

	// Register Handler
	proto.RegisterChannelHandler(service.Server(), new(handler.Channel))
	proto.RegisterRecordHandler(service.Server(), new(handler.Record))

	app, _ := filepath.Abs(os.Args[0])

	logger.Info("-------------------------------------------------------------")
	logger.Info("- Micro Service Agent -> Run")
	logger.Info("-------------------------------------------------------------")
	logger.Infof("- version      : %s", BuildVersion)
	logger.Infof("- application  : %s", app)
	logger.Infof("- md5          : %s", md5hex(app))
	logger.Infof("- build        : %s", BuildTime)
	logger.Infof("- commit       : %s", CommitID)
	logger.Info("-------------------------------------------------------------")
	// Run service
	if err := service.Run(); err != nil {
		logger.Error(err)
	}
}

func md5hex(_file string) string {
	h := md5.New()

	f, err := os.Open(_file)
	if err != nil {
		return ""
	}
	defer f.Close()

	io.Copy(h, f)

	return hex.EncodeToString(h.Sum(nil))
}
