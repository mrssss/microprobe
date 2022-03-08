package sink

import (
	"github.com/mrssss/microprobe/blueprint"
	"github.com/pkg/errors"
	"sync"
)

type Sink interface {
	Process()
	Stop()
}

var globalSinkFactory = make(map[string]func(*blueprint.SinkConfig, chan string, *sync.WaitGroup) (Sink, error))

func RegisterSink(sinkType string, f func(*blueprint.SinkConfig, chan string, *sync.WaitGroup) (Sink, error)) {
	globalSinkFactory[sinkType] = f
}

func NewSink(sc *blueprint.SinkConfig, ch chan string, done *sync.WaitGroup) (Sink, error) {
	f, ok := globalSinkFactory[sc.Type]
	if ok {
		return f(sc, ch, done)
	}
	return nil, errors.Errorf("invalid sink type=%s, no corresponding sink factory for it", sc.Type)
}
