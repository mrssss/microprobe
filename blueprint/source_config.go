package blueprint

import "github.com/pkg/errors"

type SourceConfig struct {
	Name     string      `yaml:"name"`
	Type     string      `yaml:"type"`
	Settings interface{} `yaml:"settings"`
}

func (source *SourceConfig) Validate() error {
	if source.Name == "" || source.Type == "" {
		return errors.New("name and type are required for data source")
	}

	return nil
}
