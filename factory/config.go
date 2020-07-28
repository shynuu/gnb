/*
 * RAN Configuration Factory
 */

package factory

import "free5gc/src/ran/context"

type Config struct {
	Info *Info `yaml:"info"`

	Configuration *Configuration `yaml:"configuration"`
}

type Info struct {
	Version string `yaml:"version,omitempty"`

	Description string `yaml:"description,omitempty"`
}

type Configuration struct {
	RanName string `yaml:"ranName,omitempty"`

	NetworkName context.NetworkName `yaml:"networkName,omitempty"`

	UESubnet string `yaml:"ueSubnet,omitempty"`

	UEList []context.UE `yaml:"ue,omitempty"`

	Sbi *Sbi `yaml:"sbi,omitempty"`

	AmfInterface context.AmfInterface `yaml:"amfInterface,omitempty"`

	UpfInterface context.UpfInterface `yaml:"upfInterface,omitempty"`
}

type Sbi struct {
	Scheme string `yaml:"scheme"`

	IPv4Addr string `yaml:"ipv4Addr,omitempty"`

	Port int `yaml:"port,omitempty"`
}
