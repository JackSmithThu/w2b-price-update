package main

import (
	"github.com/JackSmithThu/bs_common/conf"
	"github.com/JackSmithThu/common-ares/frame/config"
	"github.com/JackSmithThu/common-ares/frame/daemon"
	"github.com/JackSmithThu/common-ares/frame/daemon/bootstrap"
	"github.com/JackSmithThu/common-ares/frame/logs"
)

var (
	DaemonsServiceCli *daemon.DaemonsService
)

func main() {
	config := bootstrap.Init()

	Init(config)

	DaemonsServiceCli = daemon.NewDaemonsService(config.ServiceName, HandleMessage)

	logs.Info("service_name %v is ready to run", config.ServiceName)

	DaemonsServiceCli.Run()
}

func Init(config *config.BaseConfig) {
	// conf.InitDBConnect()
	conf.InitPlatformDBConnect()
}
