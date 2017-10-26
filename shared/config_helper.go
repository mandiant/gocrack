package shared

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type ServerCfg struct {
	Address       string  `yaml:"address"`
	Certificate   string  `yaml:"ssl_certificate"`
	PrivateKey    string  `yaml:"ssl_private_key"`
	CACertificate *string `yaml:"ssl_ca_certificate,omitempty"`
	UseSSL        bool    `yaml:"ssl_enabled"`
}

// LoadConfigFile reads configPath into cfg
func LoadConfigFile(configPath string, cfg interface{}) (err error) {
	var b []byte

	if b, err = ioutil.ReadFile(configPath); err != nil {
		return err
	}

	return yaml.Unmarshal(b, cfg)
}
