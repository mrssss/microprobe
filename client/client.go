package client

import (
	"github.com/mrssss/microprobe/blueprint"
	"github.com/pkg/errors"
	"sync"
)

type Client interface {
	Process()
	Stop()
}

var globalClientFactory = make(map[string]func(*blueprint.ClientConfig, *sync.WaitGroup) (Client, error))

func RegisterClient(clientType string, f func(*blueprint.ClientConfig, *sync.WaitGroup) (Client, error)) {
	globalClientFactory[clientType] = f
}

func NewClient(sc *blueprint.ClientConfig, done *sync.WaitGroup) (Client, error) {
	f, ok := globalClientFactory[sc.Type]
	if ok {
		return f(sc, done)
	}
	return nil, errors.Errorf("invalid sink type=%s, no corresponding sink factory for it", sc.Type)
}
