package configapp

import (
	"github.com/nobuenhombre/suikat/pkg/fico"
	"github.com/nobuenhombre/suikat/pkg/ge"
	"gopkg.in/yaml.v3"
)

type Service interface {
	Load(fileName string) error
	Save(fileName string) error
	Get() *Config
}

type Config struct {
	// Add your config sections here
}

func New(fileName string) (Service, error) {
	cfg := &Config{}

	err := cfg.Load(fileName)
	if err != nil {
		return nil, ge.Pin(err)
	}

	return cfg, nil
}

func (c *Config) Load(fileName string) error {
	txtConfigFile := fico.TxtFile(fileName)

	configData, err := txtConfigFile.Read()
	if err != nil {
		return ge.Pin(err)
	}

	err = yaml.Unmarshal([]byte(configData), c)
	if err != nil {
		return ge.Pin(err)
	}

	return nil
}

func (c *Config) Save(fileName string) error {
	txtConfigFile := fico.TxtFile(fileName)

	configData, err := yaml.Marshal(c)
	if err != nil {
		return ge.Pin(err)
	}

	err = txtConfigFile.Write(string(configData))
	if err != nil {
		return ge.Pin(err)
	}

	return nil
}

func (c *Config) Get() *Config {
	return c
}
