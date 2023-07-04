package config

import "github.com/ilyakaznacheev/cleanenv"

// const SaveReleasePath = "/home/yourok/numParser/releases"
const SaveReleasePath = "public/releases"

const ReleasesLimit = 0

var ProxyHost = ""
var UseProxy = false

type ConfigParser struct {
	Host     string `yaml:"host" env:"HOST_RUTOR" env-default:"http://rutor.info"`
	Port     string `yaml:"port" env:"PORT_RUTOR" env-default:"38888"`
	UseProxy string `yaml:"useproxy" env:"USEPROXY_RUTOR" env-defaults:"false"`
	Proxy    string `yaml:"proxy" env:"PROXY_RUTOR" env-default:""`
}

var cfg ConfigParser

func ReadConfigParser(vars string) (string, error) {
	err := cleanenv.ReadConfig("config.yml", &cfg)
	if err == nil {
		switch {
		case vars == "Host":
			return cfg.Host, nil
		case vars == "Port":
			return cfg.Port, nil
		case vars == "Proxy":
			return cfg.Proxy, nil
		case vars == "UseProxy":
			return cfg.UseProxy, nil
		}
	}
	return "", err
}
