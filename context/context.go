package context

import (
	"free5gc/lib/openapi/models"
)

var ranContext = RANContext{}

func init() {
	RAN_Self().NetworkName.Full = "free5GC"
}

type RANContext struct {
	HttpIpv4Port    int
	HttpIPv4Address string
	HttpIPv6Address string
	NetworkName     NetworkName // unit is second
	AmfInterface    AmfInterface
	UpfInterface    UpfInterface
	UEList          []UE
	UriScheme       models.UriScheme
	NfId            string
	Name            string
}

type UE struct {
	Supi string `yaml:"SUPI,omitempty" json:"SUPI"`

	ipv4 string `yaml:"ipv4,omitempty" json:"SUPI"`

	identifier string `yaml:"identifier,omitempty" json:"SUPI"`
}

type AmfInterface struct {
	IPv4Addr string `yaml:"ipv4Addr,omitempty"`

	Port int `yaml:"port,omitempty"`
}

type UpfInterface struct {
	IPv4Addr string `yaml:"ipv4Addr,omitempty"`

	Port int `yaml:"port,omitempty"`
}

type NetworkName struct {
	Full string `yaml:"full"`

	Short string `yaml:"short,omitempty"`
}

// Create new AMF context
func RAN_Self() *RANContext {
	return &ranContext
}
