package source

import (
	"github.com/mrssss/microprobe/blueprint"
	"github.com/pkg/errors"
	"sync"
)

type Source interface {
	Process()
	Stop()
}

var globalSinkFactory = make(map[string]func(*blueprint.SourceConfig, chan string, *sync.WaitGroup) (Source, error))

func RegisterSource(sourceType string, f func(*blueprint.SourceConfig, chan string, *sync.WaitGroup) (Source, error)) {
	globalSinkFactory[sourceType] = f
}

func NewSource(sc *blueprint.SourceConfig, ch chan string, done *sync.WaitGroup) (Source, error) {
	f, ok := globalSinkFactory[sc.Type]
	if ok {
		return f(sc, ch, done)
	}
	return nil, errors.Errorf("invalid sink type=%s, no corresponding sink factory for it", sc.Type)
}
