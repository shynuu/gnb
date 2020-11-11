package service

import (
	"bufio"
	"fmt"
	"free5gc/lib/http2_util"
	"free5gc/lib/path_util"
	"free5gc/src/app"
	"free5gc/src/gnb/context"
	"free5gc/src/gnb/factory"
	"free5gc/src/gnb/httpservice"
	"free5gc/src/gnb/logger"
	"free5gc/src/gnb/util"
	"os/exec"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type RAN struct{}

type (
	// Config information.
	Config struct {
		gnbcfg string
	}
)

var config Config

var ranCLi = []cli.Flag{
	cli.StringFlag{
		Name:  "free5gccfg",
		Usage: "common config file",
	},
	cli.StringFlag{
		Name:  "gnbcfg",
		Usage: "amf config file",
	},
}

var initLog *logrus.Entry

func init() {
	initLog = logger.InitLog
}

func (*RAN) GetCliCmd() (flags []cli.Flag) {
	return ranCLi
}

func (*RAN) Initialize(c *cli.Context) {

	config = Config{
		gnbcfg: c.String("gnbcfg"),
	}

	if config.gnbcfg != "" {
		factory.InitConfigFactory(config.gnbcfg)
	} else {
		DefaultRanConfigPath := path_util.Gofree5gcPath("free5gc/config/gnbcfg.conf")
		factory.InitConfigFactory(DefaultRanConfigPath)
	}

	if app.ContextSelf().Logger.RAN.DebugLevel != "" {
		level, err := logrus.ParseLevel(app.ContextSelf().Logger.RAN.DebugLevel)
		if err != nil {
			initLog.Warnf("Log level [%s] is not valid, set to [info] level", app.ContextSelf().Logger.RAN.DebugLevel)
			logger.SetLogLevel(logrus.InfoLevel)
		} else {
			logger.SetLogLevel(level)
			initLog.Infof("Log level is set to [%s] level", level)
		}
	} else {
		initLog.Infoln("Log level is default set to [info] level")
		logger.SetLogLevel(logrus.InfoLevel)
	}

	logger.SetReportCaller(app.ContextSelf().Logger.RAN.ReportCaller)

}

func (ran *RAN) FilterCli(c *cli.Context) (args []string) {
	for _, flag := range ran.GetCliCmd() {
		name := flag.GetName()
		value := fmt.Sprint(c.Generic(name))
		if value == "" {
			continue
		}

		args = append(args, "--"+name, value)
	}
	return args
}

func (ran *RAN) Start() {
	initLog.Infoln("Server started")

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "User-Agent", "Referrer", "Host", "Token", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowAllOrigins:  true,
		MaxAge:           86400,
	}))

	httpservice.AddService(router)

	self := context.RAN_Self()
	util.InitRanContext(self)

	addr := fmt.Sprintf("%s:%d", self.HttpIPv4Address, self.HttpIpv4Port)

	server, err := http2_util.NewServer(addr, util.RanLogPath, router)

	if server == nil {
		initLog.Errorln("Initialize HTTP server failed: %+v", err)
		return
	}

	if err != nil {
		initLog.Warnln("Initialize HTTP server: +%v", err)
	}

	serverScheme := factory.RanConfig.Configuration.Sbi.Scheme
	if serverScheme == "http" {
		err = server.ListenAndServe()
	} else if serverScheme == "https" {
		err = server.ListenAndServeTLS(util.RanPemPath, util.RanKeyPath)
	}

	if err != nil {
		initLog.Fatalln("HTTP server setup failed: %+v", err)
	}
}

func (amf *RAN) Exec(c *cli.Context) error {

	//RAN.Initialize(cfgPath, c)

	initLog.Traceln("args:", c.String("gnbcfg"))
	args := amf.FilterCli(c)
	initLog.Traceln("filter: ", args)
	command := exec.Command("./ran", args...)

	stdout, err := command.StdoutPipe()
	if err != nil {
		initLog.Fatalln(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		in := bufio.NewScanner(stdout)
		for in.Scan() {
			fmt.Println(in.Text())
		}
		wg.Done()
	}()

	stderr, err := command.StderrPipe()
	if err != nil {
		initLog.Fatalln(err)
	}
	go func() {
		in := bufio.NewScanner(stderr)
		for in.Scan() {
			fmt.Println(in.Text())
		}
		wg.Done()
	}()

	go func() {
		if err := command.Start(); err != nil {
			initLog.Errorf("RAN Start error: %v", err)
		}
		wg.Done()
	}()

	wg.Wait()

	return err
}
