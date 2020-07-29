package main

import (
	"free5gc/src/app"
	"free5gc/src/gnb/logger"
	"free5gc/src/gnb/service"
	"free5gc/src/gnb/version"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var RAN = &service.RAN{}

var appLog *logrus.Entry

func init() {
	appLog = logger.AppLog
}

func main() {
	app := cli.NewApp()
	app.Name = "5G-ran"
	appLog.Infoln(app.Name)
	appLog.Infoln("RAN version: ", version.GetVersion())
	app.Usage = "-free5gccfg common configuration file -gnbcfg ran configuration file"
	app.Action = action
	app.Flags = RAN.GetCliCmd()
	if err := app.Run(os.Args); err != nil {
		logger.AppLog.Errorf("RAN Run error: %v", err)
	}
}

func action(c *cli.Context) {
	app.AppInitializeWillInitialize(c.String("free5gccfg"))
	RAN.Initialize(c)
	RAN.Start()
}
