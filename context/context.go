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
	NGRANInterface  NGRANInterface
	GTPInterface    GTPInterface
	UEList          []UE
	UriScheme       models.UriScheme
	NfId            string
	Name            string
	Security        Security
	Snssai          Snssai
	PLMN            PLMN
}

type UE struct {
	Supi string `yaml:"SUPI,omitempty" json:"SUPI"`

	IPv4 string `yaml:"ipv4,omitempty" json:"SUPI"`
}

type AmfInterface struct {
	IPv4Addr string `yaml:"ipv4Addr,omitempty"`

	Port int `yaml:"port,omitempty"`
}

type UpfInterface struct {
	IPv4Addr string `yaml:"ipv4Addr,omitempty"`

	Port int `yaml:"port,omitempty"`
}

type NGRANInterface struct {
	IPv4Addr string `yaml:"ipv4Addr,omitempty"`

	Port int `yaml:"port,omitempty"`
}

type GTPInterface struct {
	IPv4Addr string `yaml:"ipv4Addr,omitempty"`

	Port int `yaml:"port,omitempty"`
}

type NetworkName struct {
	Full string `yaml:"full"`

	Short string `yaml:"short,omitempty"`
}

type Security struct {
	NetworkName string `yaml:"networkName"`

	K string `yaml:"k"`

	OPC string `yaml:"opc"`

	OP string `yaml:"op"`

	SQN string `yaml:"sqn"`
}

type Snssai struct {
	Sst int32 `yaml:"sst,omitempty"`

	Sd string `yaml:"sd,omitempty"`
}

type PLMN struct {
	MNC string `yaml:"mnc,omitempty"`

	MCC string `yaml:"mcc,omitempty"`
}

// Create new AMF context
func RAN_Self() *RANContext {
	return &ranContext
}
