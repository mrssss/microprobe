package blueprint

import "github.com/pkg/errors"

type SinkConfig struct {
	Type string      `yaml:"type"`
	Ctx  interface{} `yaml:"context"`
}

func (sink *SinkConfig) Validate() error {
	if sink.Type == "" {
		return errors.New("Invalid configuration, server type is not setup correctly")
	}
	return nil
}
