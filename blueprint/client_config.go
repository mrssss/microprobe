package blueprint

import "github.com/pkg/errors"

type ClientConfig struct {
	Type   string      `yaml:"type"`
	Result string      `yaml:"result"`
	Ctx    interface{} `yaml:"context"`
}

func (client *ClientConfig) Validate() error {
	if client.Type == "" {
		return errors.New("Invalid configuration, server type is not setup correctly")
	}
	return nil
}
