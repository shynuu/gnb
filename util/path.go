package util

import (
	"free5gc/lib/path_util"
)

var RanLogPath = path_util.Gofree5gcPath("free5gc/ransslkey.log")
var RanPemPath = path_util.Gofree5gcPath("free5gc/support/TLS/ran.pem")
var RanKeyPath = path_util.Gofree5gcPath("free5gc/support/TLS/ran.key")
var DefaultRanConfigPath = path_util.Gofree5gcPath("free5gc/config/gnbcfg.conf")
