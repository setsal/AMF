package config

// CommonConfig :
type CommonConfig struct {
	Bind string `valid:"ipv4"`
	Port string `valid:"port"`
}

// Conf :
var Conf CommonConfig
