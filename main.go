package main

import (
	"github.com/windrainw/bs_common/conf"
	"github.com/windrainw/common-ares/frame/config"
	"github.com/windrainw/common-ares/frame/daemon"
	"github.com/windrainw/common-ares/frame/daemon/bootstrap"
	"github.com/windrainw/common-ares/frame/logs"
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
	conf.InitDBConnect()
}
