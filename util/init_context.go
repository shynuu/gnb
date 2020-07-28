package util

import (
	"free5gc/lib/MongoDBLibrary"
	"free5gc/lib/openapi/models"
	"free5gc/src/ran/context"
	"free5gc/src/ran/factory"
	"free5gc/src/ran/logger"

	"github.com/google/uuid"
)

func InitRanContext(context *context.RANContext) {
	MongoDBLibrary.SetMongoDB("free5gc", "mongodb://127.0.0.1:27017")
	config := factory.RanConfig
	logger.UtilLog.Infof("ranconfig Info: Version[%s] Description[%s]", config.Info.Version, config.Info.Description)
	configuration := config.Configuration
	context.NfId = uuid.New().String()
	if configuration.RanName != "" {
		context.Name = configuration.RanName
	}
	sbi := configuration.Sbi
	context.UriScheme = models.UriScheme(sbi.Scheme)
	context.HttpIPv4Address = "127.0.0.1" // default localhost
	context.HttpIpv4Port = 32000          // default port
	if sbi != nil {
		if sbi.IPv4Addr != "" {
			context.HttpIPv4Address = sbi.IPv4Addr
		}
		if sbi.Port != 0 {
			context.HttpIpv4Port = sbi.Port
		}
	}

	// for i := range context.SupportTaiLists {
	// 	context.SupportTaiLists[i].Tac = TACConfigToModels(context.SupportTaiLists[i].Tac)
	// }
	context.NetworkName = configuration.NetworkName
	context.AmfInterface = configuration.AmfInterface
	context.UpfInterface = configuration.UpfInterface
	context.UEList = configuration.UEList
}
