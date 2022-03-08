package blueprint

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type BluePrint struct {
	Source SourceConfig `yaml:"source"`
	Sink   SinkConfig   `yaml:"sink"`
	Client ClientConfig `yaml:"client"`
}

func (bp *BluePrint) Validate() error {
	if err := bp.Source.Validate(); err != nil {
		return err
	}
	if err := bp.Sink.Validate(); err != nil {
		return err
	}
	if err := bp.Client.Validate(); err != nil {
		return err
	}
	return nil
}

func newBluePrintFromBytes(data []byte) (*BluePrint, error) {
	bp := &BluePrint{}
	if err := yaml.Unmarshal(data, bp); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal config data")
	}
	return bp, nil
}

func NewBluePrintFromFile(filepath string) (*BluePrint, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read config file")
	}
	return newBluePrintFromBytes(data)
}
