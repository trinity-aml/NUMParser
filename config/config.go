package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	"gopkg.in/yaml.v3"
)

// const SaveReleasePath = "/home/yourok/numParser/releases"
const SaveReleasePath = "public/releases"
const ConfigFile = "config.yml"

var ProxyHost = ""
var UseProxy = false

type ConfigParser struct {
	Host      string `yaml:"host" json:"host" env:"HOST_RUTOR" env-default:"http://rutor.info"`
	Port      string `yaml:"port" json:"port" env:"PORT_RUTOR" env-default:"38888"`
	UseProxy  bool   `yaml:"useproxy" json:"useproxy" env:"USEPROXY_RUTOR" env-default:"false"`
	Proxy     string `yaml:"proxy" json:"proxy" env:"PROXY_RUTOR" env-default:""`
	TmdbToken string `yaml:"tmdbtoken" json:"tmdbtoken"`
	AigKey    string `yaml:"aigkey" json:"aigkey"`
}

var (
	cfg ConfigParser
	mu  sync.RWMutex
)

func DefaultConfig() ConfigParser {
	return ConfigParser{
		Host: "http://rutor.info",
		Port: "38888",
	}
}

func LoadConfig() (ConfigParser, error) {
	next := DefaultConfig()
	if _, err := os.Stat(ConfigFile); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return next, err
		}
		if envErr := cleanenv.ReadEnv(&next); envErr != nil {
			return next, envErr
		}
	} else if err := cleanenv.ReadConfig(ConfigFile, &next); err != nil {
		return next, err
	}

	next.Normalize()

	mu.Lock()
	cfg = next
	mu.Unlock()

	return next, nil
}

func SaveConfig(next ConfigParser) (ConfigParser, error) {
	next.Normalize()
	if err := ValidateConfig(next); err != nil {
		return next, err
	}

	buf, err := yaml.Marshal(&next)
	if err != nil {
		return next, err
	}

	if err := os.WriteFile(ConfigFile, buf, 0644); err != nil {
		return next, err
	}

	mu.Lock()
	cfg = next
	ProxyHost = next.Proxy
	UseProxy = next.UseProxy
	mu.Unlock()

	return next, nil
}

func ValidateConfig(next ConfigParser) error {
	next.Normalize()

	if err := validateHTTPURL("host", next.Host); err != nil {
		return err
	}

	port, err := strconv.Atoi(next.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("port must be a number from 1 to 65535")
	}

	if next.Proxy != "" {
		if err := validateProxyURL(next.Proxy); err != nil {
			return err
		}
	}

	return nil
}

func (c *ConfigParser) Normalize() {
	c.Host = strings.TrimSpace(c.Host)
	c.Port = strings.TrimSpace(c.Port)
	c.Proxy = strings.TrimSpace(c.Proxy)
	c.TmdbToken = strings.TrimSpace(c.TmdbToken)
	c.AigKey = strings.TrimSpace(c.AigKey)

	if c.Host == "" {
		c.Host = "http://rutor.info"
	}
	if c.Port == "" {
		c.Port = "38888"
	}
}

func validateHTTPURL(name, raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("%s must be a valid http or https URL", name)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%s must use http or https", name)
	}
	return nil
}

func validateProxyURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("proxy must be a valid URL")
	}
	switch u.Scheme {
	case "http", "https", "socks4", "socks4a", "socks5", "socks5h":
		return nil
	default:
		return fmt.Errorf("proxy scheme must be http, https, socks4, socks4a, socks5 or socks5h")
	}
}

func ReadConfigParser(vars string) (string, error) {
	next, err := LoadConfig()
	if err != nil {
		return "", err
	}

	switch {
	case vars == "Host":
		return next.Host, nil
	case vars == "Port":
		return next.Port, nil
	case vars == "Proxy":
		return next.Proxy, nil
	case vars == "UseProxy":
		return strconv.FormatBool(next.UseProxy), nil
	case vars == "TmdbToken":
		return next.TmdbToken, nil
	case vars == "AigKey":
		return next.AigKey, nil
	}
	return "", nil
}
